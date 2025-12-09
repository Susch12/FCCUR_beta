#!/bin/bash

# FCCUR Initial Package Loader
# This script pre-loads packages into FCCUR from local directories

set -e

# Configuration
FCCUR_URL="${FCCUR_URL:-http://localhost:8080}"
API_ENDPOINT="${FCCUR_URL}/api/upload"
MATERIAL_DIR="${MATERIAL_DIR:-./Material}"
TOOLS_DIR="${TOOLS_DIR:-./Tools}"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Counters
TOTAL_UPLOADED=0
TOTAL_FAILED=0
TOTAL_SKIPPED=0

# Function to print colored messages
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Function to check if server is reachable
check_server() {
    echo "Checking FCCUR server at $FCCUR_URL..."

    if curl -s -f "${FCCUR_URL}/health" > /dev/null 2>&1; then
        print_success "Server is reachable"
        return 0
    else
        print_error "Server is not reachable at $FCCUR_URL"
        echo "Please ensure FCCUR is running and accessible."
        exit 1
    fi
}

# Function to upload a course material
upload_material() {
    local file="$1"
    local course_name="$2"
    local category="${3:-library}"

    # Extract name from filename (remove extension)
    local basename=$(basename "$file")
    local name="${basename%.*}"

    # Extract year from filename if present (YYYY format)
    local version=$(echo "$name" | grep -oE '(19|20)[0-9]{2}' | head -1)
    if [ -z "$version" ]; then
        version="2025"
    fi

    # Determine platform (always 'all' for materials)
    local platform="all"

    echo "  Uploading: $basename"

    # Upload the file
    response=$(curl -s -w "\n%{http_code}" -X POST "$API_ENDPOINT" \
        -F "content_type=material" \
        -F "course_name=$course_name" \
        -F "name=$name" \
        -F "version=$version" \
        -F "category=$category" \
        -F "platform=$platform" \
        -F "description=Course material for $course_name" \
        -F "package=@$file")

    # Extract status code
    http_code=$(echo "$response" | tail -n1)

    if [ "$http_code" -eq 201 ]; then
        print_success "Uploaded: $name (v$version)"
        ((TOTAL_UPLOADED++))
    else
        print_error "Failed to upload: $basename (HTTP $http_code)"
        ((TOTAL_FAILED++))
    fi
}

# Function to upload a tool/software
upload_tool() {
    local file="$1"
    local name="$2"
    local version="$3"
    local category="$4"
    local platform="${5:-all}"
    local description="${6:-}"

    echo "  Uploading: $(basename "$file")"

    # Upload the file
    response=$(curl -s -w "\n%{http_code}" -X POST "$API_ENDPOINT" \
        -F "content_type=tool" \
        -F "name=$name" \
        -F "version=$version" \
        -F "category=$category" \
        -F "platform=$platform" \
        -F "description=$description" \
        -F "package=@$file")

    # Extract status code
    http_code=$(echo "$response" | tail -n1)

    if [ "$http_code" -eq 201 ]; then
        print_success "Uploaded: $name v$version"
        ((TOTAL_UPLOADED++))
    else
        print_error "Failed to upload: $(basename "$file") (HTTP $http_code)"
        ((TOTAL_FAILED++))
    fi
}

# Function to load materials from a course directory
load_course_materials() {
    local course_dir="$1"
    local course_name=$(basename "$course_dir")

    # Clean up course name (replace underscores with spaces)
    course_name="${course_name//_/ }"

    echo ""
    echo "Loading materials for: $course_name"
    echo "----------------------------------------"

    # Count files
    local file_count=$(find "$course_dir" -type f \( -iname "*.pdf" -o -iname "*.docx" -o -iname "*.doc" -o -iname "*.txt" -o -iname "*.pptx" -o -iname "*.ppt" \) | wc -l)

    if [ "$file_count" -eq 0 ]; then
        print_warning "No materials found in $course_dir"
        ((TOTAL_SKIPPED++))
        return
    fi

    echo "Found $file_count material(s)"

    # Upload each file
    find "$course_dir" -type f \( -iname "*.pdf" -o -iname "*.docx" -o -iname "*.doc" -o -iname "*.txt" -o -iname "*.pptx" -o -iname "*.ppt" \) | while read -r file; do
        upload_material "$file" "$course_name" "library"
    done
}

# Function to load all course materials
load_all_materials() {
    if [ ! -d "$MATERIAL_DIR" ]; then
        print_warning "Material directory not found: $MATERIAL_DIR"
        return
    fi

    echo ""
    echo "========================================"
    echo "Loading Course Materials"
    echo "========================================"

    # Find all course directories
    find "$MATERIAL_DIR" -mindepth 1 -maxdepth 1 -type d | sort | while read -r course_dir; do
        load_course_materials "$course_dir"
    done
}

# Function to load tools from manifest file
load_tools_from_manifest() {
    local manifest="$1"

    if [ ! -f "$manifest" ]; then
        print_warning "Manifest file not found: $manifest"
        return
    fi

    echo ""
    echo "========================================"
    echo "Loading Tools from Manifest"
    echo "========================================"

    # Read manifest line by line (format: file|name|version|category|platform|description)
    while IFS='|' read -r file name version category platform description; do
        # Skip comments and empty lines
        [[ "$file" =~ ^#.*$ ]] && continue
        [[ -z "$file" ]] && continue

        # Check if file exists
        if [ ! -f "$file" ]; then
            print_warning "File not found: $file"
            ((TOTAL_SKIPPED++))
            continue
        fi

        upload_tool "$file" "$name" "$version" "$category" "$platform" "$description"
    done < "$manifest"
}

# Function to scan and load tools automatically
load_tools_auto() {
    if [ ! -d "$TOOLS_DIR" ]; then
        print_warning "Tools directory not found: $TOOLS_DIR"
        return
    fi

    echo ""
    echo "========================================"
    echo "Auto-loading Tools"
    echo "========================================"

    # Common tool file extensions
    find "$TOOLS_DIR" -type f \( -iname "*.exe" -o -iname "*.msi" -o -iname "*.dmg" -o -iname "*.pkg" -o -iname "*.deb" -o -iname "*.rpm" -o -iname "*.tar.gz" -o -iname "*.zip" -o -iname "*.iso" \) | while read -r file; do
        local basename=$(basename "$file")
        local name="${basename%.*}"
        local version="1.0"
        local category="tool"
        local platform="all"

        # Try to detect platform from filename
        if [[ "$basename" =~ windows|win|.exe|.msi ]]; then
            platform="windows"
        elif [[ "$basename" =~ macos|osx|darwin|.dmg|.pkg ]]; then
            platform="mac"
        elif [[ "$basename" =~ linux|.deb|.rpm ]]; then
            platform="linux"
        fi

        # Try to detect category from name
        if [[ "$name" =~ gcc|clang|javac|python|rustc ]]; then
            category="compiler"
        elif [[ "$name" =~ vscode|eclipse|intellij|pycharm|netbeans ]]; then
            category="ide"
        elif [[ "$name" =~ ubuntu|debian|fedora|centos|arch ]]; then
            category="os"
        fi

        upload_tool "$file" "$name" "$version" "$category" "$platform" "Auto-uploaded from $TOOLS_DIR"
    done
}

# Main script
main() {
    echo "========================================"
    echo "FCCUR Initial Package Loader"
    echo "========================================"
    echo ""
    echo "Configuration:"
    echo "  Server URL: $FCCUR_URL"
    echo "  Material Directory: $MATERIAL_DIR"
    echo "  Tools Directory: $TOOLS_DIR"
    echo ""

    # Check server connectivity
    check_server

    # Load materials
    if [ -d "$MATERIAL_DIR" ]; then
        load_all_materials
    else
        print_warning "Skipping materials - directory not found: $MATERIAL_DIR"
    fi

    # Load tools from manifest if exists
    if [ -f "tools-manifest.txt" ]; then
        load_tools_from_manifest "tools-manifest.txt"
    elif [ -d "$TOOLS_DIR" ]; then
        # Auto-load tools
        load_tools_auto
    else
        print_warning "Skipping tools - no manifest or directory found"
    fi

    # Print summary
    echo ""
    echo "========================================"
    echo "Summary"
    echo "========================================"
    echo "  Uploaded: $TOTAL_UPLOADED"
    echo "  Failed: $TOTAL_FAILED"
    echo "  Skipped: $TOTAL_SKIPPED"
    echo ""

    if [ "$TOTAL_FAILED" -gt 0 ]; then
        print_warning "Some uploads failed. Check the output above for details."
        exit 1
    else
        print_success "All packages loaded successfully!"
    fi
}

# Show usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Load initial packages into FCCUR from local directories.

OPTIONS:
    -h, --help          Show this help message
    -u, --url URL       FCCUR server URL (default: http://localhost:8080)
    -m, --materials DIR Material directory (default: ./Material)
    -t, --tools DIR     Tools directory (default: ./Tools)

ENVIRONMENT VARIABLES:
    FCCUR_URL           Server URL
    MATERIAL_DIR        Material directory path
    TOOLS_DIR           Tools directory path

EXAMPLES:
    # Load with default settings
    $0

    # Load with custom server URL
    $0 --url http://192.168.1.100:8080

    # Load only materials
    $0 --materials ./Material --tools /nonexistent

    # Using environment variables
    FCCUR_URL=http://pi.local:8080 MATERIAL_DIR=/mnt/materials $0

MANIFEST FILE:
    Create a file named 'tools-manifest.txt' with the following format:

    file_path|name|version|category|platform|description

    Example:
    /path/to/gcc.tar.gz|GCC Compiler|13.2|compiler|linux|GNU Compiler Collection
    /path/to/vscode.exe|Visual Studio Code|1.85|ide|windows|Code Editor

EOF
    exit 0
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            ;;
        -u|--url)
            FCCUR_URL="$2"
            API_ENDPOINT="${FCCUR_URL}/api/upload"
            shift 2
            ;;
        -m|--materials)
            MATERIAL_DIR="$2"
            shift 2
            ;;
        -t|--tools)
            TOOLS_DIR="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

# Run main function
main
