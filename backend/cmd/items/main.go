package main

import (
	"barter-port/internal/items/repository"
	"barter-port/internal/items/service"
	httptransport "barter-port/internal/items/transport/http"
	"barter-port/pkg/bootstrap"
	"barter-port/pkg/logger"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

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

	handlers := httptransport.NewHandlers(itemService)
	router := httptransport.NewRouter(logg, validator, handlers)

	port := bootstrap.InitPortStringFromConfig(cfg, 8080)
	log.Println("backend listening on", port)
	log.Fatal(http.ListenAndServe(port, router))
}
