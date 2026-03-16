# LPL Financial — Broker Memory

## Navigation
- Login: `https://accountview.lpl.com/web/login`
- Overview: `/web/overview` — all accounts listed with values
- Positions: `/web/positions` — grouped "By Account", collapsed by default
- Activity: `/web/activity` — transactions with date range filter

## Accounts
- Bond Portfolio (56005516): ~20 CA muni bonds, $282K
- BROKER-NR (3530): empty, skip
- Equities (25274015): 12 positions (AAPL, AIQ, GOOG, IXUS, MU, NVDA, PAVE, QQQJ, SOXX, TSLA, XAR, XLY), $522K

## Known Issues
- "Prev Years" and "Custom" date range sub-menus don't work via playwright accessibility tree
- Date range resets to "1 Month" on page navigation
- CSV always named `Activity.csv` — move immediately
- CSV Value column has tab prefix before dollar sign
- Bond CUSIPs need external mapping to tickers for import
