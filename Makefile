# FCCUR Makefile
.PHONY: build test clean install run dev deploy help migrate

# Variables
BINARY_NAME=fccur
MIGRATE_BINARY=migrate
BUILD_DIR=./bin
INSTALL_DIR=/usr/local/bin
SERVICE_FILE=deploy/fccur.service
SYSTEMD_DIR=/etc/systemd/system
DB_PATH=/var/lib/fccur/fccur.db
PACKAGES_DIR=/var/lib/fccur/packages
WEB_DIR=/var/lib/fccur/web
MIGRATIONS_DIR=./migrations
GO=go
GOFLAGS=-v

# Build binary
build:
	@echo "Building FCCUR..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build migration tool
build-migrate:
	@echo "Building migration tool..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(MIGRATE_BINARY) ./cmd/migrate
	@echo "Build complete: $(BUILD_DIR)/$(MIGRATE_BINARY)"

# Build all binaries
build-all: build build-migrate

# Build for Raspberry Pi (ARM64)
build-pi:
	@echo "Building FCCUR for Raspberry Pi (ARM64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-arm64 ./cmd/server
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)-arm64"

# Build for Raspberry Pi (ARM)
build-pi-arm:
	@echo "Building FCCUR for Raspberry Pi (ARM)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm GOARM=7 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-armv7 ./cmd/server
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)-armv7"

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	$(GO) test -v ./tests/...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)
	rm -f fccur.db fccur.db-shm fccur.db-wal
	@echo "Clean complete"

# Run in development mode
dev: build
	@echo "Running in development mode..."
	$(BUILD_DIR)/$(BINARY_NAME) -addr :8080 -db ./fccur.db -packages ./packages

# Run server
run: build
	@echo "Starting FCCUR server..."
	$(BUILD_DIR)/$(BINARY_NAME)

# Install binary to system
install: build
	@echo "Installing FCCUR to $(INSTALL_DIR)..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installation complete"

# Deploy to system (install + systemd service)
deploy: build install
	@echo "Deploying FCCUR..."
	@# Create directories
	sudo mkdir -p /var/lib/fccur/packages
	sudo mkdir -p /var/log/fccur
	@# Copy web files
	sudo cp -r web $(WEB_DIR)
	@# Install systemd service
	sudo cp $(SERVICE_FILE) $(SYSTEMD_DIR)/
	sudo systemctl daemon-reload
	@echo "Deployment complete. Enable and start with:"
	@echo "  sudo systemctl enable fccur"
	@echo "  sudo systemctl start fccur"

# Uninstall from system
uninstall:
	@echo "Uninstalling FCCUR..."
	sudo systemctl stop fccur 2>/dev/null || true
	sudo systemctl disable fccur 2>/dev/null || true
	sudo rm -f $(SYSTEMD_DIR)/fccur.service
	sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	sudo systemctl daemon-reload
	@echo "Uninstall complete. Data remains in /var/lib/fccur"

# Setup Raspberry Pi (run on target device)
setup-pi:
	@echo "Setting up FCCUR on Raspberry Pi..."
	@chmod +x deploy/setup-pi.sh
	@sudo deploy/setup-pi.sh

# Show systemd service status
status:
	sudo systemctl status fccur

# View logs
logs:
	sudo journalctl -u fccur -f

# Restart service
restart:
	sudo systemctl restart fccur

# Database Migration Commands
migrate-up: build-migrate
	@echo "Running migrations..."
	$(BUILD_DIR)/$(MIGRATE_BINARY) -db ./fccur.db -command=up

migrate-down: build-migrate
	@echo "Rolling back last migration..."
	$(BUILD_DIR)/$(MIGRATE_BINARY) -db ./fccur.db -command=down

migrate-version: build-migrate
	@echo "Checking migration version..."
	$(BUILD_DIR)/$(MIGRATE_BINARY) -db ./fccur.db -command=version

migrate-goto: build-migrate
	@echo "Migrating to version $(VERSION)..."
	@if [ -z "$(VERSION)" ]; then echo "Error: VERSION not specified. Use: make migrate-goto VERSION=2"; exit 1; fi
	$(BUILD_DIR)/$(MIGRATE_BINARY) -db ./fccur.db -command=goto -version=$(VERSION)

migrate-steps: build-migrate
	@echo "Running $(STEPS) migration steps..."
	@if [ -z "$(STEPS)" ]; then echo "Error: STEPS not specified. Use: make migrate-steps STEPS=2 or STEPS=-2"; exit 1; fi
	$(BUILD_DIR)/$(MIGRATE_BINARY) -db ./fccur.db -command=steps -steps=$(STEPS)

migrate-force: build-migrate
	@echo "⚠ WARNING: Forcing migration version to $(VERSION)"
	@if [ -z "$(VERSION)" ]; then echo "Error: VERSION not specified. Use: make migrate-force VERSION=2"; exit 1; fi
	$(BUILD_DIR)/$(MIGRATE_BINARY) -db ./fccur.db -command=force -force-version=$(VERSION)

migrate-validate: build-migrate
	@echo "Validating migrations..."
	$(BUILD_DIR)/$(MIGRATE_BINARY) -db ./fccur.db -command=validate

migrate-reset: build-migrate
	@echo "⚠ WARNING: This will reset the database!"
	@read -p "Type 'yes' to continue: " confirm && [ "$$confirm" = "yes" ]
	$(BUILD_DIR)/$(MIGRATE_BINARY) -db ./fccur.db -command=drop
	$(BUILD_DIR)/$(MIGRATE_BINARY) -db ./fccur.db -command=up

# Help
help:
	@echo "FCCUR Makefile Commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  make build              - Build the binary"
	@echo "  make build-migrate      - Build migration tool"
	@echo "  make build-all          - Build all binaries"
	@echo "  make build-pi           - Build for Raspberry Pi (ARM64)"
	@echo "  make build-pi-arm       - Build for Raspberry Pi (ARM v7)"
	@echo ""
	@echo "Test Commands:"
	@echo "  make test               - Run unit tests"
	@echo "  make test-integration   - Run integration tests"
	@echo ""
	@echo "Development Commands:"
	@echo "  make clean              - Remove build artifacts"
	@echo "  make dev                - Run in development mode"
	@echo "  make run                - Build and run server"
	@echo ""
	@echo "Deployment Commands:"
	@echo "  make install            - Install binary to system"
	@echo "  make deploy             - Full deployment (install + systemd)"
	@echo "  make uninstall          - Remove from system"
	@echo "  make setup-pi           - Setup on Raspberry Pi"
	@echo "  make status             - Show service status"
	@echo "  make logs               - View service logs"
	@echo "  make restart            - Restart service"
	@echo ""
	@echo "Migration Commands:"
	@echo "  make migrate-up         - Apply all pending migrations"
	@echo "  make migrate-down       - Rollback last migration"
	@echo "  make migrate-version    - Show current migration version"
	@echo "  make migrate-goto       - Migrate to specific version (VERSION=n)"
	@echo "  make migrate-steps      - Run n steps (STEPS=n or STEPS=-n)"
	@echo "  make migrate-force      - Force version (VERSION=n)"
	@echo "  make migrate-validate   - Validate migrations"
	@echo "  make migrate-reset      - Reset database (DANGEROUS!)"
	@echo ""
	@echo "  make help               - Show this help"
