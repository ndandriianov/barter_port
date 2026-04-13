package main

import (
	dealspb "barter-port/contracts/grpc/deals/v1"
	"barter-port/internal/deals/app"
	dealssvc "barter-port/internal/deals/application/deals"
	failuressvc "barter-port/internal/deals/application/failures"
	joinssvc "barter-port/internal/deals/application/joins"
	offergroupssvc "barter-port/internal/deals/application/offergroups"
	"barter-port/internal/deals/application/offers"
	reviewssvc "barter-port/internal/deals/application/reviews"
	"barter-port/internal/deals/infrastructure/repository/deals"
	"barter-port/internal/deals/infrastructure/repository/drafts"
	failuresrepo "barter-port/internal/deals/infrastructure/repository/failures"
	"barter-port/internal/deals/infrastructure/repository/joins"
	offergroupsrepo "barter-port/internal/deals/infrastructure/repository/offergroups"
	offersr "barter-port/internal/deals/infrastructure/repository/offers"
	reviewsrepo "barter-port/internal/deals/infrastructure/repository/reviews"
	offerphotostorage "barter-port/internal/deals/infrastructure/storage/offerphoto"
	transportgrpc "barter-port/internal/deals/infrastructure/transport/grpc"
	transporthttp "barter-port/internal/deals/infrastructure/transport/http"
	dealsh "barter-port/internal/deals/infrastructure/transport/http/deals"
	draftsh "barter-port/internal/deals/infrastructure/transport/http/drafts"
	failuresh "barter-port/internal/deals/infrastructure/transport/http/failures"
	joinsh "barter-port/internal/deals/infrastructure/transport/http/joins"
	offergroupsh "barter-port/internal/deals/infrastructure/transport/http/offergroups"
	offersh "barter-port/internal/deals/infrastructure/transport/http/offers"
	reviewsh "barter-port/internal/deals/infrastructure/transport/http/reviews"
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
	offerPhotoStorage, err := offerphotostorage.NewStorage(offerphotostorage.Config{
		Endpoint:        cfg.Storage.Endpoint,
		PublicBaseURL:   cfg.Storage.PublicBaseURL,
		Bucket:          cfg.Storage.OfferPhotoBucket,
		AccessKeyID:     cfg.Storage.AccessKeyID,
		SecretAccessKey: cfg.Storage.SecretAccessKey,
		Region:          cfg.Storage.Region,
	})
	if err != nil {
		log.Fatal("failed to initialize offer photo storage:", err)
	}

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

	offersService := offers.NewService(db, offersRepo, usersClient, offerPhotoStorage, logg)

	draftsRepo := drafts.NewRepository()
	dealsRepo := deals.NewRepository()
	failuresRepo := failuresrepo.NewRepository(dealsRepo)
	joinsRepo := joins.NewRepository()
	offerGroupsRepo := offergroupsrepo.NewRepository(db)
	reviewsRepo := reviewsrepo.NewRepository(dealsRepo)
	dealsService := dealssvc.NewService(db, draftsRepo, dealsRepo, failuresRepo, joinsRepo, offersRepo).
		WithAdminChecker(authkit.NewAdminChecker(authClient)).
		WithLogger(logg)
	if chatsClient != nil {
		dealsService = dealsService.WithChatsClient(chatsClient)
	}
	offerGroupsService := offergroupssvc.NewService(db, offerGroupsRepo, offersRepo, dealsService, usersClient)
	failuresService := failuressvc.NewService(dealsService, failuresRepo)
	joinsService := joinssvc.NewService(dealsService)
	reviewsService := reviewssvc.NewService(dealsService, reviewsRepo)

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
	offerGroupsHandlers := offergroupsh.NewHandlers(logg, offerGroupsService)
	draftsHandlers := draftsh.NewHandlers(logg, dealsService)
	dealsHandlers := dealsh.NewHandlers(logg, dealsService)
	failuresHandlers := failuresh.NewHandlers(logg, failuresService)
	joinsHandlers := joinsh.NewHandlers(logg, joinsService)
	reviewsHandlers := reviewsh.NewHandlers(logg, reviewsService)
	router := transporthttp.NewRouter(logg, validator, offersHandlers, offerGroupsHandlers, draftsHandlers, dealsHandlers, failuresHandlers, joinsHandlers, reviewsHandlers)

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
