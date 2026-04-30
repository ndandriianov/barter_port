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
	defaultSMTP4DevURL   = "http://localhost:5005"
	defaultPassword      = "password123"
	defaultTimeout       = 2 * time.Minute
	defaultHTTPTimeout   = 60 * time.Second
	defaultPollInterval  = 500 * time.Millisecond
	defaultAdminEmail    = "admin@barterport.com"
	defaultAdminPassword = "admin123"
)

type SeedConfig struct {
	BaseURL       string
	SMTP4DevURL   string
	SMTP4DevUser  string
	SMTP4DevPass  string
	Password      string
	Timeout       time.Duration
	HTTPTimeout   time.Duration
	PollInterval  time.Duration
	AdminEmail    string
	AdminPassword string
}

func ParseConfig() (SeedConfig, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	baseURL := fs.String("base-url", firstNonEmpty(os.Getenv("SEED_BASE_URL"), defaultBaseURL), "HTTP base URL for the app gateway")
	smtp4devURLDefault := defaultSMTP4DevURL
	if rawSMTP4DevURL, ok := os.LookupEnv("SEED_SMTP4DEV_URL"); ok {
		smtp4devURLDefault = rawSMTP4DevURL
	}
	smtp4devURL := fs.String("smtp4dev-url", smtp4devURLDefault, "HTTP base URL for smtp4dev API/UI")
	smtp4devUser := fs.String("smtp4dev-user", strings.TrimSpace(os.Getenv("SEED_SMTP4DEV_USER")), "Optional smtp4dev HTTP basic auth username")
	smtp4devPass := fs.String("smtp4dev-password", os.Getenv("SEED_SMTP4DEV_PASSWORD"), "Optional smtp4dev HTTP basic auth password")
	password := fs.String("password", firstNonEmpty(os.Getenv("SEED_PASSWORD"), defaultPassword), "Password for seeded demo users")
	timeout := fs.Duration("timeout", durationFromEnv("SEED_TIMEOUT", defaultTimeout), "Overall seed timeout")
	httpTimeout := fs.Duration("http-timeout", durationFromEnv("SEED_HTTP_TIMEOUT", defaultHTTPTimeout), "Timeout for a single HTTP request")
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
		SMTP4DevURL:   *smtp4devURL,
		SMTP4DevUser:  *smtp4devUser,
		SMTP4DevPass:  *smtp4devPass,
		Password:      *password,
		Timeout:       *timeout,
		HTTPTimeout:   *httpTimeout,
		PollInterval:  *pollInterval,
		AdminEmail:    *adminEmail,
		AdminPassword: *adminPassword,
	}, nil
}
