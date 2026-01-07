# LPL Financial to Wealthfolio transform
# Input: Date,Activity,Symbol,Description,Quantity,Unit Price,Value,Held In,Account Nickname,Account Number
# Output: TransactionDate,TransactionType,SecurityType,Symbol,Quantity,Amount,Price,Commission,Description
#
# Note: Uses FPAT to handle quoted CSV fields with commas

BEGIN { 
    FPAT = "([^,]*)|(\"[^\"]*\")"
    OFS=","
}

{
    # Parse input fields
    date = $1
    activity = tolower($2)
    symbol = $3
    desc = $4
    gsub(/^"|"$/, "", desc)  # Remove surrounding quotes
    qty = $5
    price = $6
    value = $7
    
    # Convert MM/DD/YYYY to YYYY-MM-DD
    split(date, d, "/")
    date = d[3] "-" sprintf("%02d", d[1]) "-" sprintf("%02d", d[2])
    
    # Clean value (remove $, tabs, handle negative)
    gsub(/[\t$"]/, "", value)
    amount = value + 0
    
    # Clean price
    gsub(/[\t$"]/, "", price)
    if (price == "-") price = 0
    
    # Clean quantity
    if (qty == "-") qty = 0
    
    # Clean symbol - skip internal refs
    if (symbol == "9999227" || symbol == "CASH") symbol = ""
    
    # Map activity types
    if (activity == "buy") type = "BUY"
    else if (activity == "sell") type = "SELL"
    else if (activity == "interest") type = "INTEREST"
    else if (activity == "reinvest interest") type = "BUY"
    else if (activity == "ica transfer") {
        type = (amount > 0) ? "TRANSFER_IN" : "TRANSFER_OUT"
    }
    else if (activity == "journal") type = "TRANSFER_OUT"
    else if (activity == "fee") type = "FEE"
    else type = ""
    
    # Skip unmapped types
    if (type == "") next
    
    # Skip rows that require a symbol but dont have one
    if ((type == "BUY" || type == "SELL") && symbol == "") next
    
    # Output in Wealthfolio format
    sec_type = "BOND"
    
    commission = 0
    if (amount < 0) amount = -amount  # Make positive for output
    
    # Escape description for CSV output
    gsub(/"/, "\"\"", desc)
    
    print date, type, sec_type, symbol, qty, amount, price, commission, "\"" desc "\""
}
