from datetime import date, timedelta

from fastapi import APIRouter, Query

from models.chart import ChartPoint, ChartResponse
from services.market import MarketService


def create_router(market: MarketService) -> APIRouter:
    router = APIRouter()

    @router.get("/chart/{symbol}", response_model=ChartResponse)
    async def get_chart(
        symbol: str,
        range: str = Query("1y", description="Time range: 1d, 5d, 1mo, 3mo, 6mo, 1y, 2y, 5y, max"),
        interval: str = Query("1d", description="Interval: 1d for daily (default), 5m for intraday"),
    ):
        symbol = symbol.upper()

        if interval != "1d" or range in ("1d", "5d"):
            points = await market.get_intraday_chart(symbol, range)
        else:
            end_date = date.today()
            start_date = _range_to_start_date(range, end_date)
            points = await market.get_daily_chart(symbol, start_date, end_date)

        chart_points = [ChartPoint(**p) for p in points]
        return ChartResponse(symbol=symbol, points=chart_points)

    return router


def _range_to_start_date(range_str: str, end: date) -> date:
    mapping = {
        "1mo": timedelta(days=30),
        "3mo": timedelta(days=90),
        "6mo": timedelta(days=180),
        "1y": timedelta(days=365),
        "2y": timedelta(days=730),
        "5y": timedelta(days=1825),
        "max": timedelta(days=3650),
    }
    delta = mapping.get(range_str, timedelta(days=365))
    return end - delta
