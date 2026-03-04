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

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	db := bootstrap.InitDatabase()
	defer db.Close()

	logg := logger.NewJSONLogger(slog.LevelDebug, "items-service", "")

	itemRepo := repository.NewItemRepository(db)
	itemService := service.NewItemService(itemRepo, logg)

	validator := bootstrap.InitLocalJWT()
	handlers := transport.NewHandlers(itemService)

	router := transport.NewRouter(logg, validator, handlers)

	addr := ":8080"
	log.Println("backend listening on", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}
