package bootstrap

import (
	"barter-port/internal/libs/authkit/validators"
	"fmt"
)

func InitLocalJWTFromConfig(cfg Config) (*validators.LocalJWT, error) {
	if cfg.JWT.AccessSecret == "" {
		return nil, fmt.Errorf("jwt access secret is empty")
	}
	return validators.NewLocalJWT(cfg.JWT.AccessSecret), nil
}
