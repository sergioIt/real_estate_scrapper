package telegram

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"web_parser/pkg/scraper"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// UserConfig stores user preferences and state, persisted as JSON.
type UserConfig struct {
	ChatID  int64              `json:"chat_id"`
	Filters scraper.FilterParams `json:"filters"`
	SeenIDs []string           `json:"seen_ids"`
}

type Bot struct {
	api        *tgbotapi.BotAPI
	scraper    scraper.RealEstateScraper
	config     UserConfig
	configPath string
	mu         sync.Mutex
}

// NewBot creates a new Telegram bot instance.
// dataDir is the path to the directory where config is stored (e.g. "data/").
func NewBot(token string, s scraper.RealEstateScraper, dataDir string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	log.Printf("Authorized on account %s", api.Self.UserName)

	b := &Bot{
		api:        api,
		scraper:    s,
		configPath: filepath.Join(dataDir, "filters.json"),
	}

	// Set default filters
	b.config.Filters = scraper.FilterParams{
		Type:         "For sale",
		PropertyType: "Flat",
		City:         "Tbilisi",
	}

	if err := b.loadConfig(); err != nil {
		log.Printf("No existing config, using defaults: %v", err)
	}

	return b, nil
}

// Start begins listening for updates (blocking).
func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			b.handleCommand(update.Message)
		}
	}

	return nil
}

func (b *Bot) handleCommand(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	command := message.Command()
	args := message.CommandArguments()

	log.Printf("Received command: %s from chat: %d", command, chatID)

	switch command {
	case "start":
		b.handleStart(chatID)
	case "latest":
		b.handleLatest(chatID, args)
	case "setcity":
		b.handleSetCity(chatID, args)
	case "setsqm":
		b.handleSetSqm(chatID, args)
	case "setarea":
		b.handleSetArea(chatID, args)
	case "filters":
		b.handleFilters(chatID)
	case "check":
		b.handleCheck(chatID)
	case "help":
		b.handleHelp(chatID)
	default:
		b.sendMessage(chatID, "Unknown command. Use /help to see available commands.")
	}
}

func (b *Bot) handleStart(chatID int64) {
	b.mu.Lock()
	b.config.ChatID = chatID
	b.mu.Unlock()
	b.saveConfig()

	text := `Welcome to Home.ss.ge Property Bot!

Available commands:
/latest - Show latest properties
/setcity - Set city filter (e.g. /setcity Tbilisi)
/setsqm - Set price per m² range (e.g. /setsqm 800 1200)
/setarea - Set area range (e.g. /setarea 40 80)
/filters - Show current filters
/check - Check for new listings now
/help - Show this help message`

	b.sendMessage(chatID, text)
}

func (b *Bot) handleLatest(chatID int64, args string) {
	limit := 10
	if args != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(args)); err == nil && n > 0 {
			limit = n
		}
	}

	b.mu.Lock()
	filters := b.config.Filters
	b.mu.Unlock()
	filters.Limit = limit

	b.sendMessage(chatID, fmt.Sprintf("Fetching latest %d properties...", limit))

	properties, err := b.scraper.Scrape(filters)
	if err != nil {
		log.Printf("Error scraping properties: %v", err)
		b.sendMessage(chatID, "Sorry, couldn't fetch properties right now. Please try again later.")
		return
	}

	if len(properties) == 0 {
		b.sendMessage(chatID, "No properties found.")
		return
	}

	text := fmt.Sprintf("*Latest %d Properties*\n\n", len(properties))
	text += formatProperties(properties)
	b.sendMessage(chatID, text)
}

func (b *Bot) handleSetCity(chatID int64, args string) {
	city := strings.TrimSpace(args)
	if city == "" {
		b.sendMessage(chatID, "Usage: /setcity CityName\nExample: /setcity Tbilisi")
		return
	}

	b.mu.Lock()
	b.config.ChatID = chatID
	b.config.Filters.City = city
	b.mu.Unlock()
	b.saveConfig()

	b.sendMessage(chatID, fmt.Sprintf("City set to %s", city))
}

func (b *Bot) handleSetSqm(chatID int64, args string) {
	parts := strings.Fields(args)
	if len(parts) != 2 {
		b.sendMessage(chatID, "Usage: /setsqm MIN MAX\nExample: /setsqm 800 1200")
		return
	}

	from, err1 := strconv.Atoi(parts[0])
	to, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || from < 0 || to < 0 {
		b.sendMessage(chatID, "Invalid values. Use numbers, e.g. /setsqm 800 1200")
		return
	}

	b.mu.Lock()
	b.config.ChatID = chatID
	b.config.Filters.PricePerSqmFrom = from
	b.config.Filters.PricePerSqmTo = to
	b.mu.Unlock()
	b.saveConfig()

	b.sendMessage(chatID, fmt.Sprintf("Price per m² filter: %d–%d $/m²", from, to))
}

func (b *Bot) handleSetArea(chatID int64, args string) {
	parts := strings.Fields(args)
	if len(parts) != 2 {
		b.sendMessage(chatID, "Usage: /setarea MIN MAX\nExample: /setarea 40 80")
		return
	}

	from, err1 := strconv.Atoi(parts[0])
	to, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || from < 0 || to < 0 {
		b.sendMessage(chatID, "Invalid values. Use numbers, e.g. /setarea 40 80")
		return
	}

	b.mu.Lock()
	b.config.ChatID = chatID
	b.config.Filters.AreaFrom = from
	b.config.Filters.AreaTo = to
	b.mu.Unlock()
	b.saveConfig()

	b.sendMessage(chatID, fmt.Sprintf("Area filter: %d–%d m²", from, to))
}

func (b *Bot) handleFilters(chatID int64) {
	b.mu.Lock()
	f := b.config.Filters
	b.mu.Unlock()

	var lines []string
	lines = append(lines, "*Current Filters:*")
	lines = append(lines, fmt.Sprintf("City: %s", f.City))

	if f.PricePerSqmFrom > 0 || f.PricePerSqmTo > 0 {
		lines = append(lines, fmt.Sprintf("Price per m²: %d–%d $", f.PricePerSqmFrom, f.PricePerSqmTo))
	} else {
		lines = append(lines, "Price per m²: not set")
	}

	if f.AreaFrom > 0 || f.AreaTo > 0 {
		lines = append(lines, fmt.Sprintf("Area: %d–%d m²", f.AreaFrom, f.AreaTo))
	} else {
		lines = append(lines, "Area: not set")
	}

	b.sendMessage(chatID, strings.Join(lines, "\n"))
}

func (b *Bot) handleCheck(chatID int64) {
	b.sendMessage(chatID, "Checking for new listings...")

	newProps := b.checkNewListings()

	if len(newProps) == 0 {
		b.sendMessage(chatID, "No new properties found.")
		return
	}

	text := fmt.Sprintf("*Found %d new properties:*\n\n", len(newProps))
	text += formatProperties(newProps)
	b.sendMessage(chatID, text)
}

func (b *Bot) handleHelp(chatID int64) {
	text := `*Home.ss.ge Property Bot Help*

*Commands:*
/start - Start the bot
/latest - Show latest 10 properties
/latest N - Show latest N properties
/setcity CityName - Set city filter
/setsqm MIN MAX - Set price per m² range
/setarea MIN MAX - Set area range
/filters - Show current filters
/check - Check for new listings
/help - Show this help message

Properties are scraped from home.ss.ge in real-time.`

	b.sendMessage(chatID, text)
}

// checkNewListings scrapes with saved filters, diffs against seen IDs,
// updates seen IDs, and returns only the new properties.
func (b *Bot) checkNewListings() []scraper.Property {
	b.mu.Lock()
	filters := b.config.Filters
	b.mu.Unlock()
	filters.Limit = 20

	properties, err := b.scraper.Scrape(filters)
	if err != nil {
		log.Printf("Error during scheduled scrape: %v", err)
		return nil
	}

	// Build set of seen IDs
	b.mu.Lock()
	seenSet := make(map[string]bool, len(b.config.SeenIDs))
	for _, id := range b.config.SeenIDs {
		seenSet[id] = true
	}
	b.mu.Unlock()

	var newProps []scraper.Property
	var newIDs []string
	for _, p := range properties {
		if !seenSet[p.ID] {
			newProps = append(newProps, p)
			newIDs = append(newIDs, p.ID)
		}
	}

	// Update seen IDs — keep last 500 to avoid unbounded growth
	if len(newIDs) > 0 {
		b.mu.Lock()
		b.config.SeenIDs = append(b.config.SeenIDs, newIDs...)
		if len(b.config.SeenIDs) > 500 {
			b.config.SeenIDs = b.config.SeenIDs[len(b.config.SeenIDs)-500:]
		}
		b.mu.Unlock()
		b.saveConfig()
	}

	return newProps
}

// loadConfig reads the JSON config from disk.
func (b *Bot) loadConfig() error {
	data, err := os.ReadFile(b.configPath)
	if err != nil {
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	var cfg UserConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	b.config = cfg
	// Ensure defaults for type/property type
	if b.config.Filters.Type == "" {
		b.config.Filters.Type = "For sale"
	}
	if b.config.Filters.PropertyType == "" {
		b.config.Filters.PropertyType = "Flat"
	}
	return nil
}

// saveConfig writes the current config to disk.
func (b *Bot) saveConfig() {
	b.mu.Lock()
	data, err := json.MarshalIndent(b.config, "", "  ")
	b.mu.Unlock()
	if err != nil {
		log.Printf("Failed to marshal config: %v", err)
		return
	}

	if err := os.MkdirAll(filepath.Dir(b.configPath), 0755); err != nil {
		log.Printf("Failed to create config directory: %v", err)
		return
	}

	if err := os.WriteFile(b.configPath, data, 0644); err != nil {
		log.Printf("Failed to save config: %v", err)
	}
}

// formatProperties formats property list for display.
func formatProperties(properties []scraper.Property) string {
	var builder strings.Builder

	for i, prop := range properties {
		builder.WriteString(fmt.Sprintf("*%d. %s*\n", i+1, prop.Title))
		builder.WriteString(fmt.Sprintf("Price: %s\n", prop.FullPrice))
		if prop.PricePerSqm != "" {
			builder.WriteString(fmt.Sprintf("Per m2: %s\n", prop.PricePerSqm))
		}
		builder.WriteString(fmt.Sprintf("Location: %s\n", prop.Location))
		if prop.SquareMeters > 0 {
			builder.WriteString(fmt.Sprintf("Size: %d m2\n", prop.SquareMeters))
		}
		builder.WriteString(fmt.Sprintf("[View Listing](%s)\n\n", prop.URL))
	}

	return builder.String()
}

// sendMessage sends a text message to a chat.
func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	if _, err := b.api.Send(msg); err != nil {
		log.Printf("Error sending message to chat %d: %v", chatID, err)
	}
}
