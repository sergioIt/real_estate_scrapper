.PHONY: help build test run docker-build docker-up docker-down docker-logs qa-up qa-down prod-up prod-down clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

build: ## Build the Go application
	go build -o web_parser ./cmd/web/main.go

test: ## Run tests
	go test ./... -v

run: ## Run the scraper application locally
	go run ./cmd/web/main.go

run-simple: ## Run simple scraper (Stage 1: 50 records to stdout)
	go run ./cmd/simple/main.go

run-bot: ## Run the telegram bot locally
	go run ./cmd/bot/main.go

docker-build: ## Build Docker image
	docker build -t web_parser:latest .

docker-up: ## Start development environment with Docker Compose
	docker-compose up -d

docker-down: ## Stop development environment
	docker-compose down

docker-logs: ## Show Docker logs
	docker-compose logs -f

docker-restart: ## Restart Docker containers
	docker-compose restart

qa-up: ## Start QA environment
	docker-compose -f docker-compose.qa.yml up -d

qa-down: ## Stop QA environment
	docker-compose -f docker-compose.qa.yml down

qa-logs: ## Show QA environment logs
	docker-compose -f docker-compose.qa.yml logs -f

prod-up: ## Start production environment
	docker-compose -f docker-compose.prod.yml up -d

prod-down: ## Stop production environment
	docker-compose -f docker-compose.prod.yml down

prod-logs: ## Show production environment logs
	docker-compose -f docker-compose.prod.yml logs -f

clean: ## Clean build artifacts and Docker volumes
	rm -f web_parser
	docker-compose down -v
	docker-compose -f docker-compose.qa.yml down -v
	docker-compose -f docker-compose.prod.yml down -v

db-backup: ## Create database backup (production)
	docker exec web_parser_db_prod pg_dump -U ${DB_USER} ${DB_NAME} > backups/backup_$(shell date +%Y%m%d_%H%M%S).sql

db-restore: ## Restore database from backup (requires BACKUP_FILE variable)
	docker exec -i web_parser_db_prod psql -U ${DB_USER} ${DB_NAME} < $(BACKUP_FILE)
