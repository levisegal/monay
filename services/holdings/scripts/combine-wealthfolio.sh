#!/bin/bash
# Combines all E*TRADE transaction CSVs for an account into a single wealthfolio import file
# Usage: ./combine-wealthfolio.sh <prefix>
# Example: ./combine-wealthfolio.sh etrade_3758_maya
#          ./combine-wealthfolio.sh etrade_2060_joint_jtwros

set -euo pipefail

PREFIX="${1:?Usage: $0 <prefix> (e.g. etrade_3758_maya)}"
TESTDATA_DIR="$(dirname "$0")/../importer/testdata"
OUTPUT="$TESTDATA_DIR/${PREFIX}_wealthfolio_import.csv"

# Header
echo "TransactionDate,TransactionType,SecurityType,Symbol,Quantity,Amount,Price,Commission,Description" > "$OUTPUT"

# Process all transaction files (skip wealthfolio files to avoid dupes)
for file in "$TESTDATA_DIR"/${PREFIX}_transactions_*.csv; do
    [[ "$file" == *_wealthfolio.csv ]] && continue
    
    # Skip header lines, account lines, empty lines, then map transaction types
    tail -n +2 "$file" | grep -v "^For Account:" | grep -v "^TransactionDate," | grep -v "^$" | \
    awk -F',' 'BEGIN {OFS=","} {
        # Convert MM/DD/YY to YYYY-MM-DD
        split($1, d, "/")
        year = (d[3] + 0 < 50) ? "20" d[3] : "19" d[3]
        $1 = year "-" sprintf("%02d", d[1]) "-" sprintf("%02d", d[2])
        
        type = $2
        amount = $6
        desc = $9
        
        # Clean up symbols
        symbol = $4
        gsub(/^ +| +$/, "", symbol)  # trim whitespace
        if (symbol ~ /^#[0-9]+$/) symbol = ""  # internal account refs like #2145605
        if (symbol ~ /^[0-9]+[A-Z][0-9]+$/) symbol = ""  # CUSIPs like 74374N102
        $4 = symbol
        
        # Map E*TRADE types to Wealthfolio types
        if (type == "Bought") type = "BUY"
        else if (type == "Sold") type = "SELL"
        else if (type == "Dividend") type = "DIVIDEND"
        else if (type == "Qualified Dividend") type = "DIVIDEND"
        else if (type == "Interest" || type == "Interest Income") type = "INTEREST"
        else if (type == "Transfer") {
            if (amount + 0 > 0) type = "TRANSFER_IN"
            else type = "TRANSFER_OUT"
        }
        else if (type == "Reorganization") {
            # Cash in lieu of fractions (0 qty, positive amount) = DIVIDEND
            # Shares added (positive qty) = ADD_HOLDING
            # Shares removed (negative qty) = REMOVE_HOLDING
            # Fees (negative amount) = FEE
            qty = $5
            if (qty + 0 > 0) type = "ADD_HOLDING"
            else if (qty + 0 < 0) type = "REMOVE_HOLDING"
            else if (amount + 0 > 0) type = "DIVIDEND"
            else if (amount + 0 < 0) type = "FEE"
        }
        else if (type == "Misc Trade") type = "FEE"
        else if (type == "Adjustment") type = "ADD_HOLDING"
        
        $2 = type
        
        # Skip rows that require a symbol but dont have one
        if ((type == "BUY" || type == "SELL" || type == "ADD_HOLDING" || type == "REMOVE_HOLDING" || type == "DIVIDEND") && symbol == "") next
        
        print
    }' >> "$OUTPUT"
done

# Remove trailing empty lines
perl -i -pe 'chomp if eof' "$OUTPUT"

echo "Created: $OUTPUT"
wc -l "$OUTPUT"
