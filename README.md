# Web Parser for home.ss.ge

A Go-based web scraper and Telegram bot for real estate listings from home.ss.ge.

## Features

- Scrapes property listings from home.ss.ge
- Telegram bot with configurable filters (city, price per m², area)
- Daily automatic notifications for new listings
- Duplicate detection — only notifies about properties you haven't seen
- Filter persistence across restarts (JSON file)
- CLI scraper for quick one-off checks

## Project Structure

```
web_parser/
├── cmd/
│   ├── bot/          # Telegram bot
│   └── simple/       # CLI scraper
├── pkg/
│   ├── scraper/      # Web scraping logic
│   └── telegram/     # Bot, commands, scheduler
├── data/             # Created at runtime
│   └── filters.json  # User config & seen listings
├── Makefile
└── .env.example
```

## Quick Start

### Prerequisites

- Go 1.23+
- Telegram bot token (get one from [@BotFather](https://t.me/BotFather))

### Setup

```bash
# Clone and install dependencies
go mod download

# Create .env from example
cp .env.example .env

# Edit .env and add your bot token
```

### Running the Telegram Bot

```bash
make run-bot
```

Then open your bot in Telegram and send `/start`.

### Running the CLI Scraper

```bash
# Fetch latest 10 properties
make run

# Fetch 5 properties
make run N=5
```

## Configuration

Environment variables (set in `.env`):

| Variable | Default | Description |
|----------|---------|-------------|
| `TELEGRAM_BOT_TOKEN` | (required) | Bot token from @BotFather |
| `CHECK_INTERVAL` | `24h` | How often to check for new listings. Go duration format: `1m`, `1h`, `24h` |

For testing, use a short interval:
```bash
CHECK_INTERVAL=1m make run-bot
```

## Bot Commands

| Command | Description |
|---------|-------------|
| `/start` | Register and show welcome message |
| `/latest` | Show latest 10 properties using saved filters |
| `/latest N` | Show latest N properties |
| `/setcity Tbilisi` | Set city filter |
| `/setsqm 800 1200` | Set price per m² range ($/m²) |
| `/setarea 40 80` | Set area range (m²) |
| `/filters` | Show current filter settings |
| `/check` | Manually check for new listings |
| `/help` | Show help message |

### Example Session

```
/start
/setcity Tbilisi
/setsqm 800 1200
/setarea 40 80
/filters
/check
```

The bot will also automatically check for new listings at the configured interval and send you any new properties it finds.

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make run-bot` | Run the Telegram bot |
| `make run` | Run the CLI scraper (default 10 records) |
| `make run N=5` | Run the CLI scraper with N records |
| `make test` | Run tests |
| `make build` | Build the scraper binary |
| `make clean` | Clean build artifacts |

## License

MIT
