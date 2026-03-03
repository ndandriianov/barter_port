package bootstrap

import "barter-port/internal/libs/authkit/validators"

func InitLocalJWT() *validators.LocalJWT {
	accessSecret := GetEnv("ACCESS_SECRET", "")
	return validators.NewLocalJWT(accessSecret)
}
