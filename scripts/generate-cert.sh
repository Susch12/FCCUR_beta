#!/bin/bash

# Self-signed Certificate Generator for FCCUR
# This script generates a self-signed TLS certificate for development/internal use

set -e

# Configuration
CERT_DIR="${CERT_DIR:-./certs}"
DOMAIN="${DOMAIN:-localhost}"
DAYS="${DAYS:-3650}"  # 10 years
KEY_SIZE=2048

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "======================================"
echo "FCCUR Certificate Generator"
echo "======================================"
echo ""

# Create certs directory
mkdir -p "$CERT_DIR"

# Check if openssl is installed
if ! command -v openssl &> /dev/null; then
    echo "Error: openssl is not installed"
    echo "Install with: sudo apt-get install openssl"
    exit 1
fi

# Certificate details
CERT_FILE="$CERT_DIR/server.crt"
KEY_FILE="$CERT_DIR/server.key"

# Check if certificate already exists
if [ -f "$CERT_FILE" ] && [ -f "$KEY_FILE" ]; then
    echo -e "${YELLOW}Warning: Certificate already exists!${NC}"
    echo "  Certificate: $CERT_FILE"
    echo "  Key: $KEY_FILE"
    echo ""
    read -p "Overwrite existing certificate? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborted."
        exit 0
    fi
fi

echo "Generating self-signed certificate..."
echo "  Domain: $DOMAIN"
echo "  Validity: $DAYS days"
echo "  Key size: $KEY_SIZE bits"
echo ""

# Generate private key and certificate
openssl req -x509 -newkey rsa:$KEY_SIZE -nodes \
    -keyout "$KEY_FILE" \
    -out "$CERT_FILE" \
    -days $DAYS \
    -subj "/C=US/ST=State/L=City/O=FCCUR/CN=$DOMAIN" \
    -addext "subjectAltName=DNS:$DOMAIN,DNS:*.${DOMAIN},IP:127.0.0.1,IP:0.0.0.0" \
    2>/dev/null

# Set proper permissions
chmod 600 "$KEY_FILE"
chmod 644 "$CERT_FILE"

echo -e "${GREEN}Certificate generated successfully!${NC}"
echo ""
echo "Files created:"
echo "  Certificate: $CERT_FILE"
echo "  Private key: $KEY_FILE"
echo ""
echo "To use with FCCUR:"
echo "  ./fccur -addr :8443 -cert $CERT_FILE -key $KEY_FILE"
echo ""
echo "Or with the binary:"
echo "  ./bin/fccur -addr :8443 -cert $CERT_FILE -key $KEY_FILE"
echo ""
echo -e "${YELLOW}Note: This is a self-signed certificate.${NC}"
echo "Your browser will show a security warning. This is normal for self-signed certificates."
echo "For production use, consider using Let's Encrypt certificates."
echo ""

# Show certificate details
echo "Certificate details:"
openssl x509 -in "$CERT_FILE" -noout -subject -dates -ext subjectAltName

echo ""
echo "To trust this certificate on your system:"
echo ""
echo "Linux (Ubuntu/Debian):"
echo "  sudo cp $CERT_FILE /usr/local/share/ca-certificates/fccur.crt"
echo "  sudo update-ca-certificates"
echo ""
echo "macOS:"
echo "  sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $CERT_FILE"
echo ""
echo "Windows:"
echo "  certutil -addstore -f \"ROOT\" $CERT_FILE"
echo ""
