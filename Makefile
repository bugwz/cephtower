BACKEND_DIR := backend
FRONTEND_DIR := frontend
BIN_DIR := bin
APP_NAME := cephtower
CONFIG ?= ../config/config.yaml
FRONTEND_PORT ?= 36901
BACKEND_STATIC_DIR := $(BACKEND_DIR)/internal/frontend/dist

.PHONY: build build-backend build-frontend run run-backend run-frontend backend-dev backend-test frontend-dev frontend-build generate-ceph-client

build: build-frontend build-backend

build-backend:
	mkdir -p $(BIN_DIR)
	cd $(BACKEND_DIR) && go build -o ../$(BIN_DIR)/$(APP_NAME) ./cmd

build-frontend:
	cd $(FRONTEND_DIR) && npm run build
	rm -rf $(BACKEND_STATIC_DIR)
	mkdir -p $(BACKEND_STATIC_DIR)
	cp -R $(FRONTEND_DIR)/dist/. $(BACKEND_STATIC_DIR)/

run:
	@trap 'kill 0' INT TERM EXIT; \
	(cd $(BACKEND_DIR) && go run ./cmd -config $(CONFIG)) & \
	(cd $(FRONTEND_DIR) && npm run dev -- --host 0.0.0.0 --port $(FRONTEND_PORT)) & \
	wait

run-backend:
	cd $(BACKEND_DIR) && go run ./cmd -config $(CONFIG)

run-frontend:
	cd $(FRONTEND_DIR) && npm run dev -- --host 0.0.0.0 --port $(FRONTEND_PORT)

backend-dev: run-backend

backend-test:
	cd $(BACKEND_DIR) && go test ./...

generate-ceph-client:
	ruby tools/generate_ceph_dashboard_client.rb

frontend-dev: run-frontend

frontend-build: build-frontend
