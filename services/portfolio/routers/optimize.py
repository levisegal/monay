import logging

from fastapi import APIRouter

from models.optimize import AnalyzeRequest, AnalyzeResponse, ReturnsRequest, ReturnsResponse
from services.market import MarketService
from services.optimizer import analyze_portfolio, compute_portfolio_returns

logger = logging.getLogger(__name__)


def create_router(market: MarketService) -> APIRouter:
    router = APIRouter()

    @router.post("/portfolio/analyze", response_model=AnalyzeResponse)
    async def post_analyze(req: AnalyzeRequest):
        return await analyze_portfolio(market, req)

    @router.post("/portfolio/returns", response_model=ReturnsResponse)
    async def post_returns(req: ReturnsRequest):
        daily_returns, rfr = await compute_portfolio_returns(market, req.holdings, req.total_value)
        return ReturnsResponse(returns=daily_returns, risk_free_rate=rfr)

    return router
