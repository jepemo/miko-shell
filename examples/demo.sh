#!/bin/bash

# Demo script for miko-shell configuration examples

set -e

echo "🐚 Miko Shell Configuration Examples Demo"
echo "=========================================="
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_color() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Check if miko-shell is available
if ! command -v ./miko-shell &> /dev/null && ! command -v miko-shell &> /dev/null; then
    print_color $RED "Error: miko-shell not found. Please build it first with 'make build'"
    exit 1
fi

# Use local binary if available, otherwise use system binary
MIKO_SHELL="./miko-shell"
if [ ! -f "$MIKO_SHELL" ]; then
    MIKO_SHELL="miko-shell"
fi

echo "Available configuration examples:"
echo

# List all example files
for file in examples/dev-config-*.example.yaml; do
    if [ -f "$file" ]; then
        language=$(basename "$file" .example.yaml | sed 's/dev-config-//')
        print_color $BLUE "  📄 $language: $file"
    fi
done

echo
print_color $YELLOW "Select a language example to try:"
echo "1) Python"
echo "2) JavaScript/Node.js"
echo "3) Go"
echo "4) Rust"
echo "5) Elixir"
echo "6) Elixir/Phoenix"
echo "7) PHP"
echo "8) Ruby"
echo "9) Ruby/Rails"
echo "10) Java"
echo "11) Next.js"
echo "12) Django"
echo "13) Spring Boot"
echo "14) Laravel"
echo "15) Exit"
echo

read -p "Enter your choice (1-15): " choice

case $choice in
    1) EXAMPLE_FILE="examples/dev-config-python.example.yaml"; LANG_NAME="Python" ;;
    2) EXAMPLE_FILE="examples/dev-config-javascript.example.yaml"; LANG_NAME="JavaScript/Node.js" ;;
    3) EXAMPLE_FILE="examples/dev-config-go.example.yaml"; LANG_NAME="Go" ;;
    4) EXAMPLE_FILE="examples/dev-config-rust.example.yaml"; LANG_NAME="Rust" ;;
    5) EXAMPLE_FILE="examples/dev-config-elixir.example.yaml"; LANG_NAME="Elixir" ;;
    6) EXAMPLE_FILE="examples/dev-config-phoenix.example.yaml"; LANG_NAME="Elixir/Phoenix" ;;
    7) EXAMPLE_FILE="examples/dev-config-php.example.yaml"; LANG_NAME="PHP" ;;
    8) EXAMPLE_FILE="examples/dev-config-ruby.example.yaml"; LANG_NAME="Ruby" ;;
    9) EXAMPLE_FILE="examples/dev-config-rails.example.yaml"; LANG_NAME="Ruby/Rails" ;;
    10) EXAMPLE_FILE="examples/dev-config-java.example.yaml"; LANG_NAME="Java" ;;
    11) EXAMPLE_FILE="examples/dev-config-nextjs.example.yaml"; LANG_NAME="Next.js" ;;
    12) EXAMPLE_FILE="examples/dev-config-django.example.yaml"; LANG_NAME="Django" ;;
    13) EXAMPLE_FILE="examples/dev-config-spring-boot.example.yaml"; LANG_NAME="Spring Boot" ;;
    14) EXAMPLE_FILE="examples/dev-config-laravel.example.yaml"; LANG_NAME="Laravel" ;;
    15) print_color $YELLOW "Goodbye!"; exit 0 ;;
    *) print_color $RED "Invalid choice. Exiting."; exit 1 ;;
esac

echo
print_color $GREEN "Selected: $LANG_NAME"
print_color $BLUE "Configuration file: $EXAMPLE_FILE"
echo

# Create a temporary directory for the demo
TEMP_DIR="/tmp/miko-shell-demo-$(date +%s)"
mkdir -p "$TEMP_DIR"

print_color $YELLOW "Creating demo environment in: $TEMP_DIR"
cp "$EXAMPLE_FILE" "$TEMP_DIR/dev-config.yaml"

cd "$TEMP_DIR"

echo
print_color $GREEN "Configuration content:"
print_color $BLUE "======================"
cat dev-config.yaml
echo
print_color $BLUE "======================"
echo

print_color $YELLOW "Now you can try the following commands:"
echo "  cd $TEMP_DIR"
echo "  $MIKO_SHELL build      # Build the container image"
echo "  $MIKO_SHELL run test   # Run the test script"
echo "  $MIKO_SHELL shell      # Open interactive shell"
echo

read -p "Do you want to build the container image now? (y/n): " build_choice

if [[ $build_choice =~ ^[Yy]$ ]]; then
    print_color $GREEN "Building container image..."
    $MIKO_SHELL build
    
    if [ $? -eq 0 ]; then
        print_color $GREEN "✅ Build successful!"
        echo
        print_color $YELLOW "You can now run:"
        echo "  $MIKO_SHELL run test"
        echo "  $MIKO_SHELL shell"
    else
        print_color $RED "❌ Build failed!"
    fi
else
    print_color $YELLOW "Skipping build. You can build later with: $MIKO_SHELL build"
fi

echo
print_color $GREEN "Demo complete! The temporary demo environment is at: $TEMP_DIR"
print_color $BLUE "To clean up later, run: rm -rf $TEMP_DIR"
