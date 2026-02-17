package main

import (
	"log"
	"os"
	"time"
	"web_parser/pkg/scraper"
	"web_parser/pkg/telegram"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	checkInterval := 24 * time.Hour
	if v := os.Getenv("CHECK_INTERVAL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			log.Fatalf("Invalid CHECK_INTERVAL %q: %v", v, err)
		}
		checkInterval = d
	}

	s := scraper.NewSSGEScraper()

	bot, err := telegram.NewBot(botToken, s, "data")
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	bot.StartScheduler(checkInterval)

	log.Println("Bot started successfully. Press Ctrl+C to stop.")

	if err := bot.Start(); err != nil {
		log.Fatalf("Bot error: %v", err)
	}
}
