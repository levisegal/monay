from pydantic import BaseModel


class Quote(BaseModel):
    symbol: str
    name: str | None
    price: float | None
    change: float | None
    change_percent: float | None
    previous_close: float | None
    volume: int | None
    asset_type: str | None
    sector: str | None = None
    industry: str | None = None
    category: str | None = None
    dividend_rate: float | None = None
    dividend_yield: float | None = None
    yield_pct: float | None = None
    net_expense_ratio: float | None = None
    fund_family: str | None = None


class QuotesResponse(BaseModel):
    quotes: list[Quote]
