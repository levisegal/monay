from pydantic import BaseModel


class ChartPoint(BaseModel):
    timestamp: str
    open: float | None
    high: float | None
    low: float | None
    close: float | None
    volume: int | None


class ChartResponse(BaseModel):
    symbol: str
    points: list[ChartPoint]
