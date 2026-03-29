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

	postgresPort nat.Port = "5432/tcp"
	kafkaPort    nat.Port = "9092/tcp"
	smtpPort     nat.Port = "25/tcp"
	smtpUIPort   nat.Port = "80/tcp"

	authHTTPPort  nat.Port = "8080/tcp"
	itemsHTTPPort nat.Port = "8080/tcp"
	usersHTTPPort nat.Port = "8080/tcp"

	PostgresDBName = "postgres"
	AuthDBName     = "auth_db"
	UsersDBName    = "users_db"

	defaultDBUser     = "postgres"
	defaultDBPassword = "postgres"

	testJWTAccessSecret  = "integration-access-secret"
	testJWTRefreshSecret = "integration-refresh-secret"
)

type FixtureOptions struct {
	NeedPostgres bool
	NeedKafka    bool
	NeedSMTP     bool

	NeedAuth  bool
	NeedItems bool
	NeedUsers bool
}

type Fixture struct {
	Ctx     context.Context
	Network *testcontainers.DockerNetwork

	Postgres testcontainers.Container
	Kafka    testcontainers.Container
	SMTP     testcontainers.Container

	Auth  testcontainers.Container
	Items testcontainers.Container
	Users testcontainers.Container

	AuthURL  string
	ItemsURL string
	UsersURL string
}

// TerminateAll останавливает все контейнеры и удаляет сеть.
// Используется в TestMain, где нет *testing.T.
func (f *Fixture) TerminateAll(ctx context.Context) error {
	var errs []error
	for _, c := range []testcontainers.Container{
		f.Users, f.Items, f.Auth,
		f.SMTP, f.Kafka, f.Postgres,
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
	}
	if opts.NeedUsers {
		opts.NeedPostgres = true
		opts.NeedKafka = true
	}

	// Параллельный запуск инфраструктурных контейнеров.
	setupInfraParallel(ctx, net, opts, f, t)

	// Сервисные контейнеры запускаются последовательно:
	// Users зависит от Auth (gRPC), поэтому Auth должен быть готов первым.
	if opts.NeedAuth {
		req := buildServiceRequest(net, "auth", string(authHTTPPort))
		req.Env = serviceEnv()
		req.Env["CONFIG_SERVICE"] = "/app/config/auth.yaml"
		req.Env["JWT_REFRESH_SECRET"] = testJWTRefreshSecret
		req.Env["MAILER_BYPASS"] = "true"
		f.Auth = startContainer(ctx, t, req)
		f.AuthURL = containerBaseURL(ctx, t, f.Auth, authHTTPPort)
	}
	if opts.NeedItems {
		req := buildServiceRequest(net, "items", string(itemsHTTPPort))
		req.Env = serviceEnv()
		req.Env["CONFIG_SERVICE"] = "/app/config/items.yaml"
		f.Items = startContainer(ctx, t, req)
		f.ItemsURL = containerBaseURL(ctx, t, f.Items, itemsHTTPPort)
	}
	if opts.NeedUsers {
		req := buildServiceRequest(net, "users", string(usersHTTPPort))
		req.Env = serviceEnv()
		req.Env["CONFIG_SERVICE"] = "/app/config/users.yaml"
		req.Env["AUTH_GRPC_ADDR"] = "auth:50051"
		f.Users = startContainer(ctx, t, req)
		f.UsersURL = containerBaseURL(ctx, t, f.Users, usersHTTPPort)
	}

	return f
}

// ────────────────────────────────────────────────────────────────
// Конструктор для TestMain (без *testing.T)
// ────────────────────────────────────────────────────────────────

// newSharedFixture создаёт Fixture без *testing.T — для использования в TestMain.
// При частичном сбое уже запущенные контейнеры сохраняются в *Fixture,
// чтобы вызывающий код мог вызвать TerminateAll для очистки.
func newSharedFixture(
	ctx context.Context,
	net *testcontainers.DockerNetwork,
	opts FixtureOptions,
) (*Fixture, error) {
	if opts.NeedAuth {
		opts.NeedPostgres = true
		opts.NeedKafka = true
		opts.NeedSMTP = true
	}
	if opts.NeedItems {
		opts.NeedPostgres = true
	}
	if opts.NeedUsers {
		opts.NeedPostgres = true
		opts.NeedKafka = true
	}

	f := &Fixture{Ctx: ctx, Network: net}

	// Параллельный запуск инфраструктуры.
	if err := setupInfraParallelShared(ctx, net, opts, f); err != nil {
		return f, err
	}

	// Сервисные контейнеры — последовательно.
	if opts.NeedAuth {
		c, err := launchAuth(ctx, net)
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
		c, err := launchItems(ctx, net)
		f.Items = c
		if err != nil {
			return f, fmt.Errorf("launch items: %w", err)
		}
		url, err := getContainerBaseURL(ctx, c, itemsHTTPPort)
		if err != nil {
			return f, fmt.Errorf("items base url: %w", err)
		}
		f.ItemsURL = url
	}
	if opts.NeedUsers {
		c, err := launchUsers(ctx, net)
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

// setupInfraParallel запускает Postgres/Kafka/SMTP параллельно.
//
// Буфер канала равен максимальному числу горутин (3): это гарантирует отсутствие
// утечек горутин — даже если t.FailNow() прервёт цикл после первой ошибки,
// оставшиеся горутины смогут отправить результат и завершиться.
//
// Регистрация cleanup выполняется ДО вызова require.NoError, чтобы очистка
// произошла даже при неудачном старте контейнера.
func setupInfraParallel(
	ctx context.Context,
	net *testcontainers.DockerNetwork,
	opts FixtureOptions,
	f *Fixture,
	t *testing.T,
) {
	t.Helper()

	ch := make(chan infraResult, 3)
	launched := 0

	if opts.NeedPostgres {
		launched++
		go func() {
			c, err := launchPostgres(ctx, net)
			ch <- infraResult{"postgres", c, err}
		}()
	}
	if opts.NeedKafka {
		launched++
		go func() {
			c, err := launchKafka(ctx, net)
			ch <- infraResult{"kafka", c, err}
		}()
	}
	if opts.NeedSMTP {
		launched++
		go func() {
			c, err := launchSMTP(ctx, net)
			ch <- infraResult{"smtp", c, err}
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
		}
	}
}

// setupInfraParallelShared — вариант для TestMain: без *testing.T, возвращает первую ошибку.
// Все частично запущенные контейнеры записываются в f, чтобы TerminateAll смог их очистить.
func setupInfraParallelShared(
	ctx context.Context,
	net *testcontainers.DockerNetwork,
	opts FixtureOptions,
	f *Fixture,
) error {
	ch := make(chan infraResult, 3)
	launched := 0

	if opts.NeedPostgres {
		launched++
		go func() {
			c, err := launchPostgres(ctx, net)
			ch <- infraResult{"postgres", c, err}
		}()
	}
	if opts.NeedKafka {
		launched++
		go func() {
			c, err := launchKafka(ctx, net)
			ch <- infraResult{"kafka", c, err}
		}()
	}
	if opts.NeedSMTP {
		launched++
		go func() {
			c, err := launchSMTP(ctx, net)
			ch <- infraResult{"smtp", c, err}
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

func launchPostgres(ctx context.Context, net *testcontainers.DockerNetwork) (testcontainers.Container, error) {
	projectRoot := mustGetProjectRoot()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16",
		ExposedPorts: []string{string(postgresPort)},
		Networks:     []string{net.Name},
		NetworkAliases: map[string][]string{
			net.Name: {postgresAlias},
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
	return launchContainer(ctx, req)
}

func launchKafka(ctx context.Context, net *testcontainers.DockerNetwork) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "apache/kafka:4.2.0",
		ExposedPorts: []string{string(kafkaPort)},
		Networks:     []string{net.Name},
		NetworkAliases: map[string][]string{
			net.Name: {kafkaAlias},
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

func launchSMTP(ctx context.Context, net *testcontainers.DockerNetwork) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "rnwood/smtp4dev:latest",
		ExposedPorts: []string{string(smtpPort), string(smtpUIPort)},
		Networks:     []string{net.Name},
		NetworkAliases: map[string][]string{
			net.Name: {smtpAlias},
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

func launchAuth(ctx context.Context, net *testcontainers.DockerNetwork) (testcontainers.Container, error) {
	req := buildServiceRequest(net, "auth", string(authHTTPPort))
	req.Env = serviceEnv()
	req.Env["CONFIG_SERVICE"] = "/app/config/auth.yaml"
	req.Env["JWT_REFRESH_SECRET"] = testJWTRefreshSecret
	req.Env["MAILER_BYPASS"] = "true"
	return launchContainer(ctx, req)
}

func launchItems(ctx context.Context, net *testcontainers.DockerNetwork) (testcontainers.Container, error) {
	req := buildServiceRequest(net, "items", string(itemsHTTPPort))
	req.Env = serviceEnv()
	req.Env["CONFIG_SERVICE"] = "/app/config/items.yaml"
	return launchContainer(ctx, req)
}

func launchUsers(ctx context.Context, net *testcontainers.DockerNetwork) (testcontainers.Container, error) {
	req := buildServiceRequest(net, "users", string(usersHTTPPort))
	req.Env = serviceEnv()
	req.Env["CONFIG_SERVICE"] = "/app/config/users.yaml"
	req.Env["AUTH_GRPC_ADDR"] = "auth:50051"
	return launchContainer(ctx, req)
}

func launchContainer(ctx context.Context, req testcontainers.ContainerRequest) (testcontainers.Container, error) {
	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

// buildServiceRequest строит ContainerRequest для сервисного контейнера без *testing.T.
func buildServiceRequest(net *testcontainers.DockerNetwork, service string, exposedPorts ...string) testcontainers.ContainerRequest {
	projectRoot := mustGetProjectRoot()
	alias := service

	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:       projectRoot,
			Dockerfile:    "Dockerfile",
			PrintBuildLog: false,
			BuildArgs: map[string]*string{
				"SERVICE": stringPtr(service),
			},
		},
		ExposedPorts: exposedPorts,
		Networks:     []string{net.Name},
		NetworkAliases: map[string][]string{
			net.Name: {alias},
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
				HostFilePath:      filepath.Join(projectRoot, "config", "auth.yaml"),
				ContainerFilePath: "/app/config/auth.yaml",
				FileMode:          0o644,
			},
			{
				HostFilePath:      filepath.Join(projectRoot, "config", "items.yaml"),
				ContainerFilePath: "/app/config/items.yaml",
				FileMode:          0o644,
			},
			{
				HostFilePath:      filepath.Join(projectRoot, "config", "users.yaml"),
				ContainerFilePath: "/app/config/users.yaml",
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

	c, err := launchPostgres(ctx, net)
	if c != nil {
		t.Cleanup(func() { require.NoError(t, testcontainers.TerminateContainer(c)) })
	}
	require.NoError(t, err)
	return c
}

func SetupKafka(ctx context.Context, net *testcontainers.DockerNetwork, t *testing.T) testcontainers.Container {
	t.Helper()

	c, err := launchKafka(ctx, net)
	if c != nil {
		t.Cleanup(func() { require.NoError(t, testcontainers.TerminateContainer(c)) })
	}
	require.NoError(t, err)
	return c
}

func SetupSMTP(ctx context.Context, net *testcontainers.DockerNetwork, t *testing.T) testcontainers.Container {
	t.Helper()

	c, err := launchSMTP(ctx, net)
	if c != nil {
		t.Cleanup(func() { require.NoError(t, testcontainers.TerminateContainer(c)) })
	}
	require.NoError(t, err)
	return c
}

func SetupAuth(ctx context.Context, net *testcontainers.DockerNetwork, t *testing.T) testcontainers.Container {
	t.Helper()
	req := buildServiceRequest(net, "auth", string(authHTTPPort))
	req.Env = serviceEnv()
	req.Env["CONFIG_SERVICE"] = "/app/config/auth.yaml"
	req.Env["JWT_REFRESH_SECRET"] = testJWTRefreshSecret
	req.Env["MAILER_BYPASS"] = "true"
	return startContainer(ctx, t, req)
}

func SetupItems(ctx context.Context, net *testcontainers.DockerNetwork, t *testing.T) testcontainers.Container {
	t.Helper()
	req := buildServiceRequest(net, "items", string(itemsHTTPPort))
	req.Env = serviceEnv()
	req.Env["CONFIG_SERVICE"] = "/app/config/items.yaml"
	return startContainer(ctx, t, req)
}

func SetupUsers(ctx context.Context, net *testcontainers.DockerNetwork, t *testing.T) testcontainers.Container {
	t.Helper()
	req := buildServiceRequest(net, "users", string(usersHTTPPort))
	req.Env = serviceEnv()
	req.Env["CONFIG_SERVICE"] = "/app/config/users.yaml"
	req.Env["AUTH_GRPC_ADDR"] = "auth:50051"
	return startContainer(ctx, t, req)
}

func serviceEnv() map[string]string {
	return map[string]string{
		"APP_ENV":           "docker",
		"CONFIG_COMMON":     "/app/config/common.yaml",
		"DB_PASSWORD":       defaultDBPassword,
		"JWT_ACCESS_SECRET": testJWTAccessSecret,
	}
}

// DumpLogsOnFailure регистрирует вывод логов контейнера при падении теста.
// Используется тестами, работающими с globalFixture (shared-контейнеры).
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

func stringPtr(value string) *string {
	return &value
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
