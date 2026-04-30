package integration

import (
	"barter-port/pkg/db"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	dockercontainer "github.com/docker/docker/api/types/container"
	dockerfilters "github.com/docker/docker/api/types/filters"
	dockernetwork "github.com/docker/docker/api/types/network"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	postgresAlias = "db"
	kafkaAlias    = "kafka"
	smtpAlias     = "smtp4dev"
	seaweedAlias  = "seaweedfs"

	postgresPort nat.Port = "5432/tcp"
	kafkaPort    nat.Port = "9092/tcp"
	smtpPort     nat.Port = "25/tcp"
	smtpUIPort   nat.Port = "80/tcp"
	seaweedPort  nat.Port = "8333/tcp"

	authHTTPPort  nat.Port = "8080/tcp"
	itemsHTTPPort nat.Port = "8080/tcp"
	usersHTTPPort nat.Port = "8080/tcp"
	chatsHTTPPort nat.Port = "8080/tcp"

	PostgresDBName = "postgres"
	AuthDBName     = "auth_db"
	UsersDBName    = "users_db"

	defaultDBUser     = "postgres"
	defaultDBPassword = "postgres"

	testJWTAccessSecret  = "integration-access-secret"
	testJWTRefreshSecret = "integration-refresh-secret"

	// stableNetworkName — имя Docker-сети, используемой при переиспользовании контейнеров.
	stableNetworkName = "barter-port-integration-test"

	// reuseEnvVar — переменная окружения для включения режима переиспользования контейнеров.
	// Установите BARTER_PORT_REUSE_CONTAINERS=true чтобы контейнеры не пересоздавались между запусками.
	reuseEnvVar = "BARTER_PORT_REUSE_CONTAINERS"

	// reusePostgresHostPortEnvVar — порт на хосте для тестового Postgres в debug/reuse-режиме.
	// По умолчанию используем 15432, чтобы не конфликтовать с локальным Postgres на 5432.
	reusePostgresHostPortEnvVar  = "BARTER_PORT_TEST_POSTGRES_HOST_PORT"
	defaultReusePostgresHostPort = "15432"
)

// shouldReuse возвращает true, если включён режим переиспользования контейнеров.
func shouldReuse() bool {
	v := os.Getenv(reuseEnvVar)
	return v == "true" || v == "1"
}

func reusePostgresHostPort() string {
	if v := os.Getenv(reusePostgresHostPortEnvVar); v != "" {
		return v
	}
	return defaultReusePostgresHostPort
}

// ensureStableNetwork создаёт Docker-сеть с фиксированным именем или возвращает имя
// уже существующей. Используется в режиме reuse.
func ensureStableNetwork(ctx context.Context) (string, error) {
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return "", fmt.Errorf("docker client: %w", err)
	}
	defer cli.Close()

	nets, err := cli.NetworkList(ctx, dockernetwork.ListOptions{
		Filters: dockerfilters.NewArgs(dockerfilters.Arg("name", stableNetworkName)),
	})
	if err != nil {
		return "", fmt.Errorf("network list: %w", err)
	}
	for _, n := range nets {
		if n.Name == stableNetworkName {
			return stableNetworkName, nil
		}
	}

	if _, err := cli.NetworkCreate(ctx, stableNetworkName, dockernetwork.CreateOptions{
		Driver: "bridge",
	}); err != nil {
		return "", fmt.Errorf("network create: %w", err)
	}
	return stableNetworkName, nil
}

type FixtureOptions struct {
	NeedPostgres bool
	NeedKafka    bool
	NeedSMTP     bool
	NeedSeaweed  bool

	NeedAuth  bool
	NeedItems bool
	NeedUsers bool
	NeedChats bool
}

type Fixture struct {
	Ctx     context.Context
	Network *testcontainers.DockerNetwork

	Postgres testcontainers.Container
	Kafka    testcontainers.Container
	SMTP     testcontainers.Container
	Seaweed  testcontainers.Container

	Auth  testcontainers.Container
	Items testcontainers.Container
	Users testcontainers.Container
	Chats testcontainers.Container

	AuthURL  string
	DealsURL string
	UsersURL string
	ChatsURL string
}

// globalFixture — единый стек контейнеров, разделяемый всеми тестами пакета.
var globalFixture *Fixture

// TerminateAll останавливает все контейнеры и удаляет сеть.
// В режиме reuse (BARTER_PORT_REUSE_CONTAINERS=true) контейнеры не останавливаются.
func (f *Fixture) TerminateAll(ctx context.Context) error {
	if shouldReuse() {
		return nil
	}

	var errs []error
	for _, c := range []testcontainers.Container{
		f.Chats, f.Users, f.Items, f.Auth,
		f.Seaweed, f.SMTP, f.Kafka, f.Postgres,
	} {
		if c != nil {
			if err := testcontainers.TerminateContainer(c); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if f.Network != nil {
		if err := f.Network.Remove(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// ────────────────────────────────────────────────────────────────
// Публичный конструктор (для тестов с собственным fixture)
// ────────────────────────────────────────────────────────────────

// NewFixture создаёт fixture с параллельным запуском инфраструктуры.
func NewFixture(t *testing.T, opts FixtureOptions) *Fixture {
	t.Helper()

	ctx := context.Background()
	net := SetupNetwork(ctx, t)

	f := &Fixture{
		Ctx:     ctx,
		Network: net,
	}

	if opts.NeedAuth {
		opts.NeedPostgres = true
		opts.NeedKafka = true
		opts.NeedSMTP = true
	}
	if opts.NeedItems {
		opts.NeedPostgres = true
		opts.NeedSeaweed = true
	}
	if opts.NeedUsers {
		opts.NeedPostgres = true
		opts.NeedKafka = true
	}
	if opts.NeedChats {
		opts.NeedPostgres = true
		opts.NeedAuth = true
		opts.NeedItems = true
		opts.NeedUsers = true
	}

	// Параллельный запуск инфраструктурных контейнеров.
	setupInfraParallel(ctx, net.Name, opts, f, t)

	// Сервисные контейнеры запускаются последовательно:
	// Users зависит от Auth (gRPC), поэтому Auth должен быть готов первым.
	if opts.NeedAuth {
		req := buildServiceRequest(net.Name, "auth", string(authHTTPPort))
		req.Env = serviceEnv()
		req.Env["CONFIG_SERVICE"] = "/app/config/auth.yaml"
		req.Env["JWT_REFRESH_SECRET"] = testJWTRefreshSecret
		req.Env["MAILER_BYPASS"] = "false" // TODO: в проде надо false
		f.Auth = startContainer(ctx, t, req)
		f.AuthURL = containerBaseURL(ctx, t, f.Auth, authHTTPPort)
	}
	if opts.NeedItems {
		req := buildServiceRequest(net.Name, "deals", string(itemsHTTPPort))
		req.Env = serviceEnv()
		req.Env["CONFIG_SERVICE"] = "/app/config/deals.yaml"
		f.Items = startContainer(ctx, t, req)
		f.DealsURL = containerBaseURL(ctx, t, f.Items, itemsHTTPPort)
	}
	if opts.NeedUsers {
		req := buildServiceRequest(net.Name, "users", string(usersHTTPPort))
		req.Env = serviceEnv()
		req.Env["CONFIG_SERVICE"] = "/app/config/users.yaml"
		req.Env["AUTH_GRPC_ADDR"] = "auth:50051"
		f.Users = startContainer(ctx, t, req)
		f.UsersURL = containerBaseURL(ctx, t, f.Users, usersHTTPPort)
	}
	if opts.NeedChats {
		req := buildServiceRequest(net.Name, "chats", string(chatsHTTPPort))
		req.Env = serviceEnv()
		req.Env["CONFIG_SERVICE"] = "/app/config/chats.yaml"
		f.Chats = startContainer(ctx, t, req)
		f.ChatsURL = containerBaseURL(ctx, t, f.Chats, chatsHTTPPort)
	}

	return f
}

// ────────────────────────────────────────────────────────────────
// Конструктор для TestMain (без *testing.T)
// ────────────────────────────────────────────────────────────────

// newSharedFixture создаёт Fixture без *testing.T — для использования в TestMain.
// При частичном сбое уже запущенные контейнеры сохраняются в *Fixture,
// чтобы вызывающий код мог вызвать TerminateAll для очистки.
func newSharedFixture(ctx context.Context, opts FixtureOptions) (*Fixture, error) {
	if opts.NeedAuth {
		opts.NeedPostgres = true
		opts.NeedKafka = true
		opts.NeedSMTP = true
	}
	if opts.NeedItems {
		opts.NeedPostgres = true
		opts.NeedSeaweed = true
	}
	if opts.NeedUsers {
		opts.NeedPostgres = true
		opts.NeedKafka = true
	}
	if opts.NeedChats {
		opts.NeedPostgres = true
		opts.NeedAuth = true
		opts.NeedItems = true
		opts.NeedUsers = true
	}

	// Определяем сеть: в режиме reuse используем фиксированное имя,
	// иначе создаём новую случайную сеть.
	var netName string
	f := &Fixture{Ctx: ctx}

	if shouldReuse() {
		var err error
		netName, err = ensureStableNetwork(ctx)
		if err != nil {
			return f, fmt.Errorf("stable network: %w", err)
		}
		// f.Network остаётся nil — сеть не удаляется при TerminateAll
	} else {
		net, err := network.New(ctx)
		if err != nil {
			return f, fmt.Errorf("create network: %w", err)
		}
		f.Network = net
		netName = net.Name
	}

	// Параллельный запуск инфраструктуры.
	if err := setupInfraParallelShared(ctx, netName, opts, f); err != nil {
		return f, err
	}

	// Сервисные контейнеры — последовательно.
	if opts.NeedAuth {
		c, err := launchAuth(ctx, netName)
		f.Auth = c
		if err != nil {
			return f, fmt.Errorf("launch auth: %w", err)
		}
		url, err := getContainerBaseURL(ctx, c, authHTTPPort)
		if err != nil {
			return f, fmt.Errorf("auth base url: %w", err)
		}
		f.AuthURL = url
	}
	if opts.NeedItems {
		c, err := launchDeals(ctx, netName)
		f.Items = c
		if err != nil {
			return f, fmt.Errorf("launch items: %w", err)
		}
		url, err := getContainerBaseURL(ctx, c, itemsHTTPPort)
		if err != nil {
			return f, fmt.Errorf("items base url: %w", err)
		}
		f.DealsURL = url
	}
	if opts.NeedUsers {
		c, err := launchUsers(ctx, netName)
		f.Users = c
		if err != nil {
			return f, fmt.Errorf("launch users: %w", err)
		}
		url, err := getContainerBaseURL(ctx, c, usersHTTPPort)
		if err != nil {
			return f, fmt.Errorf("users base url: %w", err)
		}
		f.UsersURL = url
	}
	if opts.NeedChats {
		c, err := launchChats(ctx, netName)
		f.Chats = c
		if err != nil {
			return f, fmt.Errorf("launch chats: %w", err)
		}
		url, err := getContainerBaseURL(ctx, c, chatsHTTPPort)
		if err != nil {
			return f, fmt.Errorf("chats base url: %w", err)
		}
		f.ChatsURL = url
	}

	return f, nil
}

// ────────────────────────────────────────────────────────────────
// Параллельный запуск инфраструктуры
// ────────────────────────────────────────────────────────────────

type infraResult struct {
	name      string
	container testcontainers.Container
	err       error
}

// setupInfraParallel запускает инфраструктурные контейнеры параллельно.
func setupInfraParallel(
	ctx context.Context,
	netName string,
	opts FixtureOptions,
	f *Fixture,
	t *testing.T,
) {
	t.Helper()

	ch := make(chan infraResult, 4)
	launched := 0

	if opts.NeedPostgres {
		launched++
		go func() {
			c, err := launchPostgres(ctx, netName)
			ch <- infraResult{"postgres", c, err}
		}()
	}
	if opts.NeedKafka {
		launched++
		go func() {
			c, err := launchKafka(ctx, netName)
			ch <- infraResult{"kafka", c, err}
		}()
	}
	if opts.NeedSMTP {
		launched++
		go func() {
			c, err := launchSMTP(ctx, netName)
			ch <- infraResult{"smtp", c, err}
		}()
	}
	if opts.NeedSeaweed {
		launched++
		go func() {
			c, err := launchSeaweed(ctx, netName)
			ch <- infraResult{"seaweed", c, err}
		}()
	}

	for i := 0; i < launched; i++ {
		result := <-ch

		// Регистрируем очистку ДО проверки ошибки.
		if result.container != nil {
			c := result.container
			t.Cleanup(func() {
				require.NoError(t, testcontainers.TerminateContainer(c))
			})
		}

		require.NoError(t, result.err, "инфраструктурный контейнер %q не запустился", result.name)

		switch result.name {
		case "postgres":
			f.Postgres = result.container
		case "kafka":
			f.Kafka = result.container
		case "smtp":
			f.SMTP = result.container
		case "seaweed":
			f.Seaweed = result.container
		}
	}
}

// setupInfraParallelShared — вариант для TestMain: без *testing.T, возвращает первую ошибку.
func setupInfraParallelShared(
	ctx context.Context,
	netName string,
	opts FixtureOptions,
	f *Fixture,
) error {
	ch := make(chan infraResult, 4)
	launched := 0

	if opts.NeedPostgres {
		launched++
		go func() {
			c, err := launchPostgres(ctx, netName)
			ch <- infraResult{"postgres", c, err}
		}()
	}
	if opts.NeedKafka {
		launched++
		go func() {
			c, err := launchKafka(ctx, netName)
			ch <- infraResult{"kafka", c, err}
		}()
	}
	if opts.NeedSMTP {
		launched++
		go func() {
			c, err := launchSMTP(ctx, netName)
			ch <- infraResult{"smtp", c, err}
		}()
	}
	if opts.NeedSeaweed {
		launched++
		go func() {
			c, err := launchSeaweed(ctx, netName)
			ch <- infraResult{"seaweed", c, err}
		}()
	}

	var firstErr error
	for i := 0; i < launched; i++ {
		result := <-ch

		// Сохраняем контейнер сразу — TerminateAll должен его найти при ошибке.
		switch result.name {
		case "postgres":
			f.Postgres = result.container
		case "kafka":
			f.Kafka = result.container
		case "smtp":
			f.SMTP = result.container
		case "seaweed":
			f.Seaweed = result.container
		}

		if result.err != nil && firstErr == nil {
			firstErr = fmt.Errorf("запуск %s: %w", result.name, result.err)
		}
	}

	return firstErr
}

// ────────────────────────────────────────────────────────────────
// Низкоуровневые функции запуска (без *testing.T)
// ────────────────────────────────────────────────────────────────

func launchPostgres(ctx context.Context, netName string) (testcontainers.Container, error) {
	projectRoot := mustGetProjectRoot()

	name := ""
	if shouldReuse() {
		name = "barter-port-integration-postgres"
	}

	req := testcontainers.ContainerRequest{
		Name:         name,
		Image:        "postgres:16",
		ExposedPorts: []string{string(postgresPort)},
		Networks:     []string{netName},
		NetworkAliases: map[string][]string{
			netName: {postgresAlias},
		},
		Env: map[string]string{
			"POSTGRES_USER":     defaultDBUser,
			"POSTGRES_PASSWORD": defaultDBPassword,
			"POSTGRES_DB":       PostgresDBName,
		},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(projectRoot, "init.sql"),
				ContainerFilePath: "/docker-entrypoint-initdb.d/init.sql",
				FileMode:          0o644,
			},
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(2 * time.Minute),
	}
	if shouldReuse() {
		hostPort := reusePostgresHostPort()
		req.HostConfigModifier = func(hc *dockercontainer.HostConfig) {
			if hc.PortBindings == nil {
				hc.PortBindings = nat.PortMap{}
			}
			hc.PortBindings[postgresPort] = []nat.PortBinding{{
				HostIP:   "127.0.0.1",
				HostPort: hostPort,
			}}
		}
	}
	return launchContainer(ctx, req)
}

func launchKafka(ctx context.Context, netName string) (testcontainers.Container, error) {
	name := ""
	if shouldReuse() {
		name = "barter-port-integration-kafka"
	}

	req := testcontainers.ContainerRequest{
		Name:         name,
		Image:        "apache/kafka:4.2.0",
		ExposedPorts: []string{string(kafkaPort)},
		Networks:     []string{netName},
		NetworkAliases: map[string][]string{
			netName: {kafkaAlias},
		},
		Env: map[string]string{
			"CLUSTER_ID":                                     "MkU3OEVBNTcwNTJENDM2Qk",
			"KAFKA_NODE_ID":                                  "1",
			"KAFKA_PROCESS_ROLES":                            "broker,controller",
			"KAFKA_CONTROLLER_QUORUM_VOTERS":                 "1@kafka:9093",
			"KAFKA_LISTENERS":                                "PLAINTEXT://:9092,CONTROLLER://:9093",
			"KAFKA_ADVERTISED_LISTENERS":                     "PLAINTEXT://kafka:9092",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":           "PLAINTEXT:PLAINTEXT,CONTROLLER:PLAINTEXT",
			"KAFKA_CONTROLLER_LISTENER_NAMES":                "CONTROLLER",
			"KAFKA_INTER_BROKER_LISTENER_NAME":               "PLAINTEXT",
			"KAFKA_AUTO_CREATE_TOPICS_ENABLE":                "true",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":         "1",
			"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR": "1",
			"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":            "1",
			"KAFKA_NUM_PARTITIONS":                           "1",
			"KAFKA_LOG_DIRS":                                 "/var/lib/kafka/data",
		},
		WaitingFor: wait.ForListeningPort(kafkaPort).WithStartupTimeout(2 * time.Minute),
	}
	return launchContainer(ctx, req)
}

func launchSMTP(ctx context.Context, netName string) (testcontainers.Container, error) {
	name := ""
	if shouldReuse() {
		name = "barter-port-integration-smtp"
	}

	req := testcontainers.ContainerRequest{
		Name:         name,
		Image:        "rnwood/smtp4dev:latest",
		ExposedPorts: []string{string(smtpPort), string(smtpUIPort)},
		Networks:     []string{netName},
		NetworkAliases: map[string][]string{
			netName: {smtpAlias},
		},
		Env: map[string]string{
			"ServerOptions__HostName":               smtpAlias,
			"ServerOptions__AuthenticationRequired": "true",
			"ServerOptions__Users__0__Username":     "user",
			"ServerOptions__Users__0__Password":     "password",
			"ServerOptions__TlsMode":                "StartTls",
			"ServerOptions__TlsCertificate":         "",
		},
		WaitingFor: wait.ForListeningPort(smtpPort).WithStartupTimeout(2 * time.Minute),
	}
	return launchContainer(ctx, req)
}

func launchSeaweed(ctx context.Context, netName string) (testcontainers.Container, error) {
	projectRoot := mustGetProjectRoot()

	name := ""
	if shouldReuse() {
		name = "barter-port-integration-seaweed"
	}

	req := testcontainers.ContainerRequest{
		Name:         name,
		Image:        "chrislusf/seaweedfs:3.94",
		ExposedPorts: []string{string(seaweedPort)},
		Networks:     []string{netName},
		NetworkAliases: map[string][]string{
			netName: {seaweedAlias},
		},
		Env: map[string]string{
			"S3_ACCESS_KEY_ID":     "barter-port-s3",
			"S3_SECRET_ACCESS_KEY": "barter-port-s3-secret",
			"S3_REGION":            "us-east-1",
		},
		Entrypoint: []string{"/bin/sh"},
		Cmd: []string{
			"-c",
			`sed -e "s|__S3_ACCESS_KEY_ID__|${S3_ACCESS_KEY_ID}|g" -e "s|__S3_SECRET_ACCESS_KEY__|${S3_SECRET_ACCESS_KEY}|g" /etc/seaweedfs/s3-config.json > /tmp/s3-config.json && exec weed server -dir=/data -volume.max=100 -master.volumeSizeLimitMB=128 -filer -filer.allowedOrigins="*" -s3 -s3.config=/tmp/s3-config.json`,
		},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(projectRoot, "config", "seaweedfs-s3.json"),
				ContainerFilePath: "/etc/seaweedfs/s3-config.json",
				FileMode:          0o644,
			},
		},
		WaitingFor: wait.ForListeningPort(seaweedPort).WithStartupTimeout(2 * time.Minute),
	}
	return launchContainer(ctx, req)
}

func launchAuth(ctx context.Context, netName string) (testcontainers.Container, error) {
	req := buildServiceRequest(netName, "auth", string(authHTTPPort))
	req.Env = serviceEnv()
	req.Env["CONFIG_SERVICE"] = "/app/config/auth.yaml"
	req.Env["JWT_REFRESH_SECRET"] = testJWTRefreshSecret
	req.Env["MAILER_BYPASS"] = "true"
	return launchContainer(ctx, req)
}

func launchDeals(ctx context.Context, netName string) (testcontainers.Container, error) {
	req := buildServiceRequest(netName, "deals", string(itemsHTTPPort))
	req.Env = serviceEnv()
	req.Env["CONFIG_SERVICE"] = "/app/config/deals.yaml"
	return launchContainer(ctx, req)
}

func launchUsers(ctx context.Context, netName string) (testcontainers.Container, error) {
	req := buildServiceRequest(netName, "users", string(usersHTTPPort))
	req.Env = serviceEnv()
	req.Env["CONFIG_SERVICE"] = "/app/config/users.yaml"
	req.Env["AUTH_GRPC_ADDR"] = "auth:50051"
	return launchContainer(ctx, req)
}

func launchChats(ctx context.Context, netName string) (testcontainers.Container, error) {
	req := buildServiceRequest(netName, "chats", string(chatsHTTPPort))
	req.Env = serviceEnv()
	req.Env["CONFIG_SERVICE"] = "/app/config/chats.yaml"
	return launchContainer(ctx, req)
}

func launchContainer(ctx context.Context, req testcontainers.ContainerRequest) (testcontainers.Container, error) {
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            shouldReuse(),
	})
}

// buildServiceRequest строит ContainerRequest для сервисного контейнера.
//
// Образ собирается с KeepImage: true — это сохраняет Docker build cache между запусками,
// позволяя избежать полной пересборки если код не изменился.
func buildServiceRequest(netName string, service string, exposedPorts ...string) testcontainers.ContainerRequest {
	projectRoot := mustGetProjectRoot()
	alias := service
	serviceCopy := service // копия для указателя в BuildArgs

	name := ""
	if shouldReuse() {
		name = "barter-port-integration-" + service
	}

	req := testcontainers.ContainerRequest{
		Name: name,
		FromDockerfile: testcontainers.FromDockerfile{
			Context:       projectRoot,
			Dockerfile:    "Dockerfile",
			PrintBuildLog: false,
			KeepImage:     true,
			Repo:          "barter-port-integration",
			Tag:           service,
			BuildArgs: map[string]*string{
				"SERVICE": &serviceCopy,
			},
		},
		ExposedPorts: exposedPorts,
		Networks:     []string{netName},
		NetworkAliases: map[string][]string{
			netName: {alias},
		},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      filepath.Join(projectRoot, "config", "common.yaml"),
				ContainerFilePath: "/app/config/common.yaml",
				FileMode:          0o644,
			},
			{
				HostFilePath:      filepath.Join(projectRoot, "config", "docker.yaml"),
				ContainerFilePath: "/app/config/docker.yaml",
				FileMode:          0o644,
			},
			{
				HostFilePath:      filepath.Join(projectRoot, "config", "integration.yaml"),
				ContainerFilePath: "/app/config/integration.yaml",
				FileMode:          0o644,
			},
			{
				HostFilePath:      filepath.Join(projectRoot, "config", "auth.yaml"),
				ContainerFilePath: "/app/config/auth.yaml",
				FileMode:          0o644,
			},
			{
				HostFilePath:      filepath.Join(projectRoot, "config", "deals.yaml"),
				ContainerFilePath: "/app/config/deals.yaml",
				FileMode:          0o644,
			},
			{
				HostFilePath:      filepath.Join(projectRoot, "config", "users.yaml"),
				ContainerFilePath: "/app/config/users.yaml",
				FileMode:          0o644,
			},
			{
				HostFilePath:      filepath.Join(projectRoot, "config", "chats.yaml"),
				ContainerFilePath: "/app/config/chats.yaml",
				FileMode:          0o644,
			},
		},
	}

	if len(exposedPorts) > 0 {
		req.WaitingFor = wait.ForListeningPort(nat.Port(exposedPorts[0])).
			WithStartupTimeout(2 * time.Minute)
	}

	return req
}

// ────────────────────────────────────────────────────────────────
// Публичные Setup-обёртки (обратная совместимость)
// ────────────────────────────────────────────────────────────────

func SetupNetwork(ctx context.Context, t *testing.T) *testcontainers.DockerNetwork {
	t.Helper()

	net, err := network.New(ctx)
	require.NoError(t, err)
	testcontainers.CleanupNetwork(t, net)

	return net
}

func SetupPostgres(ctx context.Context, net *testcontainers.DockerNetwork, t *testing.T) testcontainers.Container {
	t.Helper()

	c, err := launchPostgres(ctx, net.Name)
	if c != nil {
		t.Cleanup(func() { require.NoError(t, testcontainers.TerminateContainer(c)) })
	}
	require.NoError(t, err)
	return c
}

func SetupKafka(ctx context.Context, net *testcontainers.DockerNetwork, t *testing.T) testcontainers.Container {
	t.Helper()

	c, err := launchKafka(ctx, net.Name)
	if c != nil {
		t.Cleanup(func() { require.NoError(t, testcontainers.TerminateContainer(c)) })
	}
	require.NoError(t, err)
	return c
}

func SetupSMTP(ctx context.Context, net *testcontainers.DockerNetwork, t *testing.T) testcontainers.Container {
	t.Helper()

	c, err := launchSMTP(ctx, net.Name)
	if c != nil {
		t.Cleanup(func() { require.NoError(t, testcontainers.TerminateContainer(c)) })
	}
	require.NoError(t, err)
	return c
}

func SetupAuth(ctx context.Context, net *testcontainers.DockerNetwork, t *testing.T) testcontainers.Container {
	t.Helper()
	req := buildServiceRequest(net.Name, "auth", string(authHTTPPort))
	req.Env = serviceEnv()
	req.Env["CONFIG_SERVICE"] = "/app/config/auth.yaml"
	req.Env["JWT_REFRESH_SECRET"] = testJWTRefreshSecret
	req.Env["MAILER_BYPASS"] = "true"
	return startContainer(ctx, t, req)
}

func SetupDeals(ctx context.Context, net *testcontainers.DockerNetwork, t *testing.T) testcontainers.Container {
	t.Helper()
	req := buildServiceRequest(net.Name, "deals", string(itemsHTTPPort))
	req.Env = serviceEnv()
	req.Env["CONFIG_SERVICE"] = "/app/config/deals.yaml"
	return startContainer(ctx, t, req)
}

func SetupUsers(ctx context.Context, net *testcontainers.DockerNetwork, t *testing.T) testcontainers.Container {
	t.Helper()
	req := buildServiceRequest(net.Name, "users", string(usersHTTPPort))
	req.Env = serviceEnv()
	req.Env["CONFIG_SERVICE"] = "/app/config/users.yaml"
	req.Env["AUTH_GRPC_ADDR"] = "auth:50051"
	return startContainer(ctx, t, req)
}

func SetupChats(ctx context.Context, net *testcontainers.DockerNetwork, t *testing.T) testcontainers.Container {
	t.Helper()
	req := buildServiceRequest(net.Name, "chats", string(chatsHTTPPort))
	req.Env = serviceEnv()
	req.Env["CONFIG_SERVICE"] = "/app/config/chats.yaml"
	return startContainer(ctx, t, req)
}

func serviceEnv() map[string]string {
	return map[string]string{
		"APP_ENV":           "integration",
		"CONFIG_COMMON":     "/app/config/common.yaml",
		"DB_PASSWORD":       defaultDBPassword,
		"JWT_ACCESS_SECRET": testJWTAccessSecret,
	}
}

// DumpLogsOnFailure регистрирует вывод логов контейнера при падении теста.
func DumpLogsOnFailure(t *testing.T, c testcontainers.Container, name string) {
	t.Helper()
	if c == nil {
		return
	}
	t.Cleanup(func() {
		if !t.Failed() {
			return
		}
		rc, err := c.Logs(context.Background())
		if err != nil {
			t.Logf("=== не удалось получить логи [%s]: %v ===", name, err)
			return
		}
		defer rc.Close()
		if raw, err := io.ReadAll(rc); err == nil {
			t.Logf("=== container logs [%s] ===\n%s", name, string(raw))
		}
	})
}

// startContainer запускает контейнер и регистрирует вывод логов при падении теста.
func startContainer(ctx context.Context, t *testing.T, req testcontainers.ContainerRequest) testcontainers.Container {
	t.Helper()

	container, err := launchContainer(ctx, req)

	if container != nil {
		t.Cleanup(func() {
			if t.Failed() {
				rc, logsErr := container.Logs(ctx)
				if logsErr == nil {
					defer rc.Close()
					if raw, readErr := io.ReadAll(rc); readErr == nil {
						service := ""
						if s := req.FromDockerfile.BuildArgs["SERVICE"]; s != nil {
							service = *s
						}
						t.Logf("=== container logs [%s] ===\n%s", service, string(raw))
					}
				}
			}
			require.NoError(t, testcontainers.TerminateContainer(container))
		})
	}

	require.NoError(t, err)

	return container
}

// ────────────────────────────────────────────────────────────────
// URL-хелперы
// ────────────────────────────────────────────────────────────────

// getContainerBaseURL возвращает HTTP base URL без *testing.T.
func getContainerBaseURL(ctx context.Context, c testcontainers.Container, port nat.Port) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("container host: %w", err)
	}
	mappedPort, err := c.MappedPort(ctx, port)
	if err != nil {
		return "", fmt.Errorf("container mapped port: %w", err)
	}
	return fmt.Sprintf("http://%s:%s", host, mappedPort.Port()), nil
}

// containerBaseURL — обёртка над getContainerBaseURL с t-ассертами.
func containerBaseURL(ctx context.Context, t *testing.T, c testcontainers.Container, port nat.Port) string {
	t.Helper()
	url, err := getContainerBaseURL(ctx, c, port)
	require.NoError(t, err)
	return url
}

// ────────────────────────────────────────────────────────────────
// Вспомогательные функции
// ────────────────────────────────────────────────────────────────

// mustGetProjectRoot возвращает корень проекта без *testing.T.
func mustGetProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("os.Getwd: %v", err))
	}
	return filepath.Clean(filepath.Join(wd, ".."))
}

// projectRoot — обёртка с t для обратной совместимости.
func projectRoot(t *testing.T) string {
	t.Helper()
	return mustGetProjectRoot()
}

// ────────────────────────────────────────────────────────────────
// Работа с базой данных
// ────────────────────────────────────────────────────────────────

func OpenDatabase(t *testing.T, f *Fixture, dbName string) *pgxpool.Pool {
	t.Helper()
	require.NotNil(t, f)
	require.NotNil(t, f.Postgres)

	host, err := f.Postgres.Host(f.Ctx)
	require.NoError(t, err)

	mappedPort, err := f.Postgres.MappedPort(f.Ctx, postgresPort)
	require.NoError(t, err)

	pool, err := pgxpool.New(f.Ctx, db.BuildDSN(db.Config{
		DBUser:     defaultDBUser,
		DBPassword: defaultDBPassword,
		DBHost:     host,
		DBPort:     mappedPort.Port(),
		DBName:     dbName,
	}, true))
	require.NoError(t, err)

	t.Cleanup(pool.Close)
	return pool
}

func RequireTablesExist(t *testing.T, pool *pgxpool.Pool, tables ...string) {
	t.Helper()

	for _, table := range tables {
		var exists bool
		err := pool.QueryRow(
			context.Background(),
			"SELECT to_regclass($1) IS NOT NULL",
			"public."+table,
		).Scan(&exists)
		require.NoError(t, err)
		require.True(t, exists, "expected table %s to exist", table)
	}
}

func RequireAuthTablesExist(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	RequireTablesExist(
		t,
		pool,
		"email_tokens",
		"password_reset_tokens",
		"users",
		"user_creation_outbox",
		"user_creation_events",
		"user_creation_result_inbox",
		"refresh_tokens",
	)
}

func RequireUsersTablesExist(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	RequireTablesExist(
		t,
		pool,
		"users",
		"user_creation_inbox",
		"user_creation_result_outbox",
		"deleted_users",
	)
}
