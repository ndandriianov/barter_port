package integration

import (
	"barter-port/pkg/db"
	"context"
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

	if opts.NeedPostgres {
		f.Postgres = SetupPostgres(ctx, net, t)
	}
	if opts.NeedKafka {
		f.Kafka = SetupKafka(ctx, net, t)
	}
	if opts.NeedSMTP {
		f.SMTP = SetupSMTP(ctx, net, t)
	}
	if opts.NeedAuth {
		f.Auth = SetupAuth(ctx, net, t)
		f.AuthURL = containerBaseURL(ctx, t, f.Auth, authHTTPPort)
	}
	if opts.NeedItems {
		f.Items = SetupItems(ctx, net, t)
		f.ItemsURL = containerBaseURL(ctx, t, f.Items, itemsHTTPPort)
	}
	if opts.NeedUsers {
		f.Users = SetupUsers(ctx, net, t)
		f.UsersURL = containerBaseURL(ctx, t, f.Users, usersHTTPPort)
	}

	return f
}

func SetupNetwork(ctx context.Context, t *testing.T) *testcontainers.DockerNetwork {
	t.Helper()

	net, err := network.New(ctx)
	require.NoError(t, err)
	testcontainers.CleanupNetwork(t, net)

	return net
}

func SetupPostgres(ctx context.Context, net *testcontainers.DockerNetwork, t *testing.T) testcontainers.Container {
	t.Helper()

	projectRoot := projectRoot(t)
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

	postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testcontainers.TerminateContainer(postgres))
	})

	return postgres
}

func SetupKafka(ctx context.Context, net *testcontainers.DockerNetwork, t *testing.T) testcontainers.Container {
	t.Helper()

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

	kafka, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testcontainers.TerminateContainer(kafka))
	})

	return kafka
}

func SetupSMTP(ctx context.Context, net *testcontainers.DockerNetwork, t *testing.T) testcontainers.Container {
	t.Helper()

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

	smtp, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, testcontainers.TerminateContainer(smtp))
	})

	return smtp
}

func SetupAuth(ctx context.Context, net *testcontainers.DockerNetwork, t *testing.T) testcontainers.Container {
	t.Helper()

	req := serviceContainerRequest(t, net, "auth", string(authHTTPPort))
	req.Env = serviceEnv()
	req.Env["CONFIG_SERVICE"] = "/app/config/auth.yaml"
	req.Env["JWT_REFRESH_SECRET"] = testJWTRefreshSecret
	req.Env["MAILER_BYPASS"] = "true"

	return startContainer(ctx, t, req)
}

func SetupItems(ctx context.Context, net *testcontainers.DockerNetwork, t *testing.T) testcontainers.Container {
	t.Helper()

	req := serviceContainerRequest(t, net, "items", string(itemsHTTPPort))
	req.Env = serviceEnv()
	req.Env["CONFIG_SERVICE"] = "/app/config/items.yaml"

	return startContainer(ctx, t, req)
}

func SetupUsers(ctx context.Context, net *testcontainers.DockerNetwork, t *testing.T) testcontainers.Container {
	t.Helper()

	req := serviceContainerRequest(t, net, "users", string(usersHTTPPort))
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

func serviceContainerRequest(t *testing.T, net *testcontainers.DockerNetwork, service string, exposedPorts ...string) testcontainers.ContainerRequest {
	t.Helper()

	projectRoot := projectRoot(t)
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

func startContainer(ctx context.Context, t *testing.T, req testcontainers.ContainerRequest) testcontainers.Container {
	t.Helper()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	// При любом исходе — если контейнер поднялся, регистрируем вывод логов при падении теста.
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

func containerBaseURL(ctx context.Context, t *testing.T, c testcontainers.Container, port nat.Port) string {
	t.Helper()

	host, err := c.Host(ctx)
	require.NoError(t, err)

	mappedPort, err := c.MappedPort(ctx, port)
	require.NoError(t, err)

	return fmt.Sprintf("http://%s:%s", host, mappedPort.Port())
}

func projectRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)

	return filepath.Clean(filepath.Join(wd, ".."))
}

func stringPtr(value string) *string {
	return &value
}
