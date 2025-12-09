#!/bin/bash
# Package FCCUR for release (excludes tests and dev files)

VERSION=${1:-"v1.0.0"}
RELEASE_NAME="fccur-${VERSION}"

echo "📦 Packaging FCCUR ${VERSION}..."

# Create release directory
mkdir -p "releases/${RELEASE_NAME}"

# Copy source code
echo "Copying source code..."
cp -r cmd internal "releases/${RELEASE_NAME}/"

# Copy frontend
echo "Copying frontend..."
cp -r web "releases/${RELEASE_NAME}/"

# Copy migrations
echo "Copying migrations..."
cp -r migrations "releases/${RELEASE_NAME}/"

# Copy deployment files
echo "Copying deployment files..."
cp -r deploy "releases/${RELEASE_NAME}/"

# Copy build files
echo "Copying build configuration..."
cp go.mod go.sum Makefile "releases/${RELEASE_NAME}/"

# Copy documentation
echo "Copying documentation..."
cp README.md "releases/${RELEASE_NAME}/"

# Remove test files
echo "Removing test files..."
find "releases/${RELEASE_NAME}" -name "*_test.go" -delete
find "releases/${RELEASE_NAME}" -name "test-*.html" -delete

# Build binaries
echo "Building binaries..."
cd "releases/${RELEASE_NAME}"
make build-all
make build-pi
cd ../..

# Create tarball
echo "Creating release archive..."
cd releases
tar -czf "${RELEASE_NAME}.tar.gz" "${RELEASE_NAME}/"
cd ..

# Create binary-only package
echo "Creating binary-only package..."
mkdir -p "releases/${RELEASE_NAME}-binary"
cp "releases/${RELEASE_NAME}/bin/"* "releases/${RELEASE_NAME}-binary/"
cp -r "releases/${RELEASE_NAME}/web" "releases/${RELEASE_NAME}-binary/"
cp -r "releases/${RELEASE_NAME}/migrations" "releases/${RELEASE_NAME}-binary/"
cp -r "releases/${RELEASE_NAME}/deploy" "releases/${RELEASE_NAME}-binary/"
cp "releases/${RELEASE_NAME}/README.md" "releases/${RELEASE_NAME}-binary/"

cd releases
tar -czf "${RELEASE_NAME}-binary.tar.gz" "${RELEASE_NAME}-binary/"
cd ..

echo "✅ Release packages created:"
echo "   releases/${RELEASE_NAME}.tar.gz (full source)"
echo "   releases/${RELEASE_NAME}-binary.tar.gz (binary only)"
echo ""
echo "📊 Package contents:"
du -sh "releases/${RELEASE_NAME}.tar.gz"
du -sh "releases/${RELEASE_NAME}-binary.tar.gz"
