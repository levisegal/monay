import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from config import get_settings
from routers import chart, portfolio, quotes
from services.cache import PriceCache
from services.holdings_client import HoldingsClient
from services.market import MarketService

logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    settings = get_settings()
    logger.info("Starting Portfolio Service...")
    logger.info(f"  Holdings URL: {settings.holdings_url}")
    logger.info(f"  Cache path: {settings.cache_path}")
    logger.info(f"  Listen addr: {settings.listen_addr}")

    cache = PriceCache()
    await cache.connect()
    holdings = HoldingsClient()
    market = MarketService(cache)

    app.state.cache = cache
    app.state.holdings = holdings
    app.state.market = market

    app.include_router(
        quotes.create_router(market), prefix="/api/v1", tags=["quotes"]
    )
    app.include_router(chart.create_router(market), prefix="/api/v1", tags=["chart"])
    app.include_router(
        portfolio.create_router(holdings, market), prefix="/api/v1", tags=["portfolio"]
    )

    yield

    logger.info("Shutting down Portfolio Service...")
    await holdings.close()
    await cache.close()


app = FastAPI(
    title="Monay Portfolio Service",
    description="Gateway service for portfolio data enriched with live market prices",
    version="0.1.0",
    lifespan=lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
    allow_credentials=False,
)


@app.get("/api/v1/health")
async def health():
    return {"status": "ok"}


if __name__ == "__main__":
    import uvicorn

    settings = get_settings()
    uvicorn.run(app, host=settings.host, port=settings.port)
