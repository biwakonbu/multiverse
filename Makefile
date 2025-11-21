.PHONY: help build test test-unit test-integration test-docker test-codex test-all coverage clean lint docker-build run-sample

help:
	@echo "agent-runner development commands:"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build           - Build binary to bin/agent-runner"
	@echo "  make run-sample      - Run with sample_task_go.yaml"
	@echo "  make docker-build    - Build Docker image (agent-runner-codex:latest)"
	@echo ""
	@echo "Testing:"
	@echo "  make test-unit       - Run unit tests only (fast)"
	@echo "  make test-integration- Run mock integration tests (fast)"
	@echo "  make test-docker     - Run Docker sandbox tests (requires Docker)"
	@echo "  make test-codex      - Run Codex integration tests (requires Docker & auth)"
	@echo "  make test-all        - Run all tests with all tags"
	@echo "  make test            - Alias for test-all"
	@echo ""
	@echo "Quality:"
	@echo "  make coverage        - Generate coverage report (coverage.html)"
	@echo "  make lint            - Run golangci-lint"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean           - Remove binaries and test artifacts"
	@echo ""

# Build targets
build:
	go build -o bin/agent-runner ./cmd/agent-runner

# Test targets
test-unit:
	go test -v -race ./pkg/... ./internal/...

test-integration:
	go test -v ./test/integration/...

test-docker:
	go test -v -tags=docker -timeout=10m ./test/sandbox/...

test-codex:
	go test -v -tags=codex -timeout=10m ./test/codex/...

test-all: test-unit test-integration
	go test -v -tags=docker,codex -timeout=15m ./...

test: test-all

# Coverage target
coverage:
	go test -coverprofile=coverage.out -tags=docker,codex ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out | tail -5

# Lint target
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not found, installing..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run --timeout=5m ./...

# Docker targets
docker-build:
	docker build -t agent-runner-codex:latest ./sandbox/

# Run targets
run-sample: build
	./bin/agent-runner < sample_task_go.yaml

# Cleanup target
clean:
	rm -rf bin/ coverage.out coverage.html
	docker ps -a --filter "label=created-by=agent-runner" -q | xargs -r docker rm -f 2>/dev/null || true

# Dev target - run tests with coverage on file changes (requires entr)
watch:
	@which entr > /dev/null || (echo "entr not found. Install with: brew install entr (macOS) or apt-get install entr (Linux)"; exit 1)
	find . -name "*.go" | entr -r make test-unit

.DEFAULT_GOAL := help
