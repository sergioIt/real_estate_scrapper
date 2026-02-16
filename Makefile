.PHONY: help build test run run-bot clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

build: ## Build the scraper binary
	go build -o web_parser ./cmd/simple/main.go

test: ## Run tests
	go test ./... -v

run: ## Run the scraper (use N=5 to fetch 5 records, default 10)
	go run ./cmd/simple/main.go $(if $(N),-n $(N),)

run-bot: ## Run the telegram bot
	go run ./cmd/bot/main.go

clean: ## Clean build artifacts
	rm -f web_parser
