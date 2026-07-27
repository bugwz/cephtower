BACKEND_DIR := backend
FRONTEND_DIR := frontend
BIN_DIR := bin
RELEASE_DIR := dist
APP_NAME := cephtower
CONFIG ?= app/config/config.yaml
RUN_CONFIG := app/config/config.yaml
CONFIG_TEMPLATE := config/config.yaml
FRONTEND_PORT ?= 36901
BACKEND_STATIC_DIR := $(BACKEND_DIR)/internal/webui/frontend/dist
DEV_BACKEND_BIN := $(BIN_DIR)/$(APP_NAME)-dev
CGO_ENABLED ?= 0
RELEASE_TARGETS ?= darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64
RELEASE_BUILD_TARGETS := $(if $(TARGET),$(TARGET),$(RELEASE_TARGETS))
VERSION ?= $(shell cat VERSION 2>/dev/null || echo 0.0.0)
GIT_SHA ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
GO_MIN_VERSION := 1.26
GO_CACHE ?= $(CURDIR)/.gocache
GO_MODULE_CACHE ?= $(CURDIR)/.gomodcache
NODE_MIN_MAJOR := 20
NPM_MIN_MAJOR := 10

.PHONY: check-env check-backend-env check-frontend-env ensure-frontend-deps ensure-run-config build build-backend build-frontend release run run-backend run-frontend test test-backend test-frontend

check-env: check-backend-env check-frontend-env

check-backend-env:
	@set -e; \
	if ! command -v go >/dev/null 2>&1; then \
		echo "Environment is not ready: Go was not found. Please install Go $(GO_MIN_VERSION)+, then run make again."; \
		exit 1; \
	fi; \
	go_version=$$(go env GOVERSION 2>/dev/null || go version | awk '{print $$3}'); \
	go_version=$${go_version#go}; \
	go_major=$${go_version%%.*}; \
	go_rest=$${go_version#*.}; \
	go_minor=$${go_rest%%.*}; \
	go_min_major=$$(echo "$(GO_MIN_VERSION)" | cut -d. -f1); \
	go_min_minor=$$(echo "$(GO_MIN_VERSION)" | cut -d. -f2); \
	if [ "$$go_major" -lt "$$go_min_major" ] || { [ "$$go_major" -eq "$$go_min_major" ] && [ "$$go_minor" -lt "$$go_min_minor" ]; }; then \
		echo "Environment is not ready: found Go $$go_version, but this project requires Go $(GO_MIN_VERSION)+. Please upgrade Go, then run make again."; \
		exit 1; \
	fi

check-frontend-env:
	@set -e; \
	if ! command -v node >/dev/null 2>&1; then \
		echo "Environment is not ready: Node.js was not found. Please install Node.js $(NODE_MIN_MAJOR)+, then run make again."; \
		exit 1; \
	fi; \
	if ! command -v npm >/dev/null 2>&1; then \
		echo "Environment is not ready: npm was not found. Please install npm $(NPM_MIN_MAJOR)+, then run make again."; \
		exit 1; \
	fi; \
	node_major=$$(node -v | sed 's/^v//' | cut -d. -f1); \
	npm_major=$$(npm -v | cut -d. -f1); \
	if [ "$$node_major" -lt "$(NODE_MIN_MAJOR)" ]; then \
		echo "Environment is not ready: found Node.js $$(node -v), but this project requires Node.js $(NODE_MIN_MAJOR)+. Please upgrade Node.js, then run make again."; \
		exit 1; \
	fi; \
	if [ "$$npm_major" -lt "$(NPM_MIN_MAJOR)" ]; then \
		echo "Environment is not ready: found npm $$(npm -v), but this project requires npm $(NPM_MIN_MAJOR)+. Please upgrade npm, then run make again."; \
		exit 1; \
	fi

ensure-frontend-deps: check-frontend-env
	@if [ ! -d "$(FRONTEND_DIR)/node_modules" ] || [ ! -x "$(FRONTEND_DIR)/node_modules/.bin/vite" ]; then \
		echo "Frontend dependencies are missing or incomplete. Running npm install in $(FRONTEND_DIR)."; \
		cd $(FRONTEND_DIR) && npm install; \
	fi

build: check-env build-frontend build-backend

build-backend: check-backend-env build-frontend
	mkdir -p $(BIN_DIR)
	cd $(BACKEND_DIR) && env CGO_ENABLED=$(CGO_ENABLED) GOCACHE=$(GO_CACHE) GOMODCACHE=$(GO_MODULE_CACHE) go build -o ../$(BIN_DIR)/$(APP_NAME) ./cmd

build-frontend: test-frontend
	rm -rf $(BACKEND_STATIC_DIR)
	mkdir -p $(BACKEND_STATIC_DIR)
	touch $(BACKEND_STATIC_DIR)/.keep
	cp -R $(FRONTEND_DIR)/dist/. $(BACKEND_STATIC_DIR)/

release: check-env build-frontend
	mkdir -p $(RELEASE_DIR)
	@set -e; \
	for target in $(RELEASE_BUILD_TARGETS); do \
		goos=$${target%/*}; \
		goarch=$${target#*/}; \
		suffix=""; \
		if [ "$$goos" = "windows" ]; then suffix=".exe"; fi; \
		output="$(CURDIR)/$(RELEASE_DIR)/$(APP_NAME)-$(VERSION)-$(GIT_SHA)-$$goos-$$goarch$$suffix"; \
		echo "Building $$output"; \
		(cd $(BACKEND_DIR) && env CGO_ENABLED=$(CGO_ENABLED) GOOS=$$goos GOARCH=$$goarch GOCACHE=$(GO_CACHE) GOMODCACHE=$(GO_MODULE_CACHE) go build -o "$$output" ./cmd); \
	done

ensure-run-config:
	@if [ ! -f "$(RUN_CONFIG)" ]; then \
		cd $(BACKEND_DIR) && env CGO_ENABLED=$(CGO_ENABLED) GOCACHE=$(GO_CACHE) GOMODCACHE=$(GO_MODULE_CACHE) go run ./tools/config-init \
			-template ../$(CONFIG_TEMPLATE) -target ../$(RUN_CONFIG) -server-dir ./app; \
		echo "Created $(RUN_CONFIG) from $(CONFIG_TEMPLATE)."; \
	fi

run: ensure-run-config check-env ensure-frontend-deps
	@set -e; \
	mkdir -p $(BIN_DIR); \
	(cd $(BACKEND_DIR) && env CGO_ENABLED=$(CGO_ENABLED) GOCACHE=$(GO_CACHE) GOMODCACHE=$(GO_MODULE_CACHE) go build -o ../$(DEV_BACKEND_BIN) ./cmd); \
	trap 'kill 0' INT TERM EXIT; \
	(./$(DEV_BACKEND_BIN) -config $(CONFIG)) & \
	(cd $(FRONTEND_DIR) && npm run dev -- --host 0.0.0.0 --port $(FRONTEND_PORT) --strictPort) & \
	wait

run-backend: check-backend-env
	mkdir -p $(BIN_DIR)
	cd $(BACKEND_DIR) && env CGO_ENABLED=$(CGO_ENABLED) GOCACHE=$(GO_CACHE) GOMODCACHE=$(GO_MODULE_CACHE) go build -o ../$(DEV_BACKEND_BIN) ./cmd
	./$(DEV_BACKEND_BIN) -config $(CONFIG)

run-frontend: ensure-frontend-deps
	cd $(FRONTEND_DIR) && npm run dev -- --host 0.0.0.0 --port $(FRONTEND_PORT) --strictPort

test: test-backend test-frontend

test-backend: check-backend-env
	cd $(BACKEND_DIR) && env CGO_ENABLED=$(CGO_ENABLED) GOCACHE=$(GO_CACHE) GOMODCACHE=$(GO_MODULE_CACHE) go test ./...
	cd $(BACKEND_DIR) && env CGO_ENABLED=$(CGO_ENABLED) GOCACHE=$(GO_CACHE) GOMODCACHE=$(GO_MODULE_CACHE) go run ./tools/openapi-gen -check -output api/openapi-v1.yaml

test-frontend: ensure-frontend-deps
	cd $(FRONTEND_DIR) && npm run build
