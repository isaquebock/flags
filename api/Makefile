.PHONY: run test lint build docker docker-build docker-run clean help

help:
	@echo "Feature Flags API — Go targets:"
	@echo "  make run          — Start the API server"
	@echo "  make test         — Run all tests"
	@echo "  make lint         — Run linter (golangci-lint recommended)"
	@echo "  make build        — Build binary to ./api"
	@echo "  make docker       — Build and run Docker image"
	@echo "  make docker-build — Build Docker image only"
	@echo "  make docker-run   — Run Docker container"
	@echo "  make clean        — Remove build artifacts"

run:
	@echo "Starting API server..."
	go run ./cmd/api

test:
	@echo "Running tests..."
	go test -v ./...

lint:
	@echo "Running linter..."
	golangci-lint run ./...

build:
	@echo "Building binary..."
	CGO_ENABLED=0 go build -o api ./cmd/api

docker-build:
	@echo "Building Docker image..."
	docker build -t flags-api:latest .

docker-run: docker-build
	@echo "Running Docker container..."
	docker run -p 8080:8080 \
		-e REDIS_ADDR=host.docker.internal:6379 \
		-e INTERNAL_TOKEN=change-me \
		-e PORT=8080 \
		flags-api:latest

docker: docker-run

clean:
	@echo "Cleaning build artifacts..."
	rm -f api
	go clean

.PHONY: $(MAKECMDGOALS)
