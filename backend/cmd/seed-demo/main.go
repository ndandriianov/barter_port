package main

import (
	"barter-port/cmd/seed-demo/seed-demo"
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

	cfg, err := seed_demo.ParseConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	client := &seed_demo.SeedClient{
		BaseURL: strings.TrimRight(cfg.BaseURL, "/"),
		HttpClient: &http.Client{
			Timeout: 15 * time.Second,
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
