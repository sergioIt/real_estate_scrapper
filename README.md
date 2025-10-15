# Web Parser for home.ss.ge

A simple Go-based web scraper for real estate listings from home.ss.ge

## Features

- 🏠 Scrapes property listings from home.ss.ge
- 📋 Fetches last 10 unique records
- 🔍 Duplicate detection using URL tracking
- 💻 Clean stdout output

## Project Structure

```
web_parser/
├── cmd/
│   └── simple/       # Simple scraper (Stage 1)
├── pkg/
│   └── scraper/      # Web scraping logic
├── Makefile
├── requirements.md
└── go.mod
```

## Quick Start

### Prerequisites

- Go 1.23+

### Running

```bash
# Using Makefile
make run-simple

# Or directly
go run ./cmd/simple/main.go
```

### Output

```
=== Found 10 unique properties ===

1. 2 room Flat For Sale
   Price: 1,811 $
   Location: Didi digomi
   URL: https://home.ss.ge/en/real-estate/2-room-Flat-For-Sale-Didi-digomi-33288661

2. 3 room Flat For Sale
   Price: 1,496 $
   Location: Saburtalo
   URL: https://home.ss.ge/en/real-estate/3-room-Flat-For-Sale-Saburtalo-33336506

...

Total: 10 unique properties
```

## What It Scrapes

For each property:
- **Title**: Room count and property type (e.g., "3 room Flat For Sale")
- **Price**: In USD ($) or Georgian Lari (₾)
- **Location**: Neighborhood/district name
- **URL**: Direct link to the listing

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make run-simple` | Run the simple scraper (10 records) |
| `make test` | Run tests |
| `make build` | Build Go binary |

## Development Stages

- [x] Stage 1: Fetch last 10 unique records to stdout
- [ ] Stage 2: Add parameter for N records

See `requirements.md` for detailed roadmap.

## License

MIT
