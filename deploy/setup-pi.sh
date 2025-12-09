#!/bin/bash

# FCCUR Raspberry Pi Setup Script
# This script sets up FCCUR on a Raspberry Pi

set -e

echo "======================================"
echo "FCCUR Raspberry Pi Setup"
echo "======================================"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root (use sudo)"
    exit 1
fi

# Detect architecture
ARCH=$(uname -m)
echo "Detected architecture: $ARCH"

# Check if running on Raspberry Pi
if [ ! -f /proc/device-tree/model ]; then
    echo "Warning: This doesn't appear to be a Raspberry Pi"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Create fccur user if it doesn't exist
if ! id -u fccur >/dev/null 2>&1; then
    echo "Creating fccur user..."
    useradd -r -s /bin/false -d /var/lib/fccur fccur
else
    echo "User fccur already exists"
fi

# Create necessary directories
echo "Creating directories..."
mkdir -p /var/lib/fccur/packages
mkdir -p /var/lib/fccur/web
mkdir -p /var/log/fccur

# Set ownership
chown -R fccur:fccur /var/lib/fccur
chown -R fccur:fccur /var/log/fccur

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Installing Go..."

    # Determine Go architecture
    if [ "$ARCH" = "aarch64" ]; then
        GO_ARCH="arm64"
    elif [ "$ARCH" = "armv7l" ]; then
        GO_ARCH="armv6l"
    else
        echo "Unsupported architecture: $ARCH"
        exit 1
    fi

    # Download and install Go
    GO_VERSION="1.21.5"
    cd /tmp
    wget "https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
    rm -rf /usr/local/go
    tar -C /usr/local -xzf "go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"
    rm "go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"

    # Add Go to PATH
    if ! grep -q "/usr/local/go/bin" /etc/profile; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    fi
    export PATH=$PATH:/usr/local/go/bin

    echo "Go installed successfully"
else
    echo "Go is already installed: $(go version)"
fi

# Install SQLite if not present
if ! command -v sqlite3 &> /dev/null; then
    echo "Installing SQLite3..."
    apt-get update
    apt-get install -y sqlite3 libsqlite3-dev
else
    echo "SQLite3 is already installed"
fi

# Install GCC if not present (needed for CGO)
if ! command -v gcc &> /dev/null; then
    echo "Installing GCC..."
    apt-get update
    apt-get install -y build-essential
else
    echo "GCC is already installed"
fi

# Copy web files
if [ -d "web" ]; then
    echo "Copying web files..."
    cp -r web/* /var/lib/fccur/web/
    chown -R fccur:fccur /var/lib/fccur/web
else
    echo "Warning: web directory not found"
fi

# Build the binary
if [ -f "cmd/server/main.go" ]; then
    echo "Building FCCUR..."
    go build -v -o /usr/local/bin/fccur ./cmd/server
    chmod +x /usr/local/bin/fccur
    echo "Binary installed to /usr/local/bin/fccur"
else
    echo "Error: Source files not found. Are you in the FCCUR directory?"
    exit 1
fi

# Install systemd service
if [ -f "deploy/fccur.service" ]; then
    echo "Installing systemd service..."
    cp deploy/fccur.service /etc/systemd/system/
    systemctl daemon-reload
    echo "Systemd service installed"
else
    echo "Warning: fccur.service not found"
fi

# Configure firewall if ufw is installed
if command -v ufw &> /dev/null; then
    echo "Configuring firewall..."
    ufw allow 8080/tcp comment "FCCUR HTTP"
    echo "Firewall configured (port 8080 opened)"
fi

# Enable and start service
echo ""
read -p "Enable and start FCCUR service now? (Y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Nn]$ ]]; then
    systemctl enable fccur
    systemctl start fccur

    echo ""
    echo "Service started. Checking status..."
    sleep 2
    systemctl status fccur --no-pager
fi

# Print summary
echo ""
echo "======================================"
echo "Setup Complete!"
echo "======================================"
echo ""
echo "FCCUR has been installed successfully."
echo ""
echo "Configuration:"
echo "  - Binary: /usr/local/bin/fccur"
echo "  - Database: /var/lib/fccur/fccur.db"
echo "  - Packages: /var/lib/fccur/packages"
echo "  - Web files: /var/lib/fccur/web"
echo "  - Logs: /var/log/fccur (or use: journalctl -u fccur -f)"
echo ""
echo "Service management:"
echo "  - Start:   sudo systemctl start fccur"
echo "  - Stop:    sudo systemctl stop fccur"
echo "  - Restart: sudo systemctl restart fccur"
echo "  - Status:  sudo systemctl status fccur"
echo "  - Logs:    sudo journalctl -u fccur -f"
echo ""
echo "Access the web interface at:"
echo "  http://$(hostname -I | awk '{print $1}'):8080"
echo ""
echo "Next steps:"
echo "  1. Upload packages via the admin interface"
echo "  2. Configure your network to allow client access"
echo "  3. Consider setting up a static IP for the Pi"
echo ""
