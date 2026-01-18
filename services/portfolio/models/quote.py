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


class QuotesResponse(BaseModel):
    quotes: list[Quote]
