#!/bin/bash
# Combines brokerage transaction CSVs into a single Wealthfolio import file
# Usage: ./combine.sh <broker> <account>
# Example: ./combine.sh etrade maya-3758
#          ./combine.sh lpl bond-5516

set -euo pipefail

BROKER="${1:?Usage: $0 <broker> <account> (e.g. etrade maya-3758)}"
ACCOUNT="${2:?Usage: $0 <broker> <account> (e.g. etrade maya-3758)}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TRANSFORM="$SCRIPT_DIR/transforms/${BROKER}.awk"
ACCOUNT_DIR="$SCRIPT_DIR/../../importer/testdata/$BROKER/$ACCOUNT"
OUTPUT="$ACCOUNT_DIR/wealthfolio_import.csv"

if [[ ! -d "$ACCOUNT_DIR" ]]; then
    echo "Error: Directory not found: $ACCOUNT_DIR"
    exit 1
fi

if [[ ! -f "$TRANSFORM" ]]; then
    echo "Error: Transform not found: $TRANSFORM"
    echo "Available transforms:"
    ls "$SCRIPT_DIR/transforms/"
    exit 1
fi

# Header
echo "TransactionDate,TransactionType,SecurityType,Symbol,Quantity,Amount,Price,Commission,Description" > "$OUTPUT"

# Process all transaction files
for file in "$ACCOUNT_DIR"/transactions_*.csv; do
    [[ ! -f "$file" ]] && continue
    
    # Skip header lines, account lines, empty lines, then apply transform
    tail -n +2 "$file" | \
        grep -v "^For Account:" | \
        grep -v "^TransactionDate," | \
        grep -v "^Date,Activity" | \
        grep -v "^$" | \
        gawk -f "$TRANSFORM" >> "$OUTPUT"
done

# Remove trailing empty lines
perl -i -pe 'chomp if eof' "$OUTPUT"

echo "Created: $OUTPUT"
wc -l "$OUTPUT"
