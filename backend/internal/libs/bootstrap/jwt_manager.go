package bootstrap

import (
	"barter-port/internal/libs/jwt"
	"fmt"
)

func InitJWTManagerFromConfig(cfg Config) (*jwt.Manager, error) {
	if cfg.JWT.AccessSecret == "" {
		return nil, fmt.Errorf("jwt access secret is empty")
	}
	if cfg.JWT.RefreshSecret == "" {
		return nil, fmt.Errorf("jwt refresh secret is empty")
	}
	if cfg.JWT.AccessTTL <= 0 {
		return nil, fmt.Errorf("jwt access ttl must be greater than 0")
	}
	if cfg.JWT.RefreshTTL <= 0 {
		return nil, fmt.Errorf("jwt refresh ttl must be greater than 0")
	}

	return jwt.NewManager(jwt.Config{
		AccessSecret:  cfg.JWT.AccessSecret,
		RefreshSecret: cfg.JWT.RefreshSecret,
		AccessTTL:     cfg.JWT.AccessTTL,
		RefreshTTL:    cfg.JWT.RefreshTTL,
	}), nil
}
