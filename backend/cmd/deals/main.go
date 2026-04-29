package main

import (
	dealspb "barter-port/contracts/grpc/deals/v1"
	"barter-port/internal/deals/app"
	dealssvc "barter-port/internal/deals/application/deals"
	failuressvc "barter-port/internal/deals/application/failures"
	joinssvc "barter-port/internal/deals/application/joins"
	offerreportssvc "barter-port/internal/deals/application/offer-reports"
	offergroupssvc "barter-port/internal/deals/application/offergroups"
	"barter-port/internal/deals/application/offers"
	reviewssvc "barter-port/internal/deals/application/reviews"
	statisticssvc "barter-port/internal/deals/application/statistics"
	penaltyoutbox "barter-port/internal/deals/infrastructure/kafka/producer/penalty-outbox"
	"barter-port/internal/deals/infrastructure/repository/deals"
	"barter-port/internal/deals/infrastructure/repository/drafts"
	failuresrepo "barter-port/internal/deals/infrastructure/repository/failures"
	"barter-port/internal/deals/infrastructure/repository/joins"
	offerreportsrepo "barter-port/internal/deals/infrastructure/repository/offer-reports"
	offergroupsrepo "barter-port/internal/deals/infrastructure/repository/offergroups"
	offersr "barter-port/internal/deals/infrastructure/repository/offers"
	offerreportoutboxrepo "barter-port/internal/deals/infrastructure/repository/reputation-events-outbox"
	reviewsrepo "barter-port/internal/deals/infrastructure/repository/reviews"
	statsrepo "barter-port/internal/deals/infrastructure/repository/statistics"
	itemphotostorage "barter-port/internal/deals/infrastructure/storage/itemphoto"
	offerphotostorage "barter-port/internal/deals/infrastructure/storage/offerphoto"
	transportgrpc "barter-port/internal/deals/infrastructure/transport/grpc"
	transporthttp "barter-port/internal/deals/infrastructure/transport/http"
	dealsh "barter-port/internal/deals/infrastructure/transport/http/deals"
	draftsh "barter-port/internal/deals/infrastructure/transport/http/drafts"
	failuresh "barter-port/internal/deals/infrastructure/transport/http/failures"
	favouritesh "barter-port/internal/deals/infrastructure/transport/http/favourites"
	joinsh "barter-port/internal/deals/infrastructure/transport/http/joins"
	offerreportsh "barter-port/internal/deals/infrastructure/transport/http/offer-reports"
	offergroupsh "barter-port/internal/deals/infrastructure/transport/http/offergroups"
	offersh "barter-port/internal/deals/infrastructure/transport/http/offers"
	reviewsh "barter-port/internal/deals/infrastructure/transport/http/reviews"
	statisticsh "barter-port/internal/deals/infrastructure/transport/http/statistics"
	tagsh "barter-port/internal/deals/infrastructure/transport/http/tags"
	"barter-port/pkg/authkit"
	"barter-port/pkg/bootstrap"
	"barter-port/pkg/kafkax"
	"barter-port/pkg/logger"
	"context"
	"errors"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/oklog/run"
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
	itemPhotoStorage, err := itemphotostorage.NewStorage(itemphotostorage.Config{
		Endpoint:        cfg.Storage.Endpoint,
		PublicBaseURL:   cfg.Storage.PublicBaseURL,
		Bucket:          cfg.Storage.OfferPhotoBucket,
		AccessKeyID:     cfg.Storage.AccessKeyID,
		SecretAccessKey: cfg.Storage.SecretAccessKey,
		Region:          cfg.Storage.Region,
	})
	if err != nil {
		log.Fatal("failed to initialize item photo storage:", err)
	}

	authClient, authConn, err := app.InitAuthGRPCClient(cfg)
	if err != nil {
		log.Fatal("failed to initialize auth grpc client:", err)
	}
	defer func() {
		if closeErr := authConn.Close(); closeErr != nil {
			logg.Warn("failed to close auth grpc connection", slog.Any("error", closeErr))
		}
	}()

	usersClient, usersConn, err := app.InitUsersGRPCClient(cfg)
	if err != nil {
		log.Fatal("failed to initialize users grpc client:", err)
	}
	defer func() {
		if closeErr := usersConn.Close(); closeErr != nil {
			logg.Warn("failed to close users grpc connection", slog.Any("error", closeErr))
		}
	}()

	chatsClient, chatsConn, err := app.InitChatsGRPCClient(cfg)
	if err != nil {
		logg.Warn("failed to initialize chats grpc client, deal->chat integration disabled", slog.Any("error", err))
	} else {
		defer func() {
			if closeErr := chatsConn.Close(); closeErr != nil {
				logg.Warn("failed to close chats grpc connection", slog.Any("error", closeErr))
			}
		}()
	}

	adminChecker := authkit.NewAdminChecker(authClient)
	draftsRepo := drafts.NewRepository()
	offersService := offers.NewService(db, offersRepo, draftsRepo, usersClient, offerPhotoStorage, adminChecker, logg)

	dealsRepo := deals.NewRepository()
	failuresRepo := failuresrepo.NewRepository(dealsRepo)
	joinsRepo := joins.NewRepository()
	offerGroupsRepo := offergroupsrepo.NewRepository(db)
	reviewsRepo := reviewsrepo.NewRepository(dealsRepo)
	dealsService := dealssvc.NewService(db, draftsRepo, dealsRepo, failuresRepo, joinsRepo, offersRepo, itemPhotoStorage).
		WithReputationOutbox(offerreportoutboxrepo.NewRepository()).
		WithReputationRewardPoints(cfg.Reputation.DealCompletionRewardPoints, cfg.Reputation.ReviewCreationRewardPoints).
		WithUsersClient(usersClient).
		WithAdminChecker(adminChecker).
		WithLogger(logg)
	if chatsClient != nil {
		dealsService = dealsService.WithChatsClient(chatsClient)
	}
	offerGroupsService := offergroupssvc.NewService(db, offerGroupsRepo, offersRepo, dealsService, usersClient)
	failuresService := failuressvc.NewService(dealsService, failuresRepo)
	joinsService := joinssvc.NewService(dealsService)
	reviewsService := reviewssvc.NewService(dealsService, reviewsRepo)

	// Offer reports
	offerReportsRepo := offerreportsrepo.NewRepository(db)
	offerReportOutboxRepo := dealsService.ReputationOutboxRepository()
	offerReportsService := offerreportssvc.NewService(db, offersRepo, offerReportsRepo, offerReportOutboxRepo, adminChecker, logg)

	// Penalty outbox Kafka producer
	topicInitCtx, cancelTopicInit := context.WithTimeout(context.Background(), cfg.Kafka.WriteTimeout)
	defer cancelTopicInit()

	reputationTopic := cfg.Kafka.ReputationTopic
	if reputationTopic == "" {
		reputationTopic = cfg.Kafka.OfferReportPenaltyTopic
	}
	if reputationTopic == "" {
		log.Fatal("failed to initialize reputation topic: kafka.reputation_topic is not configured")
	}

	if err = kafkax.EnsureTopic(topicInitCtx, cfg.Kafka.Brokers, reputationTopic, 1, 1); err != nil {
		log.Fatal("failed to ensure reputation topic:", err)
	}

	kafkaWriter := kafkax.NewWriter(cfg.Kafka.Brokers, reputationTopic)
	penaltyPublisher := kafkax.NewOutboxPublisher(
		kafkaWriter,
		logg,
		cfg.Kafka.Brokers,
		reputationTopic,
		cfg.Kafka.BatchSize,
		cfg.Kafka.PollInterval,
		cfg.Kafka.WriteTimeout,
	)
	penaltyProducer := penaltyoutbox.NewPenaltyOutboxProducer(db, offerReportOutboxRepo, logg, penaltyPublisher)

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

	statisticsRepository := statsrepo.NewRepository(db)
	statisticsService := statisticssvc.NewService(statisticsRepository).WithAuthClient(authClient)

	offersHandlers := offersh.NewHandlers(offersService)
	offerGroupsHandlers := offergroupsh.NewHandlers(logg, offerGroupsService)
	draftsHandlers := draftsh.NewHandlers(logg, dealsService)
	dealsHandlers := dealsh.NewHandlers(logg, dealsService)
	failuresHandlers := failuresh.NewHandlers(logg, failuresService)
	joinsHandlers := joinsh.NewHandlers(logg, joinsService)
	reviewsHandlers := reviewsh.NewHandlers(logg, reviewsService)
	offerReportsHandlers := offerreportsh.NewHandlers(offerReportsService)
	statisticsHandlers := statisticsh.NewHandlers(logg, statisticsService)
	tagsHandlers := tagsh.NewHandlers(offersService)
	favouritesHandlers := favouritesh.NewHandlers(offersService)
	router := transporthttp.NewRouter(logg, validator, offersHandlers, favouritesHandlers, offerGroupsHandlers, draftsHandlers, dealsHandlers, failuresHandlers, joinsHandlers, reviewsHandlers, offerReportsHandlers, statisticsHandlers, tagsHandlers)

	port := bootstrap.InitPortStringFromConfig(cfg, 8080)
	httpServer := &http.Server{
		Addr:    port,
		Handler: router,
	}

	var g run.Group
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g.Add(func() error {
		logg.Info("gRPC server listening", slog.String("addr", grpcListenAddr))
		return grpcServer.Serve(listener)
	}, func(error) {
		grpcServer.GracefulStop()
	})

	g.Add(func() error {
		return penaltyProducer.Run(ctx)
	}, func(error) {
		cancel()
		_ = penaltyProducer.Close()
	})

	g.Add(func() error {
		log.Println("deals http server listening on", port)
		return httpServer.ListenAndServe()
	}, func(error) {
		_ = httpServer.Close()
	})

	if err := g.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("deals service failed:", err)
	}
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
