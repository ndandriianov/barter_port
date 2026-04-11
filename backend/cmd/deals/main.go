package main

import (
	dealspb "barter-port/contracts/grpc/deals/v1"
	"barter-port/internal/deals/app"
	dealssvc "barter-port/internal/deals/application/deals"
	"barter-port/internal/deals/application/offers"
	"barter-port/internal/deals/infrastructure/repository/deals"
	"barter-port/internal/deals/infrastructure/repository/drafts"
	"barter-port/internal/deals/infrastructure/repository/joins"
	offersr "barter-port/internal/deals/infrastructure/repository/offers"
	transportgrpc "barter-port/internal/deals/infrastructure/transport/grpc"
	transporthttp "barter-port/internal/deals/infrastructure/transport/http"
	dealsh "barter-port/internal/deals/infrastructure/transport/http/deals"
	draftsh "barter-port/internal/deals/infrastructure/transport/http/drafts"
	failuresh "barter-port/internal/deals/infrastructure/transport/http/failures"
	joinsh "barter-port/internal/deals/infrastructure/transport/http/joins"
	offersh "barter-port/internal/deals/infrastructure/transport/http/offers"
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

	err = bootstrap.RunMigrationsFromConfig(cfg)
	if err != nil {
		log.Fatal("deals - run migrations:", err)
	}

	db, err := bootstrap.InitDatabaseFromConfig(cfg)
	if err != nil {
		log.Fatal("failed to initialize database:", err)
	}
	defer db.Close()

	logg := logger.NewJSONLogger(slog.LevelDebug, "deals-service", "")

	offersRepo := offersr.NewRepository(db)

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

	chatsClient, chatsConn, err := app.InitChatsGRPCClient(cfg)
	if err != nil {
		logg.Warn("failed to initialize chats grpc client, deal->chat integration disabled", slog.Any("error", err))
	} else {
		defer chatsConn.Close()
	}

	offersService := offers.NewService(offersRepo, usersClient, logg)

	draftsRepo := drafts.NewRepository()
	dealsRepo := deals.NewRepository()
	joinsRepo := joins.NewRepository()
	dealsService := dealssvc.NewService(db, draftsRepo, dealsRepo, joinsRepo, offersRepo).
		WithAdminChecker(authkit.NewAdminChecker(authClient)).
		WithLogger(logg)
	if chatsClient != nil {
		dealsService = dealsService.WithChatsClient(chatsClient)
	}

	validator, err := bootstrap.InitLocalJWTFromConfig(cfg)
	if err != nil {
		log.Fatal("failed to initialize JWT validator:", err)
	}

	grpcListenAddr := cfg.DealsGRPCListenAddr
	if grpcListenAddr == "" {
		grpcListenAddr = ":50054"
	}
	listener, err := net.Listen("tcp", grpcListenAddr)
	if err != nil {
		log.Fatal("failed to listen on grpc port:", err)
	}
	grpcServer := grpc.NewServer()
	dealspb.RegisterDealsServiceServer(grpcServer, transportgrpc.NewServer(dealsService))

	go func() {
		logg.Info("gRPC server listening", slog.String("addr", grpcListenAddr))
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatal("gRPC server failed:", err)
		}
	}()

	offersHandlers := offersh.NewHandlers(offersService)
	draftsHandlers := draftsh.NewHandlers(logg, dealsService)
	dealsHandlers := dealsh.NewHandlers(logg, dealsService)
	failuresHandlers := failuresh.NewHandlers(logg, dealsService)
	joinsHandlers := joinsh.NewHandlers(logg, dealsService)
	router := transporthttp.NewRouter(logg, validator, offersHandlers, draftsHandlers, dealsHandlers, failuresHandlers, joinsHandlers)

	port := bootstrap.InitPortStringFromConfig(cfg, 8080)
	log.Println("backend listening on", port)
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

	const localPath = "./config/deals.yaml"
	if _, err := os.Stat(localPath); err == nil {
		return localPath
	}

	return os.Getenv("CONFIG_SERVICE")
}
