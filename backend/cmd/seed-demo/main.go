package main

import (
	"barter-port/cmd/seed-demo/seed-demo"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg, err := seed_demo.ParseConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	client := &seed_demo.SeedClient{
		BaseURL:      strings.TrimRight(cfg.BaseURL, "/"),
		SMTP4DevURL:  strings.TrimRight(cfg.SMTP4DevURL, "/"),
		SMTP4DevUser: cfg.SMTP4DevUser,
		SMTP4DevPass: cfg.SMTP4DevPass,
		HttpClient: &http.Client{
			Timeout: cfg.HTTPTimeout,
		},
		PollInterval: cfg.PollInterval,
	}

	summary, err := seed_demo.RunSeed(ctx, client, cfg)
	if err != nil {
		log.Fatal(err)
	}

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Seed completed against %s\n%s\n", cfg.BaseURL, data)
}
