package main

import (
	"barter-port/internal/items/repository"
	"barter-port/internal/items/service"
	"barter-port/internal/items/transport"
	"barter-port/internal/libs/jwt"
	"barter-port/internal/libs/platform/database"
	"barter-port/internal/libs/platform/logger"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	DbConfigPath := getEnv("DB_CONFIG_PATH", "")
	dbConfig := database.MustLoad(DbConfigPath)
	db, err := database.NewPostgres(dbConfig)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer db.Close()

	accessSecret := getEnv("ACCESS_SECRET", "")
	refreshSecret := getEnv("REFRESH_SECRET", "")
	accessTTL := getEnv("ACCESS_TTL", "")
	refreshTTL := getEnv("REFRESH_TTL", "")
	accessTTLMinutes := time.Duration(mustInt(accessTTL)) * time.Minute
	refreshTTLMinutes := time.Duration(mustInt(refreshTTL)) * time.Minute

	itemRepo := repository.NewItemRepository(db)

	jwtManager := jwt.NewManager(jwt.Config{
		AccessSecret:  accessSecret,
		RefreshSecret: refreshSecret,
		AccessTTL:     accessTTLMinutes,
		RefreshTTL:    refreshTTLMinutes,
	})

	logg := logger.NewJSONLogger(slog.LevelDebug, "items-service", "")

	itemService := service.NewItemService(itemRepo)
	handlers := transport.NewHandlers(itemService)
	router := transport.NewRouter(logg, jwtManager, handlers)

	addr := ":8080"
	log.Println("backend listening on", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func mustInt(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("invalid integer value: %s", s)
	}
	return v
}
