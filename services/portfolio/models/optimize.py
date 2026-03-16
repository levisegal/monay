from pydantic import BaseModel


class HoldingInput(BaseModel):
    symbol: str
    value: float
    sector: str
    asset_class: str
    account_type: str
    is_muni: bool = False
    security_type: str = ""
    dividend_yield: float = 0.0


class Constraints(BaseModel):
    max_position_weight: float
    max_sector_weight: float
    target_equity_range: list[float]
    target_fixed_income_range: list[float]
    target_cash_range: list[float]


class PositionConviction(BaseModel):
    thesis: str
    range: list[float]


class SectorConviction(BaseModel):
    thesis: str
    target_weight: float
    instruments: list[str] = []


class StrategyConviction(BaseModel):
    thesis: str
    target_weight: float
    min_yield: float = 0.0
    instruments: list[str] = []
    location: str = ""


class AssetClassOverride(BaseModel):
    thesis: str = ""
    target_range: list[float]


class Convictions(BaseModel):
    positions: dict[str, PositionConviction] = {}
    sectors: dict[str, SectorConviction] = {}
    strategies: dict[str, StrategyConviction] = {}
    asset_classes: dict[str, AssetClassOverride] = {}


class ReturnsRequest(BaseModel):
    holdings: list[HoldingInput]
    total_value: float


class DailyReturn(BaseModel):
    date: str
    value: float


class ReturnsResponse(BaseModel):
    returns: list[DailyReturn]
    risk_free_rate: float


class AnalyzeRequest(BaseModel):
    holdings: list[HoldingInput]
    total_value: float
    constraints: Constraints
    convictions: Convictions | None = None


class RiskMetrics(BaseModel):
    annualized_volatility: float | None = None
    annualized_return: float | None = None
    cvar_95: float | None = None
    max_drawdown: float | None = None
    sharpe_ratio: float | None = None
    risk_free_rate: float | None = None


class TrailingReturns(BaseModel):
    one_month: float | None = None
    three_month: float | None = None
    six_month: float | None = None
    ytd: float | None = None
    one_year: float | None = None
    two_year_annualized: float | None = None


class MonthlyReturn(BaseModel):
    period: str
    return_pct: float


class SymbolPerformance(BaseModel):
    symbol: str
    weight: float
    one_month: float | None = None
    three_month: float | None = None
    ytd: float | None = None
    one_year: float | None = None
    contribution_one_year: float | None = None


class PerformanceSummary(BaseModel):
    as_of: str
    history_start: str
    current: TrailingReturns
    target: TrailingReturns
    monthly_returns: list[MonthlyReturn] = []
    symbols: list[SymbolPerformance] = []


class CorrelationCluster(BaseModel):
    symbols: list[str]
    avg_correlation: float
    combined_weight: float


class ConstraintViolation(BaseModel):
    violation_type: str
    name: str
    current: float
    limit: float
    excess: float


class LocationIssue(BaseModel):
    symbol: str
    name: str
    current_account_type: str
    recommended_account_type: str
    reason: str
    value: float


class ConvictionStatus(BaseModel):
    type: str
    name: str
    thesis: str
    target_weight: float
    current_weight: float
    status: str


class ConsolidationRec(BaseModel):
    symbol: str
    name: str
    current_weight: float
    reason: str
    merge_into: str


class RebalanceDelta(BaseModel):
    symbol: str
    name: str
    sector: str
    current_weight: float
    target_weight: float
    delta_dollars: float
    action: str
    best_account: str
    note: str = ""


class AnalyzeResponse(BaseModel):
    risk_metrics: RiskMetrics
    target_risk_metrics: RiskMetrics
    performance: PerformanceSummary | None = None
    correlation_clusters: list[CorrelationCluster]
    constraint_violations: list[ConstraintViolation]
    location_issues: list[LocationIssue]
    target_weights: dict[str, float]
    rebalance_deltas: list[RebalanceDelta]
    conviction_statuses: list[ConvictionStatus] = []
    consolidation_recs: list[ConsolidationRec] = []
