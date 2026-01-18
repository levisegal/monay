from pydantic import BaseModel


class PortfolioSummary(BaseModel):
    total_value: float
    total_cost_basis: float
    total_unrealized_gain: float
    day_change: float
    day_change_percent: float
    holding_count: int


class EnrichedHolding(BaseModel):
    account_id: str
    account_name: str
    symbol: str
    name: str | None
    quantity: float
    cost_basis: float | None
    current_price: float | None
    market_value: float | None
    unrealized_gain: float | None
    unrealized_gain_percent: float | None
    day_change: float | None
    allocation_percent: float | None
    asset_type: str | None


class EnrichedPortfolio(BaseModel):
    summary: PortfolioSummary
    holdings: list[EnrichedHolding]
