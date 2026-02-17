package telegram

import (
	"fmt"
	"log"
	"time"
)

// StartScheduler launches a background goroutine that periodically checks
// for new listings and sends them to the saved chat ID.
func (b *Bot) StartScheduler(interval time.Duration) {
	b.mu.Lock()
	b.checkInterval = interval
	b.mu.Unlock()

	go func() {
		log.Printf("Scheduler started, interval: %s", interval)

		// Run an immediate check on startup
		b.scheduledCheck()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			b.scheduledCheck()
		}
	}()
}

func (b *Bot) scheduledCheck() {
	b.mu.Lock()
	chatID := b.config.ChatID
	b.mu.Unlock()

	if chatID == 0 {
		log.Println("Scheduled check skipped: no chat ID configured (send /start to the bot first)")
		return
	}

	log.Println("Running scheduled check for new listings...")

	b.mu.Lock()
	b.lastCheck = time.Now()
	b.mu.Unlock()

	newProps := b.checkNewListings()

	if len(newProps) == 0 {
		log.Println("Scheduled check: no new properties found")
		return
	}

	log.Printf("Scheduled check: found %d new properties", len(newProps))

	text := fmt.Sprintf("*Found %d new properties:*\n\n", len(newProps))
	text += formatProperties(newProps)
	b.sendMessage(chatID, text)
}
