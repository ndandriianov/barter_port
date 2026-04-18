package seed_demo

import (
	"errors"
	"flag"
	"os"
	"strings"
	"time"
)

const (
	defaultBaseURL       = "http://localhost:80"
	defaultPassword      = "password123"
	defaultAvatarBaseURL = "http://localhost:8333/avatars"
	defaultTimeout       = 2 * time.Minute
	defaultPollInterval  = 500 * time.Millisecond
)

type SeedConfig struct {
	BaseURL       string
	Password      string
	AvatarBaseURL string
	Timeout       time.Duration
	PollInterval  time.Duration
}

func ParseConfig() (SeedConfig, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	baseURL := fs.String("base-url", firstNonEmpty(os.Getenv("SEED_BASE_URL"), defaultBaseURL), "HTTP base URL for the app gateway")
	password := fs.String("password", firstNonEmpty(os.Getenv("SEED_PASSWORD"), defaultPassword), "Password for seeded demo users")
	avatarBaseURL := fs.String("avatar-base-url", firstNonEmpty(os.Getenv("SEED_AVATAR_BASE_URL"), defaultAvatarBaseURL), "Base URL for demo avatars")
	timeout := fs.Duration("timeout", durationFromEnv("SEED_TIMEOUT", defaultTimeout), "Overall seed timeout")
	pollInterval := fs.Duration("poll-interval", durationFromEnv("SEED_POLL_INTERVAL", defaultPollInterval), "Polling interval for async readiness checks")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return SeedConfig{}, err
	}

	if strings.TrimSpace(*baseURL) == "" {
		return SeedConfig{}, errors.New("base URL must not be empty")
	}

	return SeedConfig{
		BaseURL:       *baseURL,
		Password:      *password,
		AvatarBaseURL: strings.TrimRight(*avatarBaseURL, "/"),
		Timeout:       *timeout,
		PollInterval:  *pollInterval,
	}, nil
}
