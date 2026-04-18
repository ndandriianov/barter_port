package seed_demo

import (
	"io"
	"os"
	"strings"
	"time"
)

func containsStatus(statuses []int, status int) bool {
	for _, expected := range statuses {
		if expected == status {
			return true
		}
	}

	return false
}

func closeBody(closer io.Closer) {
	_ = closer.Close()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}

	return ""
}

func durationFromEnv(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}

	return value
}
