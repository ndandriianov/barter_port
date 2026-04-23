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
	defaultTimeout       = 2 * time.Minute
	defaultPollInterval  = 500 * time.Millisecond
	defaultAdminEmail    = "admin@barterport.com"
	defaultAdminPassword = "admin"
)

type SeedConfig struct {
	BaseURL       string
	Password      string
	Timeout       time.Duration
	PollInterval  time.Duration
	AdminEmail    string
	AdminPassword string
}

func ParseConfig() (SeedConfig, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	baseURL := fs.String("base-url", firstNonEmpty(os.Getenv("SEED_BASE_URL"), defaultBaseURL), "HTTP base URL for the app gateway")
	password := fs.String("password", firstNonEmpty(os.Getenv("SEED_PASSWORD"), defaultPassword), "Password for seeded demo users")
	timeout := fs.Duration("timeout", durationFromEnv("SEED_TIMEOUT", defaultTimeout), "Overall seed timeout")
	pollInterval := fs.Duration("poll-interval", durationFromEnv("SEED_POLL_INTERVAL", defaultPollInterval), "Polling interval for async readiness checks")
	adminEmail := fs.String("admin-email", firstNonEmpty(os.Getenv("SEED_ADMIN_EMAIL"), defaultAdminEmail), "Admin account email for moderation endpoints")
	adminPassword := fs.String("admin-password", firstNonEmpty(os.Getenv("SEED_ADMIN_PASSWORD"), defaultAdminPassword), "Admin account password")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return SeedConfig{}, err
	}

	if strings.TrimSpace(*baseURL) == "" {
		return SeedConfig{}, errors.New("base URL must not be empty")
	}

	return SeedConfig{
		BaseURL:       *baseURL,
		Password:      *password,
		Timeout:       *timeout,
		PollInterval:  *pollInterval,
		AdminEmail:    *adminEmail,
		AdminPassword: *adminPassword,
	}, nil
}
