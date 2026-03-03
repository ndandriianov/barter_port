package bootstrap

import (
	"barter-port/internal/libs/platform/database"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDatabase() *pgxpool.Pool {
	DbConfigPath := GetEnv("DB_CONFIG_PATH", "")
	dbConfig := database.MustLoad(DbConfigPath)
	db, err := database.NewPostgres(dbConfig)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	return db
}
