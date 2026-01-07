# E*TRADE to Wealthfolio transform
# Input: TransactionDate,TransactionType,SecurityType,Symbol,Quantity,Amount,Price,Commission,Description
# Output: Same format with Wealthfolio types

BEGIN { FS=","; OFS="," }

{
    # Convert MM/DD/YY to YYYY-MM-DD
    split($1, d, "/")
    year = (d[3] + 0 < 50) ? "20" d[3] : "19" d[3]
    $1 = year "-" sprintf("%02d", d[1]) "-" sprintf("%02d", d[2])
    
    type = $2
    amount = $6
    qty = $5
    
    # Clean up symbols
    symbol = $4
    gsub(/^ +| +$/, "", symbol)
    if (symbol ~ /^#[0-9]+$/) symbol = ""        # internal account refs
    if (symbol ~ /^[0-9]+[A-Z][0-9]+$/) symbol = ""  # CUSIPs
    $4 = symbol
    
    # Map types
    if (type == "Bought") type = "BUY"
    else if (type == "Sold") type = "SELL"
    else if (type == "Dividend" || type == "Qualified Dividend") type = "DIVIDEND"
    else if (type == "Interest" || type == "Interest Income") type = "INTEREST"
    else if (type == "Transfer") {
        type = (amount + 0 > 0) ? "TRANSFER_IN" : "TRANSFER_OUT"
    }
    else if (type == "Reorganization") {
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
}

