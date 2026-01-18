import asyncio
from datetime import date, timedelta
from typing import TYPE_CHECKING

import yfinance as yf

if TYPE_CHECKING:
    from services.cache import PriceCache


class MarketService:
    def __init__(self, cache: "PriceCache"):
        self.cache = cache

    async def get_quotes(self, symbols: list[str]) -> list[dict]:
        if not symbols:
            return []

        def fetch():
            tickers = yf.Tickers(" ".join(symbols))
            results = []
            for symbol in symbols:
                ticker = tickers.tickers.get(symbol.upper())
                if not ticker:
                    continue
                info = ticker.info
                results.append(
                    {
                        "symbol": symbol.upper(),
                        "name": info.get("shortName") or info.get("longName"),
                        "price": info.get("currentPrice")
                        or info.get("regularMarketPrice"),
                        "change": info.get("regularMarketChange"),
                        "change_percent": info.get("regularMarketChangePercent"),
                        "previous_close": info.get("previousClose"),
                        "volume": info.get("volume"),
                        "asset_type": _determine_asset_type(info),
                    }
                )
            return results

        return await asyncio.to_thread(fetch)

    async def get_daily_chart(
        self, symbol: str, start_date: date, end_date: date
    ) -> list[dict]:
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
