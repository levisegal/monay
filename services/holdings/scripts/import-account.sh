#!/bin/bash
#
# Import transactions for an account
#
# Usage:
#   ./scripts/import-account.sh <broker> <account-name> <testdata-path>
#
# Examples:
#   ./scripts/import-account.sh etrade "Maya" importer/testdata/etrade/maya-3758
#   ./scripts/import-account.sh lpl "Equities 4015" importer/testdata/lpl/equities-4015
#

set -e

BROKER="$1"
ACCOUNT_NAME="$2"
DATA_PATH="$3"

if [ -z "$BROKER" ] || [ -z "$ACCOUNT_NAME" ] || [ -z "$DATA_PATH" ]; then
    echo "Usage: $0 <broker> <account-name> <testdata-path>"
    echo ""
    echo "Examples:"
    echo "  $0 etrade \"Maya\" importer/testdata/etrade/maya-3758"
    echo "  $0 lpl \"Equities 4015\" importer/testdata/lpl/equities-4015"
    exit 1
fi

if [ ! -d "$DATA_PATH" ]; then
    echo "Error: Directory not found: $DATA_PATH"
    exit 1
fi

# Build file arguments
FILES=$(find "$DATA_PATH" -name "transactions_*.csv" | sort | while read f; do
    echo "--file $f"
done | tr '\n' ' ')

if [ -z "$FILES" ]; then
    echo "Error: No transactions_*.csv files found in $DATA_PATH"
    exit 1
fi

echo "=== Importing $ACCOUNT_NAME from $BROKER ==="
echo ""

# Clear existing data
echo "Clearing existing data..."
go run cmd/main.go lots clear --account-name "$ACCOUNT_NAME" 2>/dev/null || true

# Import all transaction files
echo "Importing transactions..."
eval "go run cmd/main.go import --broker $BROKER --account-name \"$ACCOUNT_NAME\" $FILES"

# Process lots
echo ""
echo "Processing lots..."
go run cmd/main.go lots process --account-name "$ACCOUNT_NAME"

# Show holdings
echo ""
go run cmd/main.go holdings list --account-name "$ACCOUNT_NAME"

# Check for gaps
echo ""
echo "Checking for gaps..."
go run cmd/main.go lots check --account-name "$ACCOUNT_NAME"

