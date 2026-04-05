package main

import (
	"barter-port/internal/deals/app"
	dealssvc "barter-port/internal/deals/application/deals"
	"barter-port/internal/deals/application/offers"
	"barter-port/internal/deals/infrastructure/repository/deals"
	"barter-port/internal/deals/infrastructure/repository/drafts"
	offersr "barter-port/internal/deals/infrastructure/repository/offers"
	transporthttp "barter-port/internal/deals/infrastructure/transport/http"
	dealsh "barter-port/internal/deals/infrastructure/transport/http/deals"
	draftsh "barter-port/internal/deals/infrastructure/transport/http/drafts"
	offersh "barter-port/internal/deals/infrastructure/transport/http/offers"
	"barter-port/pkg/bootstrap"
	"barter-port/pkg/logger"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
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

	usersClient, conn, err := app.InitUsersGRPCClient(cfg)
	if err != nil {
		log.Fatal("failed to initialize users grpc client:", err)
	}
	defer conn.Close()

	offersService := offers.NewService(offersRepo, usersClient, logg)

	draftsRepo := drafts.NewRepository()
	dealsRepo := deals.NewRepository()
	dealsService := dealssvc.NewService(db, draftsRepo, dealsRepo)

	validator, err := bootstrap.InitLocalJWTFromConfig(cfg)
	if err != nil {
		log.Fatal("failed to initialize JWT validator:", err)
	}

	offersHandlers := offersh.NewHandlers(offersService)
	draftsHandlers := draftsh.NewHandlers(logg, dealsService)
	dealsHandlers := dealsh.NewHandlers(logg, dealsService)
	router := transporthttp.NewRouter(logg, validator, offersHandlers, draftsHandlers, dealsHandlers)

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
