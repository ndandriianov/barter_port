package main

import (
	"barter-port/internal/items/repository"
	"barter-port/internal/items/service"
	"barter-port/internal/items/transport"
	"barter-port/internal/libs/bootstrap"
	"barter-port/internal/libs/platform/logger"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

//go:generate bash ../../scripts/generate-swagger-items.sh

// @title Barter Port API
// @version 1.0.0
// @description API for Barter Port
// @host localhost:80
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	_ = godotenv.Load()

	cfg, err := bootstrap.LoadConfig(bootstrap.ConfigOptions{
		CommonPath:  os.Getenv("CONFIG_COMMON"),
		ServicePath: os.Getenv("CONFIG_SERVICE"),
		AppEnv:      os.Getenv("APP_ENV"),
	})
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	db, err := bootstrap.InitDatabaseFromConfig(cfg)
	if err != nil {
		log.Fatal("failed to initialize database:", err)
	}
	defer db.Close()

	logg := logger.NewJSONLogger(slog.LevelDebug, "items-service", "")

	itemRepo := repository.NewItemRepository(db)
	itemService := service.NewItemService(itemRepo, logg)

	validator, err := bootstrap.InitLocalJWTFromConfig(cfg)
	if err != nil {
		log.Fatal("failed to initialize JWT validator:", err)
	}

	handlers := transport.NewHandlers(itemService)
	router := transport.NewRouter(logg, validator, handlers)

	port := bootstrap.InitPortStringFromConfig(cfg, 8080)
	log.Println("backend listening on", port)
	log.Fatal(http.ListenAndServe(port, router))
}
