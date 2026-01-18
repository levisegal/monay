from fastapi import APIRouter, Query

from models.portfolio import EnrichedHolding, EnrichedPortfolio, PortfolioSummary
from services.holdings_client import HoldingsClient
from services.market import MarketService

router = APIRouter()


def create_router(holdings: HoldingsClient, market: MarketService) -> APIRouter:
    @router.get("/portfolio/enriched", response_model=EnrichedPortfolio)
    async def get_enriched_portfolio(
        account_id: str | None = Query(None, description="Filter by account ID"),
    ):
        raw_holdings = await holdings.get_holdings(account_id)

        if not raw_holdings:
            return EnrichedPortfolio(
                summary=PortfolioSummary(
                    total_value=0,
                    total_cost_basis=0,
                    total_unrealized_gain=0,
                    day_change=0,
                    day_change_percent=0,
                    holding_count=0,
                ),
                holdings=[],
            )

        symbols = list({h.get("symbol") for h in raw_holdings if h.get("symbol")})
        quotes_by_symbol = {}
        if symbols:
            raw_quotes = await market.get_quotes(symbols)
            quotes_by_symbol = {q["symbol"]: q for q in raw_quotes}

        enriched = []
        total_value = 0.0
        total_cost = 0.0
        total_day_change = 0.0

        for h in raw_holdings:
            symbol = h.get("symbol", "")
            quote = quotes_by_symbol.get(symbol, {})

            quantity = h.get("quantity", 0.0)
            current_price = quote.get("price")
            cost_basis_micros = h.get("cost_basis_micros")
            cost_basis = cost_basis_micros / 1_000_000 if cost_basis_micros else None

            market_value = quantity * current_price if current_price else None
            unrealized_gain = None
            unrealized_gain_percent = None
            if market_value and cost_basis:
                unrealized_gain = market_value - cost_basis
                unrealized_gain_percent = (
                    (unrealized_gain / cost_basis) * 100 if cost_basis > 0 else 0
                )

            day_change = None
            change_per_share = quote.get("change")
            if change_per_share is not None and quantity:
                day_change = change_per_share * quantity

            if market_value:
                total_value += market_value
            if cost_basis:
                total_cost += cost_basis
            if day_change:
                total_day_change += day_change

            enriched.append(
                EnrichedHolding(
                    account_id=h.get("account_id", ""),
                    account_name=h.get("account_name", ""),
                    symbol=symbol,
                    name=quote.get("name") or h.get("security_name"),
                    quantity=quantity,
                    cost_basis=cost_basis,
                    current_price=current_price,
                    market_value=market_value,
                    unrealized_gain=unrealized_gain,
                    unrealized_gain_percent=unrealized_gain_percent,
                    day_change=day_change,
                    allocation_percent=None,
                    asset_type=quote.get("asset_type"),
                )
            )

        for e in enriched:
            if e.market_value and total_value > 0:
                e.allocation_percent = (e.market_value / total_value) * 100

        total_unrealized = total_value - total_cost if total_cost > 0 else 0
        day_change_percent = (
            (total_day_change / (total_value - total_day_change)) * 100
            if total_value > total_day_change
            else 0
        )

        return EnrichedPortfolio(
            summary=PortfolioSummary(
                total_value=round(total_value, 2),
                total_cost_basis=round(total_cost, 2),
                total_unrealized_gain=round(total_unrealized, 2),
                day_change=round(total_day_change, 2),
                day_change_percent=round(day_change_percent, 2),
                holding_count=len(enriched),
            ),
            holdings=enriched,
        )

    return router
