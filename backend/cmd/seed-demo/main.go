package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg, err := parseConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	client := &seedClient{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		pollInterval: cfg.PollInterval,
	}

	summary, err := runSeed(ctx, client, cfg)
	if err != nil {
		log.Fatal(err)
	}

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Seed completed against %s\n%s\n", cfg.BaseURL, data)
}
