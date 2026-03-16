import asyncio
import re
from concurrent.futures import ThreadPoolExecutor
from datetime import date, timedelta
from typing import TYPE_CHECKING

import yfinance as yf

if TYPE_CHECKING:
    from services.cache import PriceCache


def _is_valid_ticker(symbol: str) -> bool:
    """Filter out CUSIPs and other non-tradeable identifiers.

    CUSIPs are 9-character alphanumeric identifiers (e.g., 542433VL8, 870462SA7).
    Valid stock/ETF tickers are typically 1-5 letters, sometimes with a dot (BRK.B).
    Crypto pairs use a hyphen (e.g., BTC-USD).
    """
    if not symbol:
        return False
    if len(symbol) >= 8 and re.search(r"\d", symbol):
        return False
    if re.fullmatch(r"[A-Z]{1,5}(\.[A-Z])?", symbol.upper()):
        return True
    if re.fullmatch(r"[A-Z]{2,6}", symbol.upper()):
        return True
    if re.fullmatch(r"[A-Z]{2,5}-[A-Z]{2,5}", symbol.upper()):
        return True
    return False


class MarketService:
    def __init__(self, cache: "PriceCache"):
        self.cache = cache

    async def get_quotes(self, symbols: list[str]) -> list[dict]:
        if not symbols:
            return []

        tradeable = [s for s in symbols if _is_valid_ticker(s)]
        if not tradeable:
            return []

        cached = await self.cache.get_cached_quotes(tradeable)
        missing = [s for s in tradeable if s.upper() not in cached]

        if missing:
            fresh = await self._fetch_quotes_from_yfinance(missing)
            if fresh:
                await self.cache.store_quotes(fresh)
                for q in fresh:
                    cached[q["symbol"]] = q

        return list(cached.values())

    async def _fetch_quotes_from_yfinance(self, symbols: list[str]) -> list[dict]:
        def fetch():
            df = yf.download(
                symbols,
                period="2d",
                interval="1d",
                progress=False,
                threads=True,
            )
            if df.empty:
                return []

            def _get_info(sym):
                try:
                    return sym, yf.Ticker(sym).info
                except Exception:
                    return sym, {}

            ticker_info = {}
            with ThreadPoolExecutor(max_workers=10) as pool:
                for sym, info in pool.map(lambda s: _get_info(s), symbols):
                    ticker_info[sym] = info

            results = []
            for symbol in symbols:
                try:
                    if len(symbols) == 1:
                        close_col = df["Close"]
                    else:
                        if symbol not in df["Close"].columns:
                            continue
                        close_col = df["Close"][symbol]

                    if hasattr(close_col, 'columns'):
                        close_col = close_col.iloc[:, 0]
                    if close_col.empty or close_col.isna().all():
                        continue

                    prices = close_col.dropna()
                    if len(prices) < 1:
                        continue

                    current_price = float(prices.iloc[-1])
                    prev_close = float(prices.iloc[-2]) if len(prices) >= 2 else None
                    change = current_price - prev_close if prev_close else None
                    change_pct = (change / prev_close * 100) if prev_close else None

                    info = ticker_info.get(symbol, {})
                    results.append(
                        {
                            "symbol": symbol.upper(),
                            "name": info.get("shortName") or info.get("longName"),
                            "price": current_price,
                            "change": change,
                            "change_percent": change_pct,
                            "previous_close": prev_close,
                            "volume": None,
                            "asset_type": _determine_asset_type(info),
                            "sector": info.get("sector"),
                            "industry": info.get("industry"),
                            "category": info.get("category"),
                            "dividend_rate": info.get("dividendRate"),
                            "dividend_yield": info.get("dividendYield"),
                            "yield_pct": info.get("yield"),
                        }
                    )
                except (KeyError, IndexError):
                    continue
            return results

        return await asyncio.to_thread(fetch)

    async def get_daily_chart(
        self, symbol: str, start_date: date, end_date: date
    ) -> list[dict]:
        if not _is_valid_ticker(symbol):
            return []

        cached = await self.cache.get_daily_prices(symbol, start_date, end_date)
        cached_dates = {p["timestamp"] for p in cached}

        all_dates = set()
        current = start_date
        while current <= end_date:
            all_dates.add(current.isoformat())
            current += timedelta(days=1)

        missing = all_dates - cached_dates
        if missing:
            new_points = await self._fetch_daily_history(
                symbol, min(missing), max(missing)
            )
            if new_points:
                await self.cache.store_daily_prices(symbol, new_points)
                cached = await self.cache.get_daily_prices(symbol, start_date, end_date)

        return cached

    async def get_intraday_chart(self, symbol: str, range_str: str) -> list[dict]:
        if not _is_valid_ticker(symbol):
            return []

        def fetch():
            ticker = yf.Ticker(symbol)
            interval, period = _parse_range(range_str)
            hist = ticker.history(period=period, interval=interval)
            points = []
            for idx, row in hist.iterrows():
                ts = idx.isoformat() if hasattr(idx, "isoformat") else str(idx)
                points.append(
                    {
                        "timestamp": ts,
                        "open": row.get("Open"),
                        "high": row.get("High"),
                        "low": row.get("Low"),
                        "close": row.get("Close"),
                        "volume": int(row.get("Volume", 0)),
                    }
                )
            return points

        return await asyncio.to_thread(fetch)

    async def _fetch_daily_history(
        self, symbol: str, start: str, end: str
    ) -> list[dict]:
        def fetch():
            ticker = yf.Ticker(symbol)
            hist = ticker.history(start=start, end=end, interval="1d")
            points = []
            for idx, row in hist.iterrows():
                points.append(
                    {
                        "timestamp": idx.strftime("%Y-%m-%d"),
                        "open": row.get("Open"),
                        "high": row.get("High"),
                        "low": row.get("Low"),
                        "close": row.get("Close"),
                        "volume": int(row.get("Volume", 0)),
                    }
                )
            return points

        return await asyncio.to_thread(fetch)


def _determine_asset_type(info: dict) -> str:
    qtype = info.get("quoteType", "").lower()
    if qtype == "etf":
        return "etf"
    if qtype == "mutualfund":
        return "mutual_fund"
    if qtype in ("equity", ""):
        return "equity"
    return qtype


def _parse_range(range_str: str) -> tuple[str, str]:
    mapping = {
        "1d": ("5m", "1d"),
        "5d": ("15m", "5d"),
        "1mo": ("1d", "1mo"),
        "3mo": ("1d", "3mo"),
        "6mo": ("1d", "6mo"),
        "1y": ("1d", "1y"),
        "2y": ("1wk", "2y"),
        "5y": ("1wk", "5y"),
        "max": ("1mo", "max"),
    }
    return mapping.get(range_str, ("1d", "1mo"))
