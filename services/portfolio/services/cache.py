from datetime import date, datetime

import aiosqlite

from config import get_settings

SCHEMA = """
CREATE TABLE IF NOT EXISTS daily_prices (
    symbol TEXT NOT NULL,
    date TEXT NOT NULL,
    open REAL,
    high REAL,
    low REAL,
    close REAL,
    volume INTEGER,
    PRIMARY KEY (symbol, date)
);

CREATE TABLE IF NOT EXISTS quote_cache (
    symbol TEXT PRIMARY KEY,
    price REAL,
    change REAL,
    change_percent REAL,
    previous_close REAL,
    cached_at TEXT NOT NULL
);
"""


class PriceCache:
    def __init__(self, db_path: str | None = None):
        settings = get_settings()
        self.db_path = db_path or settings.cache_path
        self._db: aiosqlite.Connection | None = None

    async def connect(self):
        self._db = await aiosqlite.connect(self.db_path)
        for stmt in SCHEMA.strip().split(";"):
            if stmt.strip():
                await self._db.execute(stmt)
        await self._db.commit()

    async def close(self):
        if self._db:
            await self._db.close()
            self._db = None

    async def get_daily_prices(
        self, symbol: str, start_date: date, end_date: date
    ) -> list[dict]:
        if not self._db:
            raise RuntimeError("Database not connected")

        cursor = await self._db.execute(
            """
            SELECT date, open, high, low, close, volume
            FROM daily_prices
            WHERE symbol = ? AND date >= ? AND date <= ?
            ORDER BY date ASC
            """,
            (symbol.upper(), start_date.isoformat(), end_date.isoformat()),
        )
        rows = await cursor.fetchall()
        return [
            {
                "timestamp": row[0],
                "open": row[1],
                "high": row[2],
                "low": row[3],
                "close": row[4],
                "volume": row[5],
            }
            for row in rows
        ]

    async def get_cached_dates(self, symbol: str) -> set[str]:
        if not self._db:
            raise RuntimeError("Database not connected")

        cursor = await self._db.execute(
            "SELECT date FROM daily_prices WHERE symbol = ?", (symbol.upper(),)
        )
        rows = await cursor.fetchall()
        return {row[0] for row in rows}

    async def store_daily_prices(self, symbol: str, points: list[dict]):
        if not self._db:
            raise RuntimeError("Database not connected")

        await self._db.executemany(
            """
            INSERT OR REPLACE INTO daily_prices (symbol, date, open, high, low, close, volume)
            VALUES (?, ?, ?, ?, ?, ?, ?)
            """,
            [
                (
                    symbol.upper(),
                    p["timestamp"],
                    p.get("open"),
                    p.get("high"),
                    p.get("low"),
                    p.get("close"),
                    p.get("volume"),
                )
                for p in points
            ],
        )
        await self._db.commit()

    async def get_cached_quotes(self, symbols: list[str], max_age_minutes: int = 5) -> dict[str, dict]:
        """Get cached quotes that are not stale."""
        if not self._db:
            raise RuntimeError("Database not connected")

        if not symbols:
            return {}

        placeholders = ",".join("?" for _ in symbols)
        cursor = await self._db.execute(
            f"""
            SELECT symbol, price, change, change_percent, previous_close, cached_at
            FROM quote_cache
            WHERE symbol IN ({placeholders})
            """,
            [s.upper() for s in symbols],
        )
        rows = await cursor.fetchall()

        result = {}
        for row in rows:
            cached_at = datetime.fromisoformat(row[5])
            age_minutes = (datetime.utcnow() - cached_at).total_seconds() / 60
            if age_minutes <= max_age_minutes:
                result[row[0]] = {
                    "symbol": row[0],
                    "name": None,
                    "price": row[1],
                    "change": row[2],
                    "change_percent": row[3],
                    "previous_close": row[4],
                    "volume": None,
                    "asset_type": "equity",
                }
        return result

    async def store_quotes(self, quotes: list[dict]):
        """Store quotes in cache."""
        if not self._db:
            raise RuntimeError("Database not connected")

        now = datetime.utcnow().isoformat()
        await self._db.executemany(
            """
            INSERT OR REPLACE INTO quote_cache (symbol, price, change, change_percent, previous_close, cached_at)
            VALUES (?, ?, ?, ?, ?, ?)
            """,
            [
                (
                    q["symbol"].upper(),
                    q.get("price"),
                    q.get("change"),
                    q.get("change_percent"),
                    q.get("previous_close"),
                    now,
                )
                for q in quotes
            ],
        )
        await self._db.commit()
