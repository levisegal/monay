import tempfile
from datetime import date
from pathlib import Path

import pytest

import sys
sys.path.insert(0, str(Path(__file__).parent.parent))

from services.cache import PriceCache


@pytest.fixture
async def cache():
    with tempfile.NamedTemporaryFile(suffix=".db", delete=False) as f:
        db_path = f.name

    c = PriceCache(db_path)
    await c.connect()
    yield c
    await c.close()
    Path(db_path).unlink(missing_ok=True)


@pytest.mark.asyncio
async def test_store_and_retrieve_prices(cache):
    points = [
        {"timestamp": "2024-01-15", "open": 100.0, "high": 105.0, "low": 99.0, "close": 104.0, "volume": 1000000},
        {"timestamp": "2024-01-16", "open": 104.0, "high": 108.0, "low": 103.0, "close": 107.0, "volume": 1200000},
    ]

    await cache.store_daily_prices("AAPL", points)

    result = await cache.get_daily_prices("AAPL", date(2024, 1, 1), date(2024, 1, 31))

    assert len(result) == 2
    assert result[0]["timestamp"] == "2024-01-15"
    assert result[0]["close"] == 104.0
    assert result[1]["timestamp"] == "2024-01-16"
    assert result[1]["close"] == 107.0


@pytest.mark.asyncio
async def test_get_cached_dates(cache):
    points = [
        {"timestamp": "2024-01-15", "open": 100.0, "high": 105.0, "low": 99.0, "close": 104.0, "volume": 1000000},
        {"timestamp": "2024-01-17", "open": 106.0, "high": 110.0, "low": 105.0, "close": 109.0, "volume": 1100000},
    ]

    await cache.store_daily_prices("MSFT", points)

    dates = await cache.get_cached_dates("MSFT")

    assert dates == {"2024-01-15", "2024-01-17"}


@pytest.mark.asyncio
async def test_symbol_case_insensitive(cache):
    points = [
        {"timestamp": "2024-01-15", "open": 100.0, "high": 105.0, "low": 99.0, "close": 104.0, "volume": 1000000},
    ]

    await cache.store_daily_prices("aapl", points)

    result = await cache.get_daily_prices("AAPL", date(2024, 1, 1), date(2024, 1, 31))
    assert len(result) == 1


@pytest.mark.asyncio
async def test_upsert_replaces_existing(cache):
    await cache.store_daily_prices("AAPL", [
        {"timestamp": "2024-01-15", "open": 100.0, "high": 105.0, "low": 99.0, "close": 104.0, "volume": 1000000},
    ])

    await cache.store_daily_prices("AAPL", [
        {"timestamp": "2024-01-15", "open": 101.0, "high": 106.0, "low": 100.0, "close": 105.0, "volume": 1100000},
    ])

    result = await cache.get_daily_prices("AAPL", date(2024, 1, 1), date(2024, 1, 31))
    assert len(result) == 1
    assert result[0]["close"] == 105.0
