from datetime import date

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
"""


class PriceCache:
    def __init__(self, db_path: str | None = None):
        settings = get_settings()
        self.db_path = db_path or settings.cache_path
        self._db: aiosqlite.Connection | None = None

    async def connect(self):
        self._db = await aiosqlite.connect(self.db_path)
        await self._db.execute(SCHEMA)
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
