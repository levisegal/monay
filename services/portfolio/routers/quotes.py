from fastapi import APIRouter, Query

from models.quote import Quote, QuotesResponse
from services.market import MarketService

router = APIRouter()


def create_router(market: MarketService) -> APIRouter:
    @router.get("/quotes", response_model=QuotesResponse)
    async def get_quotes(symbols: str = Query(..., description="Comma-separated symbols")):
        symbol_list = [s.strip().upper() for s in symbols.split(",") if s.strip()]
        raw_quotes = await market.get_quotes(symbol_list)
        quotes = [Quote(**q) for q in raw_quotes]
        return QuotesResponse(quotes=quotes)

    return router
