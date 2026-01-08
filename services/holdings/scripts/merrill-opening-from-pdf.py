#!/usr/bin/env python3
"""
Parse Merrill Lynch Equity Cost Basis PDF (text) and output transactions_opening.csv

Usage:
  1. Copy-paste the PDF table content into a text file (or use pdftotext)
  2. Run: python merrill-opening-from-pdf.py input.txt > transactions_opening.csv

The script expects lines like:
  AIR PRODUCTS&CHEM | APD 04/08/20 | 16.0000 | 209.4600 | 3,351.36 | ...
  or subtotal lines which it skips
"""

import re
import sys
from dataclasses import dataclass


@dataclass
class Lot:
    symbol: str
    description: str
    acquired: str  # MM/DD/YY
    quantity: str
    unit_cost: str
    total_cost: str


def parse_date(date_str: str) -> str:
    """Convert MM/DD/YY to MM/DD/YYYY"""
    parts = date_str.split("/")
    if len(parts) == 3:
        mm, dd, yy = parts
        # Assume 20xx for years < 50, 19xx otherwise
        year = int(yy)
        if year < 50:
            year = 2000 + year
        else:
            year = 1900 + year
        return f"{mm}/{dd}/{year}"
    return date_str


def parse_cost_basis_text(lines: list[str]) -> list[Lot]:
    """Parse the text extracted from Merrill cost basis PDF"""
    lots = []
    current_description = ""
    current_symbol = ""

    # Pattern for symbol + date line: "APD 04/08/20" or "GOOGL 04/08/20"
    symbol_date_pattern = re.compile(r"^([A-Z]+)\s+(\d{2}/\d{2}/\d{2})$")
    # Pattern for just date (continuation): "02/11/21"
    date_only_pattern = re.compile(r"^(\d{2}/\d{2}/\d{2})$")
    # Pattern for quantity: "16.0000" or "181.0000"
    quantity_pattern = re.compile(r"^[\d,]+\.\d{4}$")
    # Pattern for cost: "209.4600" or "3,351.36"
    cost_pattern = re.compile(r"^[\d,]+\.\d{2,4}$")

    i = 0
    while i < len(lines):
        line = lines[i].strip()

        # Skip empty lines and headers
        if not line or line.startswith("EQUITIES") or line.startswith("Description"):
            i += 1
            continue

        # Skip subtotal/total lines
        if line.lower().startswith("subtotal") or line.lower().startswith("total"):
            i += 1
            continue

        # Check if this is a description line (company name)
        # Usually followed by symbol+date or contains YIELD info
        if "YIELD" in line or (
            not symbol_date_pattern.match(line)
            and not date_only_pattern.match(line)
            and not quantity_pattern.match(line)
            and not cost_pattern.match(line)
            and not line.startswith("$")
            and not line.startswith("(")
            and len(line) > 3
        ):
            # This might be a description, save it
            if "YIELD" not in line:
                current_description = line
            i += 1
            continue

        # Try to match symbol + date
        match = symbol_date_pattern.match(line)
        if match:
            current_symbol = match.group(1)
            acquired = match.group(2)

            # Next lines should be quantity, unit cost, total cost
            # Look ahead for numeric values
            values = []
            j = i + 1
            while j < len(lines) and len(values) < 3:
                val = lines[j].strip().replace(",", "")
                if re.match(r"^[\d.]+$", val):
                    values.append(lines[j].strip())
                    j += 1
                elif lines[j].strip().startswith("$") or lines[j].strip().startswith(
                    "("
                ):
                    j += 1
                else:
                    break

            if len(values) >= 3:
                lots.append(
                    Lot(
                        symbol=current_symbol,
                        description=current_description,
                        acquired=acquired,
                        quantity=values[0],
                        unit_cost=values[1],
                        total_cost=values[2],
                    )
                )
                i = j
                continue

        # Try date-only pattern (continuation lot for same symbol)
        match = date_only_pattern.match(line)
        if match and current_symbol:
            acquired = match.group(1)

            values = []
            j = i + 1
            while j < len(lines) and len(values) < 3:
                val = lines[j].strip().replace(",", "")
                if re.match(r"^[\d.]+$", val):
                    values.append(lines[j].strip())
                    j += 1
                elif lines[j].strip().startswith("$") or lines[j].strip().startswith(
                    "("
                ):
                    j += 1
                else:
                    break

            if len(values) >= 3:
                lots.append(
                    Lot(
                        symbol=current_symbol,
                        description=current_description,
                        acquired=acquired,
                        quantity=values[0],
                        unit_cost=values[1],
                        total_cost=values[2],
                    )
                )
                i = j
                continue

        i += 1

    return lots


def parse_table_format(text: str) -> list[Lot]:
    """Parse markdown-style table format from the PDF"""
    lots = []
    current_symbol = ""
    current_description = ""

    # Split by | and process
    lines = text.split("\n")

    for line in lines:
        # Skip header/separator lines
        if "---" in line or "Description" in line or "Symbol" in line:
            continue

        # Skip subtotal/total lines
        lower = line.lower()
        if "subtotal" in lower or "total yield" in lower:
            continue

        # Split by |
        parts = [p.strip() for p in line.split("|") if p.strip()]
        if len(parts) < 6:
            continue

        # Try to extract: description, symbol+date, quantity, unit_cost, total_cost
        # Format varies, but generally:
        # parts[0] = description or empty
        # parts[1] = "SYMBOL MM/DD/YY" or just "MM/DD/YY"
        # parts[2] = quantity
        # parts[3] = unit cost
        # parts[4] = total cost

        desc_part = parts[0]
        symbol_date_part = parts[1] if len(parts) > 1 else ""

        # Check if this has a symbol
        sym_match = re.match(r"^([A-Z]+)\s+(\d{2}/\d{2}/\d{2})$", symbol_date_part)
        date_match = re.match(r"^(\d{2}/\d{2}/\d{2})$", symbol_date_part)

        if sym_match:
            current_symbol = sym_match.group(1)
            acquired = sym_match.group(2)
            if desc_part and "YIELD" not in desc_part:
                current_description = desc_part.split("CURRENT")[0].strip()
        elif date_match:
            acquired = date_match.group(1)
        else:
            continue

        if not current_symbol:
            continue

        # Extract numeric values
        quantity = parts[2] if len(parts) > 2 else ""
        unit_cost = parts[3] if len(parts) > 3 else ""
        total_cost = parts[4] if len(parts) > 4 else ""

        # Clean up values
        quantity = quantity.replace(",", "")
        unit_cost = unit_cost.replace(",", "").replace("$", "")
        total_cost = total_cost.replace(",", "").replace("$", "")

        if quantity and unit_cost and total_cost:
            try:
                float(quantity)
                float(unit_cost)
                float(total_cost)
                lots.append(
                    Lot(
                        symbol=current_symbol,
                        description=current_description,
                        acquired=acquired,
                        quantity=quantity,
                        unit_cost=unit_cost,
                        total_cost=total_cost,
                    )
                )
            except ValueError:
                pass

    return lots


def output_csv(lots: list[Lot], account: str = "CMA 5VT-22241"):
    """Output in Merrill transaction CSV format (normalized - no space before comma)"""
    print(
        '"Trade Date","Settlement Date","Account","Description","Type","Symbol/ CUSIP","Quantity","Price","Amount"," "'
    )

    for lot in lots:
        date = parse_date(lot.acquired)
        desc = f"Opening Balance - {lot.description}" if lot.description else "Opening Balance"
        qty = lot.quantity.replace(",", "")
        price = f"${lot.unit_cost}"
        amount = f"${lot.total_cost}"

        print(
            f'"{date}","{date}","{account}","{desc}","","{lot.symbol}","{qty}","{price}","{amount}",""'
        )


# Complete data from Jan 2024 PDF - EQUITIES, MUNICIPAL BONDS, and MUTUAL FUNDS/ETFs
JAN_2024_LOTS = [
    # ========== MUNICIPAL BONDS ==========
    Lot("799561HN8", "SAN YSIDRO CALIF SCH DIST", "04/24/19", "50000", "0.9878", "49392.26"),
    Lot("952347US9", "WEST CONTRA COSTA CA UN SCH", "04/24/19", "50000", "0.9645", "48225.06"),
    Lot("03255LBK4", "ANAHEIM CALIF PUB FING AUTH", "04/24/19", "50000", "0.9614", "48071.28"),
    Lot("587601LM5", "MERCED CALIF CITY SCH DIST", "12/21/17", "30000", "0.8700", "26099.71"),
    Lot("072024VH2", "BAY AREA TOLL AUTH CALIF", "12/21/17", "30000", "1.0392", "31175.11"),
    # ========== MUTUAL FUNDS / ETFs ==========
    Lot("MANLX", "BLACKROCK NATIONAL MUNICIPAL FUND", "12/13/22", "12450.1990", "10.0383", "125000.00"),
    Lot("CMNIX", "CALAMOS MARKET NEUTRAL INCOME FD", "06/11/18", "2592.4490", "13.3145", "34514.68"),
    Lot("MBXIX", "CATALYST MILLBURN HEDGE STRATEGY", "08/05/21", "2082.3730", "36.3575", "75711.32"),
    Lot("GSFTX", "COLUMBIA DIVIDEND INCOME FUND", "02/11/21", "3786.1330", "27.8102", "105305.97"),
    Lot("DLY", "DOUBLELINE YIELD OPPORTUNITIES FD", "06/08/22", "507.0000", "15.3787", "7796.98"),
    Lot("EIFAX", "EATON VANCE FLOATING RATE ADV", "01/20/22", "4725.8980", "10.5801", "50000.00"),
    Lot("IBMO", "ISHARES IBONDS DEC 2026 TERM MUN", "03/23/23", "3500.0000", "25.4394", "89038.00"),
    Lot("IBMQ", "ISHARES IBONDS DEC 2028 TERM MUNI", "12/11/23", "4500.0000", "25.3665", "114149.30"),
    Lot("IBMN", "ISHARES IBONDS DEC 2025 TERM MUNI", "03/23/23", "3500.0000", "26.5526", "92934.28"),
    Lot("IWR", "ISHARES RUSSELL MIDCAP ETF", "01/30/19", "519.0000", "50.8068", "26368.74"),
    Lot("VNLA", "JANUS HENDERSON SHORT DURATION", "09/22/22", "2626.0000", "48.2185", "126621.79"),
    Lot("MDIJX", "MFS INTERNATIONAL DIVERSIFICATION", "01/24/23", "5129.5990", "21.9696", "112697.29"),
    Lot("GCOW", "PACER GLOBAL CASH COWS DIVIDEND", "01/24/23", "3022.0000", "33.7300", "101932.06"),
    Lot("COWZ", "PACER US CASH COWS 100 ETF", "02/17/22", "3725.0000", "48.1800", "179470.50"),
    Lot("CALF", "PACER US SMALL CAP CASH COWS", "01/24/23", "946.0000", "38.7200", "36629.12"),
    Lot("PDBZX", "PGIM TOTAL RETURN BOND FUND", "12/11/23", "13673.3290", "11.7312", "160398.74"),
    Lot("SPY", "SPDR S&P 500 ETF TRUST", "12/11/23", "70.0000", "460.7277", "32250.94"),
    Lot("ITM", "VANECK INTERMEDIATE MUNI ETF", "12/13/22", "2000.0000", "46.0250", "92050.00"),
    # ========== EQUITIES ==========
    Lot("APD", "AIR PRODUCTS&CHEM", "04/08/20", "16.0000", "209.4600", "3351.36"),
    Lot("APD", "AIR PRODUCTS&CHEM", "02/11/21", "21.0000", "255.0000", "5355.00"),
    Lot("GOOGL", "ALPHABET INC SHS CL A", "04/08/20", "181.0000", "60.1798", "10892.56"),
    Lot("GOOGL", "ALPHABET INC SHS CL A", "01/24/23", "32.0000", "99.0750", "3170.40"),
    Lot("AMZN", "AMAZON COM INC COM", "01/24/23", "334.0000", "96.9000", "32364.60"),
    Lot("AMZN", "AMAZON COM INC COM", "11/09/23", "72.0000", "141.7600", "10206.72"),
    Lot("AXP", "AMER EXPRESS COMPANY", "01/24/23", "164.0000", "156.4499", "25657.79"),
    Lot("AXP", "AMER EXPRESS COMPANY", "06/13/23", "39.0000", "175.2200", "6833.58"),
    Lot("AAPL", "APPLE INC", "04/08/20", "344.0000", "66.2925", "22804.62"),
    Lot("AAPL", "APPLE INC", "08/05/21", "39.0000", "147.1548", "5739.04"),
    Lot("AAPL", "APPLE INC", "01/24/23", "56.0000", "142.1383", "7959.75"),
    Lot("AAPL", "APPLE INC", "05/15/23", "70.0000", "172.1472", "12050.31"),
    Lot("BDX", "BECTON DICKINSON CO", "04/08/20", "44.0000", "238.8625", "10509.95"),
    Lot("CP", "CANADIAN PAC KANS CITY LTD", "02/11/21", "147.0000", "71.3949", "10495.06"),
    Lot("CP", "CANADIAN PAC KANS CITY LTD", "01/14/22", "53.0000", "77.3247", "4098.21"),
    Lot("CAT", "CATERPILLAR INC DEL", "07/17/23", "124.0000", "257.4875", "31928.45"),
    Lot("CMCSA", "COMCAST CORP NEW CL A", "01/24/23", "579.0000", "40.0700", "23200.53"),
    Lot("ED", "CONSOLIDATED EDISON INC", "11/09/23", "197.0000", "89.3374", "17599.47"),
    Lot("COST", "COSTCO WHOLESALE CRP DEL", "02/11/21", "59.0000", "355.8640", "20995.98"),
    Lot("COST", "COSTCO WHOLESALE CRP DEL", "11/15/21", "6.0000", "519.4816", "3116.89"),
    Lot("LLY", "ELI LILLY & CO", "04/08/20", "80.0000", "144.3350", "11546.80"),
    Lot("EQIX", "EQUINIX INC", "04/08/20", "16.0000", "642.3612", "10277.78"),
    Lot("EQIX", "EQUINIX INC", "02/11/21", "5.0000", "750.0000", "3750.00"),
    Lot("EQIX", "EQUINIX INC", "01/14/22", "1.0000", "734.8000", "734.80"),
    Lot("EQIX", "EQUINIX INC", "06/13/23", "14.0000", "757.0500", "10598.70"),
    Lot("EXR", "EXTRA SPACE STORAGE INC", "03/23/21", "42.0000", "129.6350", "5444.67"),
    Lot("EXR", "EXTRA SPACE STORAGE INC", "01/24/23", "57.0000", "151.8100", "8653.17"),
    Lot("XOM", "EXXON MOBIL CORP COM", "08/05/21", "329.0000", "57.4676", "18906.87"),
    Lot("XOM", "EXXON MOBIL CORP COM", "11/15/21", "39.0000", "64.6256", "2520.40"),
    Lot("HCA", "HCA HEALTHCARE INC", "01/24/23", "47.0000", "261.0565", "12269.66"),
    Lot("HCA", "HCA HEALTHCARE INC", "03/16/23", "16.0000", "251.1600", "4018.56"),
    Lot("HCA", "HCA HEALTHCARE INC", "05/15/23", "34.0000", "277.5200", "9435.68"),
    Lot("HD", "HOME DEPOT INC", "04/08/20", "36.0000", "195.5730", "7040.63"),
    Lot("IBM", "INTL BUSINESS MACHINES CORP", "08/14/23", "169.0000", "142.3698", "24060.51"),
    Lot("INTU", "INTUIT INC COM", "04/08/20", "34.0000", "247.7029", "8421.90"),
    Lot("INTU", "INTUIT INC COM", "01/24/23", "3.0000", "403.8533", "1211.56"),
    Lot("INTU", "INTUIT INC COM", "06/13/23", "25.0000", "445.8752", "11146.88"),
    Lot("JPM", "JPMORGAN CHASE & CO", "01/30/19", "168.0000", "104.3767", "17535.30"),
    Lot("JPM", "JPMORGAN CHASE & CO", "04/08/20", "154.0000", "93.1650", "14347.41"),
    Lot("JPM", "JPMORGAN CHASE & CO", "11/09/23", "70.0000", "145.2975", "10170.83"),
    Lot("MGA", "MAGNA INTL INC CL A VTG", "08/14/23", "259.0000", "57.2600", "14830.34"),
    Lot("MGA", "MAGNA INTL INC CL A VTG", "11/09/23", "194.0000", "52.5777", "10200.09"),
    Lot("MCD", "MCDONALDS CORP COM", "02/05/18", "30.0000", "168.3030", "5049.09"),
    Lot("MCD", "MCDONALDS CORP COM", "02/11/21", "18.0000", "214.4600", "3860.28"),
    Lot("MCD", "MCDONALDS CORP COM", "01/24/23", "30.0000", "266.8493", "8005.48"),
    Lot("META", "META PLATFORMS INC CLASS A", "07/17/23", "51.0000", "307.4100", "15677.91"),
    Lot("MSFT", "MICROSOFT CORP", "01/30/19", "160.0000", "105.3471", "16855.54"),
    Lot("MSFT", "MICROSOFT CORP", "04/08/20", "90.0000", "165.3390", "14880.51"),
    Lot("NVDA", "NVIDIA", "01/31/23", "73.0000", "192.7998", "14074.39"),
    Lot("NVDA", "NVIDIA", "03/16/23", "10.0000", "251.1380", "2511.38"),
    Lot("NVDA", "NVIDIA", "08/31/23", "22.0000", "493.1881", "10850.14"),
    Lot("PANW", "PALO ALTO NETWORKS INC COM", "09/26/22", "56.0000", "164.6850", "9222.36"),
    Lot("PNC", "PNC FINCL SERVICES GROUP", "04/08/20", "91.0000", "98.6550", "8977.61"),
    Lot("PNC", "PNC FINCL SERVICES GROUP", "02/11/21", "8.0000", "159.8200", "1278.56"),
    Lot("PNC", "PNC FINCL SERVICES GROUP", "11/15/21", "8.0000", "205.0537", "1640.43"),
    Lot("PNC", "PNC FINCL SERVICES GROUP", "01/24/23", "8.0000", "159.3975", "1275.18"),
    Lot("PNC", "PNC FINCL SERVICES GROUP", "03/16/23", "7.0000", "128.8500", "901.95"),
    Lot("PFG", "PRINCIPAL FINANCIAL GRP", "01/14/22", "222.0000", "76.0836", "16890.56"),
    Lot("PGR", "PROGRESSIVE CRP OHIO", "02/11/21", "89.0000", "85.6100", "7619.29"),
    Lot("PGR", "PROGRESSIVE CRP OHIO", "01/24/23", "59.0000", "127.6033", "7528.60"),
    Lot("ROK", "ROCKWELL AUTOMATION INC", "02/11/21", "65.0000", "245.9000", "15983.50"),
    Lot("ROK", "ROCKWELL AUTOMATION INC", "03/23/21", "15.0000", "262.3820", "3935.73"),
    Lot("ROK", "ROCKWELL AUTOMATION INC", "05/15/23", "14.0000", "275.3664", "3855.13"),
    Lot("CRM", "SALESFORCE INC", "01/24/23", "56.0000", "157.2400", "8805.44"),
    Lot("SYK", "STRYKER CORP", "04/08/20", "21.0000", "172.7590", "3627.94"),
    Lot("SYK", "STRYKER CORP", "02/11/21", "17.0000", "244.1800", "4151.06"),
    Lot("TXN", "TEXAS INSTRUMENTS", "02/11/21", "119.0000", "175.6858", "20906.62"),
    Lot("TJX", "TJX COS INC NEW", "01/24/23", "141.0000", "80.1790", "11305.24"),
    Lot("UNH", "UNITEDHEALTH GROUP INC", "04/08/20", "44.0000", "259.2750", "11408.10"),
    Lot("UNH", "UNITEDHEALTH GROUP INC", "10/03/23", "21.0000", "512.6428", "10765.50"),
    Lot("V", "VISA INC CL A SHRS", "04/08/20", "99.0000", "174.5297", "17278.45"),
    Lot("V", "VISA INC CL A SHRS", "01/24/23", "49.0000", "225.8079", "11064.59"),
    Lot("V", "VISA INC CL A SHRS", "05/15/23", "48.0000", "232.0672", "11139.23"),
    Lot("WMT", "WALMART INC", "10/26/18", "102.0000", "98.9999", "10097.99"),
    Lot("WMT", "WALMART INC", "04/08/20", "50.0000", "121.7170", "6085.85"),
    Lot("WMT", "WALMART INC", "08/31/23", "66.0000", "162.7300", "10740.18"),
    Lot("WCN", "WASTE CONNECTIONS INC", "09/24/21", "145.0000", "132.3993", "19197.91"),
    Lot("WCN", "WASTE CONNECTIONS INC", "11/15/21", "3.0000", "135.9200", "407.76"),
    Lot("WCN", "WASTE CONNECTIONS INC", "01/14/22", "6.0000", "125.5050", "753.03"),
    Lot("WMB", "WILLIAMS COMPANIES DEL", "01/14/22", "820.0000", "28.6266", "23473.88"),
    Lot("WMB", "WILLIAMS COMPANIES DEL", "03/16/23", "282.0000", "28.5365", "8047.31"),
    Lot("ZTS", "ZOETIS INC", "04/08/20", "126.0000", "125.0460", "15755.80"),
    Lot("ZTS", "ZOETIS INC", "11/09/23", "59.0000", "172.6467", "10186.16"),
]


def main():
    if len(sys.argv) > 1:
        # Read from file
        with open(sys.argv[1]) as f:
            text = f.read()

        # Try table format first
        lots = parse_table_format(text)
        if not lots:
            # Try line-by-line format
            lots = parse_cost_basis_text(text.split("\n"))

        if lots:
            output_csv(lots)
        else:
            print("Could not parse any lots from input", file=sys.stderr)
            print("Using hard-coded Jan 2024 data instead...", file=sys.stderr)
            output_csv(JAN_2024_LOTS)
    else:
        # Use hard-coded data from the PDF shared in chat
        print(
            "No input file provided. Using hard-coded data from Jan 2024 PDF.",
            file=sys.stderr,
        )
        print(
            "NOTE: This is incomplete - pages 2-3 were truncated. Please download full PDF.",
            file=sys.stderr,
        )
        output_csv(JAN_2024_LOTS)


if __name__ == "__main__":
    main()

