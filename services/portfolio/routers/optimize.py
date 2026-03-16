import logging

from fastapi import APIRouter

from models.optimize import AnalyzeRequest, AnalyzeResponse
from services.market import MarketService
from services.optimizer import analyze_portfolio

logger = logging.getLogger(__name__)


def create_router(market: MarketService) -> APIRouter:
    router = APIRouter()

    @router.post("/portfolio/analyze", response_model=AnalyzeResponse)
    async def post_analyze(req: AnalyzeRequest):
        return await analyze_portfolio(market, req)

    return router
