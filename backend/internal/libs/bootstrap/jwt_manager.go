package bootstrap

import (
	"barter-port/internal/libs/jwt"
	"time"
)

func InitJWTManager() *jwt.Manager {
	accessSecret := GetEnv("ACCESS_SECRET", "")
	refreshSecret := GetEnv("REFRESH_SECRET", "")
	accessTTL := GetEnv("ACCESS_TTL", "")
	refreshTTL := GetEnv("REFRESH_TTL", "")
	accessTTLMinutes := time.Duration(mustInt(accessTTL)) * time.Minute
	refreshTTLMinutes := time.Duration(mustInt(refreshTTL)) * time.Minute

	return jwt.NewManager(jwt.Config{
		AccessSecret:  accessSecret,
		RefreshSecret: refreshSecret,
		AccessTTL:     accessTTLMinutes,
		RefreshTTL:    refreshTTLMinutes,
	})
}
