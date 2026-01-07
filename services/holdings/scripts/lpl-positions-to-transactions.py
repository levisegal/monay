#!/usr/bin/env python3
"""
Convert LPL positions.csv to transactions_opening.csv for import.

Usage:
    python scripts/lpl-positions-to-transactions.py \
        importer/testdata/lpl/bond-5516/positions.csv \
        importer/testdata/lpl/bond-5516/transactions_opening.csv
"""

import csv
import re
import sys

def clean_amount(s):
    """Clean currency string: '$11,369.76 ' -> '11369.76'"""
    s = s.strip().replace('$', '').replace(',', '').replace('-', '')
    return s if s else '0'

def clean_quantity(s):
    """Clean quantity string: '10,000.00 ' -> '10000'"""
    s = s.strip().replace(',', '')
    # Remove decimal portion for bonds (face value)
    if '.' in s:
        s = s.split('.')[0]
    return s if s else '0'

def main():
    if len(sys.argv) < 3:
        print(f"Usage: {sys.argv[0]} <positions.csv> <output.csv> [--exclude SYMBOL,...]")
        sys.exit(1)
    
    input_file = sys.argv[1]
    output_file = sys.argv[2]
    
    # Parse exclusions
    exclude_symbols = set()
    for i, arg in enumerate(sys.argv):
        if arg == '--exclude' and i + 1 < len(sys.argv):
            exclude_symbols = set(sys.argv[i + 1].split(','))
    
    print(f"Excluding symbols with existing transactions: {exclude_symbols}")
    
    # Read positions CSV
    with open(input_file, 'r') as f:
        content = f.read()
    
    # Parse rows (handling multi-line descriptions)
    rows = []
    reader = csv.reader(content.splitlines())
    header = next(reader)
    
    current_row = []
    for line in reader:
        if not line:
            continue
        
        # Check if this is a new row (starts with account number)
        if line[0] and line[0].isdigit() and len(line[0]) == 8:
            if current_row:
                rows.append(current_row)
            current_row = line
        else:
            # Continuation of previous row's description
            if current_row and len(line) > 0:
                # Append to description field
                current_row[4] = current_row[4] + ' ' + ' '.join(line)
    
    if current_row:
        rows.append(current_row)
    
    # Write output
    with open(output_file, 'w') as f:
        f.write("Date,Activity,Symbol,Description,Quantity,Unit Price,Value,Held In,Account Nickname,Account Number\n")
        
        for row in rows:
            if len(row) < 12:
                continue
            
            symbol = row[3].strip()
            description = row[4].strip().replace('\n', ' ')
            quantity = clean_quantity(row[5])
            cost_basis = clean_amount(row[11])
            
            # Skip cash account
            if symbol == '9999227':
                continue
            
            # Skip symbols that have buy transactions in transaction files
            # These will be imported from the actual transactions
            if symbol in exclude_symbols:
                continue
            
            # Skip if no cost basis
            if cost_basis == '0':
                continue
            
            # Use 01/01/19 as opening date (before our transaction history)
            f.write(f"01/01/2019,buy,{symbol},\"{description}\",{quantity},$0.00,-${cost_basis},cash,Bond Portfolio,56005516\n")
    
    print(f"Wrote {len(rows)} positions to {output_file}")

if __name__ == '__main__':
    main()

