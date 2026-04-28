package main

import (
	chatspb "barter-port/contracts/grpc/chats/v1"
	"barter-port/internal/chats/app"
	"barter-port/internal/chats/application"
	"barter-port/internal/chats/infrastructure/repository"
	transportgrpc "barter-port/internal/chats/infrastructure/transport/grpc"
	transporthttp "barter-port/internal/chats/infrastructure/transport/http"
	"barter-port/pkg/authkit"
	"barter-port/pkg/bootstrap"
	"barter-port/pkg/logger"
	"errors"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	_ = godotenv.Load()

	cfg, err := loadConfig()
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	if err = bootstrap.RunMigrationsFromConfig(cfg); err != nil {
		log.Fatal("chats - run migrations:", err)
	}

	db, err := bootstrap.InitDatabaseFromConfig(cfg)
	if err != nil {
		log.Fatal("failed to initialize database:", err)
	}
	defer db.Close()

	logg := logger.NewJSONLogger(slog.LevelDebug, "chats-service", "")

	authClient, authConn, err := app.InitAuthGRPCClient(cfg)
	if err != nil {
		log.Fatal("failed to initialize auth grpc client:", err)
	}
	defer authConn.Close()

	usersClient, usersConn, err := app.InitUsersGRPCClient(cfg)
	if err != nil {
		log.Fatal("failed to initialize users grpc client:", err)
	}
	defer usersConn.Close()

	dealsClient, dealsConn, err := app.InitDealsGRPCClient(cfg)
	if err != nil {
		log.Fatal("failed to initialize deals grpc client:", err)
	}
	defer dealsConn.Close()

	repo := repository.NewRepository(db)
	chatsService := application.NewService(repo).
		WithAdminChecker(authkit.NewAdminChecker(authClient)).
		WithDealsClient(dealsClient).
		WithUsersClient(usersClient)

	validator, err := bootstrap.InitLocalJWTFromConfig(cfg)
	if err != nil {
		log.Fatal("failed to initialize JWT validator:", err)
	}

	// Start gRPC server
	grpcListenAddr := cfg.ChatsGRPCListenAddr
	if grpcListenAddr == "" {
		grpcListenAddr = ":50053"
	}
	listener, err := net.Listen("tcp", grpcListenAddr)
	if err != nil {
		log.Fatal("failed to listen on grpc port:", err)
	}
	grpcServer := grpc.NewServer()
	chatspb.RegisterChatsServiceServer(grpcServer, transportgrpc.NewServer(chatsService))

	go func() {
		logg.Info("gRPC server listening", slog.String("addr", grpcListenAddr))
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatal("gRPC server failed:", err)
		}
	}()

	// Start HTTP server
	handlers := transporthttp.NewHandlers(logg, chatsService, usersClient)
	router := transporthttp.NewRouter(logg, validator, handlers)

	port := bootstrap.InitPortStringFromConfig(cfg, 8083)
	logg.Info("HTTP server listening", slog.String("addr", port))
	log.Fatal(http.ListenAndServe(port, router))
}

func loadConfig() (bootstrap.Config, error) {
	cfg, err := bootstrap.LoadConfig(bootstrap.ConfigOptions{
		CommonPath:  os.Getenv("CONFIG_COMMON"),
		ServicePath: resolveServiceConfigPath(),
		AppEnv:      os.Getenv("APP_ENV"),
	})
	if err != nil {
		return bootstrap.Config{}, errors.New("failed to load config: " + err.Error())
	}

	return cfg, nil
}

func resolveServiceConfigPath() string {
	if path := os.Getenv("CONFIG_SERVICE"); path != "" {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	const localPath = "./config/chats.yaml"
	if _, err := os.Stat(localPath); err == nil {
		return localPath
	}

	return os.Getenv("CONFIG_SERVICE")
}
