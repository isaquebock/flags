export PATH := /usr/local/go/bin:/opt/homebrew/bin:$(CURDIR)/.tools/redis/bin:$(PATH)

PID_DIR := .run/pids
LOG_DIR := .run/logs

.PHONY: dev stop logs install check-deps \
        dev-redis dev-api dev-function dev-console \
        stop-redis stop-api stop-function \
        test lint help

# ── Help ──────────────────────────────────────────────────────────────────────

help:
	@echo ""
	@echo "Feature Flags — monorepo targets"
	@echo ""
	@echo "  make dev         Start Redis + API + Function + Console (all-in-one)"
	@echo "  make stop        Kill all background services"
	@echo "  make logs        Tail logs from all background services"
	@echo "  make install     Install all project dependencies"
	@echo "  make check-deps  Verify required tools are installed"
	@echo ""
	@echo "Individual services:"
	@echo "  make dev-redis       Start Redis (daemonized)"
	@echo "  make dev-api         Start Go API in background  → :8080"
	@echo "  make dev-function    Start Edge Function server   → :3000"
	@echo "  make dev-console     Start Console dev server     → :5173"
	@echo ""
	@echo "Tests:"
	@echo "  make test            Run all test suites"
	@echo "  make test-api        Run Go API tests"
	@echo "  make test-function   Run Edge Function tests"
	@echo ""

# ── Dependency check ──────────────────────────────────────────────────────────

check-deps:
	@command -v go           >/dev/null 2>&1 || { echo "✗ go not found";           exit 1; }
	@command -v redis-server >/dev/null 2>&1 || { echo "✗ redis-server not found"; exit 1; }
	@command -v yarn         >/dev/null 2>&1 || { echo "✗ yarn not found";         exit 1; }
	@command -v node         >/dev/null 2>&1 || { echo "✗ node not found";         exit 1; }
	@echo "✓ All required tools found."

# ── Install ───────────────────────────────────────────────────────────────────

install: install-redis-check install-api install-function install-console

install-redis-check:
	@if ! command -v redis-server >/dev/null 2>&1; then \
		echo "⚠  redis-server not found — run 'make install-redis' to build it locally"; \
	fi

install-redis:
	@echo "→ Building Redis from source into .tools/redis/..."
	@mkdir -p .tools
	@curl -fsSL https://download.redis.io/redis-stable.tar.gz | tar xz -C .tools
	@mv .tools/redis-stable .tools/redis-src
	@$(MAKE) -C .tools/redis-src
	@mkdir -p .tools/redis/bin
	@cp .tools/redis-src/src/redis-server .tools/redis-src/src/redis-cli .tools/redis/bin/
	@rm -rf .tools/redis-src
	@echo "✓ Redis built. Add to PATH:"
	@echo "   export PATH=\"$$PWD/.tools/redis/bin:\$$PATH\""

install-api:
	@if command -v go >/dev/null 2>&1; then \
		echo "→ Go modules..."; \
		cd api && go mod download; \
	else \
		echo "⚠  go not found — skipping API deps (install Go: https://go.dev/dl/)"; \
	fi

install-function:
	@echo "→ Function dependencies..."
	@cd function && npm install

install-console:
	@echo "→ Console dependencies..."
	@cd console && HUSKY=0 yarn install

# ── Dev (all services) ────────────────────────────────────────────────────────

dev: | $(PID_DIR) $(LOG_DIR)
	@$(MAKE) -s dev-redis
	@$(MAKE) -s dev-api
	@$(MAKE) -s dev-function
	@echo ""
	@echo "  API      → http://localhost:8080/healthz"
	@echo "  Function → http://localhost:3000/v1/snapshot"
	@echo "  Console  → http://localhost:5173/feature-flags"
	@echo ""
	@echo "  Logs: make logs   |   Stop background services: make stop"
	@echo ""
	cd console && yarn dev

# ── Individual start targets ──────────────────────────────────────────────────

dev-redis: | $(PID_DIR) $(LOG_DIR)
	@if redis-cli ping >/dev/null 2>&1; then \
		echo "→ Redis already running"; \
	elif command -v redis-server >/dev/null 2>&1; then \
		redis-server \
			--daemonize yes \
			--pidfile $(abspath $(PID_DIR)/redis.pid) \
			--logfile $(abspath $(LOG_DIR)/redis.log) \
			--loglevel notice; \
		echo "→ Redis started"; \
	else \
		echo ""; \
		echo "✗ redis-server not found. Install it with:"; \
		echo ""; \
		echo "   macOS (Homebrew):  brew install redis"; \
		echo "   macOS (no brew):   make install-redis"; \
		echo "   Linux (apt):       sudo apt install redis-server"; \
		echo ""; \
		exit 1; \
	fi

dev-api: | $(PID_DIR) $(LOG_DIR)
	@echo "→ Starting Go API on :8080..."
	@cd api && go run ./cmd/api >$(abspath $(LOG_DIR)/api.log) 2>&1 & echo $$! >$(abspath $(PID_DIR)/api.pid)
	@echo "  log: $(LOG_DIR)/api.log"

dev-function: | $(PID_DIR) $(LOG_DIR)
	@echo "→ Starting Edge Function dev server on :3000..."
	@cd function && npm run dev >$(abspath $(LOG_DIR)/function.log) 2>&1 & echo $$! >$(abspath $(PID_DIR)/function.pid)
	@echo "  log: $(LOG_DIR)/function.log"

dev-console:
	cd console && yarn dev

# ── Stop ──────────────────────────────────────────────────────────────────────

stop: stop-api stop-function stop-redis
	@echo "✓ All background services stopped."

stop-redis:
	@redis-cli shutdown >/dev/null 2>&1 && echo "→ Redis stopped" || true
	@rm -f $(PID_DIR)/redis.pid

stop-api:
	@[ -f $(PID_DIR)/api.pid ] \
		&& kill $$(cat $(PID_DIR)/api.pid) 2>/dev/null \
		&& echo "→ API stopped" || true
	@rm -f $(PID_DIR)/api.pid

stop-function:
	@[ -f $(PID_DIR)/function.pid ] \
		&& kill $$(cat $(PID_DIR)/function.pid) 2>/dev/null \
		&& echo "→ Function stopped" || true
	@rm -f $(PID_DIR)/function.pid

# ── Logs ──────────────────────────────────────────────────────────────────────

logs: | $(LOG_DIR)
	@tail -f $(LOG_DIR)/*.log

# ── Tests ─────────────────────────────────────────────────────────────────────

test: test-api test-function

test-api:
	@echo "→ Go API tests..."
	@cd api && go test ./...

test-function:
	@echo "→ Edge Function tests..."
	@cd function && npm test

# ── Internal ──────────────────────────────────────────────────────────────────

$(PID_DIR) $(LOG_DIR):
	@mkdir -p $@
