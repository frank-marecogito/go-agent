#!/bin/bash
# Infrastructure Validator Startup Script
# Run this at the start of every development session

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Infrastructure Validator - Session Startup                  ║${NC}"
echo -e "${BLUE}║  $(date '+%Y-%m-%d %H:%M:%S')                                           ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
echo

# Parse arguments
ONCE=false
MONITOR=false
VERBOSE=false
EXPECTED_DIM=768

while [[ $# -gt 0 ]]; do
    case $1 in
        --once|-o)
            ONCE=true
            shift
            ;;
        --monitor|-m)
            MONITOR=true
            shift
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --dim|-d)
            EXPECTED_DIM="$2"
            shift 2
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --once, -o       Run validation once and exit"
            echo "  --monitor, -m    Run with monitoring loop (10 min intervals)"
            echo "  --verbose, -v    Enable verbose output"
            echo "  --dim, -d N      Expected vector dimension (default: 768)"
            echo "  --help, -h       Show this help message"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ Go is not installed${NC}"
    exit 1
fi

# Build the validator if needed
echo -e "${BLUE}Building validator...${NC}"
cd "$PROJECT_ROOT"
go build -o "$SCRIPT_DIR/validator" ./cmd/validator/ 2>/dev/null || {
    echo -e "${YELLOW}! Build failed, running with go run${NC}"
    BINARY="go run ./cmd/validator/"
}
cd - > /dev/null

# Set up flags
FLAGS="-expected-dim $EXPECTED_DIM"
if $ONCE; then
    FLAGS="$FLAGS -once"
fi
if $VERBOSE; then
    FLAGS="$FLAGS -v"
fi
if $MONITOR; then
    FLAGS="$FLAGS -interval 10m"
fi

# Run the validator
echo
if [ -n "$BINARY" ]; then
    cd "$PROJECT_ROOT"
    $BINARY $FLAGS
else
    "$SCRIPT_DIR/validator" $FLAGS
fi

# Capture exit code
EXIT_CODE=$?

echo
if [ $EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✓ Infrastructure validation passed${NC}"
else
    echo -e "${RED}✗ Infrastructure validation failed${NC}"
fi

exit $EXIT_CODE