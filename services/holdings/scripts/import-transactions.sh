#!/usr/bin/env bash
# Import all transaction CSVs from testdata in chronological order
# Usage: ./scripts/import-transactions.sh [--process-lots]
#
# Expects files named: {broker}_{account_number}_{account_name}_transactions_{year}.csv
# Examples:
#   etrade_3758_maya_transactions_2021.csv
#   schwab_1234_roth_ira_transactions_2023.csv

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_DIR="$(dirname "$SCRIPT_DIR")"
TESTDATA_DIR="$SERVICE_DIR/importer/testdata"

PROCESS_LOTS=false
if [[ "$1" == "--process-lots" ]]; then
    PROCESS_LOTS=true
fi

cd "$SERVICE_DIR"

# Get unique broker_acctnum_name combinations
keys=$(ls "$TESTDATA_DIR"/*_*_*_transactions_*.csv 2>/dev/null | \
    xargs -n1 basename | \
    sed -E 's/_transactions_[0-9]+\.csv$//' | \
    sort -u)

if [[ -z "$keys" ]]; then
    echo "No transaction files found in $TESTDATA_DIR"
    exit 1
fi

echo "Found accounts:"
for key in $keys; do
    broker=$(echo "$key" | cut -d'_' -f1)
    acct_name=$(echo "$key" | sed -E 's/^[^_]+_[0-9]+_//' | tr '_' ' ')
    echo "  - $broker: $acct_name"
done
echo ""

# Import each account's transactions in chronological order
for key in $keys; do
    broker=$(echo "$key" | cut -d'_' -f1)
    acct_name=$(echo "$key" | sed -E 's/^[^_]+_[0-9]+_//' | tr '_' ' ')
    # Title case
    acct_display=$(echo "$acct_name" | awk '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) tolower(substr($i,2))}1')
    
    echo "=== Importing account: $acct_display ($broker) ==="
    
    # Find all files for this account and sort by year
    files=$(ls "$TESTDATA_DIR"/"${key}"_transactions_*.csv 2>/dev/null | sort -t_ -k5 -n)
    
    for file in $files; do
        echo "  Importing: $(basename "$file")"
        go run cmd/main.go import --broker "$broker" --file "$file" --account-name "$acct_display"
    done
    
    if [[ "$PROCESS_LOTS" == "true" ]]; then
        echo "  Processing lots for: $acct_display"
        go run cmd/main.go lots process --account-name "$acct_display"
    fi
    
    echo ""
done

echo "=== Import complete ==="

if [[ "$PROCESS_LOTS" == "false" ]]; then
    echo ""
    echo "To process tax lots, run:"
    for key in $keys; do
        acct_name=$(echo "$key" | sed -E 's/^[^_]+_[0-9]+_//' | tr '_' ' ')
        acct_display=$(echo "$acct_name" | awk '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) tolower(substr($i,2))}1')
        echo "  go run cmd/main.go lots process --account-name \"$acct_display\""
    done
fi
