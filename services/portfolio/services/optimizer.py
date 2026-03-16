import asyncio
import logging
import math
from collections import defaultdict
from datetime import date, timedelta
from typing import TYPE_CHECKING, cast

import numpy as np
import pandas as pd

from models.optimize import (
    AnalyzeRequest,
    AnalyzeResponse,
    ConsolidationRec,
    ConstraintViolation,
    ConvictionStatus,
    Convictions,
    CorrelationCluster,
    LocationIssue,
    MonthlyReturn,
    PerformanceSummary,
    RebalanceDelta,
    RiskMetrics,
    SymbolPerformance,
    TrailingReturns,
)

if TYPE_CHECKING:
    from services.market import MarketService

logger = logging.getLogger(__name__)

TRADING_DAYS_PER_YEAR = 252
RISK_FREE_RATE = 0.045


async def analyze_portfolio(
    market: "MarketService", req: AnalyzeRequest
) -> AnalyzeResponse:
    symbols, weights = aggregate_symbol_weights(req.holdings, req.total_value)

    returns_df = await _fetch_returns(market, symbols)

    risk_metrics = compute_risk_metrics(returns_df, weights)
    clusters = find_correlation_clusters(returns_df, weights)
    violations = find_constraint_violations(
        req.holdings, req.total_value, req.constraints
    )
    location_issues = find_location_issues(req.holdings)

    conviction_statuses: list[ConvictionStatus] = []
    consolidation_recs: list[ConsolidationRec] = []

    if req.convictions:
        target_weights = compute_conviction_target_weights(
            req.holdings,
            req.total_value,
            req.constraints,
            req.convictions,
        )
        conviction_statuses = compute_conviction_statuses(
            req.holdings,
            req.total_value,
            req.convictions,
        )
        consolidation_recs = find_consolidation_candidates(
            returns_df,
            weights,
            req.holdings,
            req.convictions,
        )
        deltas = compute_rebalance_deltas(req.holdings, target_weights, req.total_value)
        gap_deltas = compute_gap_fill_deltas(
            req.holdings,
            target_weights,
            req.total_value,
            req.convictions,
        )
        deltas.extend(gap_deltas)
        annotate_conviction_deltas(deltas, req.convictions, consolidation_recs)
    else:
        target_weights = compute_target_weights(
            req.holdings, req.total_value, req.constraints
        )
        deltas = compute_rebalance_deltas(req.holdings, target_weights, req.total_value)

    all_symbols = list(set(symbols + [d.symbol for d in deltas]))
    target_weight_list = [target_weights.get(s, 0.0) for s in all_symbols]
    target_risk = compute_risk_metrics(
        returns_df, dict(zip(all_symbols, target_weight_list))
    )
    performance = compute_performance_summary(returns_df, weights, target_weights)

    return AnalyzeResponse(
        risk_metrics=risk_metrics,
        target_risk_metrics=target_risk,
        performance=performance,
        correlation_clusters=clusters,
        constraint_violations=violations,
        location_issues=location_issues,
        target_weights=target_weights,
        rebalance_deltas=deltas,
        conviction_statuses=conviction_statuses,
        consolidation_recs=consolidation_recs,
    )


def aggregate_symbol_weights(
    holdings: list, total_value: float
) -> tuple[list[str], dict[str, float]]:
    if total_value <= 0:
        return [], {}

    symbol_values: dict[str, float] = defaultdict(float)
    for holding in holdings:
        symbol_values[holding.symbol] += holding.value

    return list(symbol_values.keys()), {
        symbol: value / total_value for symbol, value in symbol_values.items()
    }


def compute_risk_metrics(
    returns_df: pd.DataFrame, weights: dict[str, float]
) -> RiskMetrics:
    if returns_df.empty:
        return RiskMetrics()

    available = [s for s in weights if s in returns_df.columns]
    if len(available) < 2:
        return RiskMetrics()

    df = returns_df[available].dropna()
    if len(df) < 30:
        return RiskMetrics()

    w = np.array([weights.get(s, 0.0) for s in available])
    w_sum = w.sum()
    if w_sum <= 0:
        return RiskMetrics()
    w = w / w_sum

    cov = np.cov(np.asarray(df.to_numpy(dtype=float), dtype=float), rowvar=False)
    port_var = float(w @ cov @ w)
    daily_vol = math.sqrt(max(port_var, 0))
    ann_vol = daily_vol * math.sqrt(TRADING_DAYS_PER_YEAR)

    port_returns = df.values @ w
    mean_daily = float(np.mean(port_returns))
    ann_return = mean_daily * TRADING_DAYS_PER_YEAR

    sharpe = (ann_return - RISK_FREE_RATE) / ann_vol if ann_vol > 0 else None

    sorted_returns = np.sort(port_returns)
    cutoff = max(int(len(sorted_returns) * 0.05), 1)
    cvar_daily = float(np.mean(sorted_returns[:cutoff]))
    cvar_ann = cvar_daily * math.sqrt(TRADING_DAYS_PER_YEAR)

    cumulative = np.cumprod(1 + port_returns)
    running_max = np.maximum.accumulate(cumulative)
    drawdowns = (cumulative - running_max) / running_max
    max_dd = float(np.min(drawdowns))

    return RiskMetrics(
        annualized_volatility=round(ann_vol, 4),
        cvar_95=round(cvar_ann, 4),
        max_drawdown=round(max_dd, 4),
        sharpe_ratio=round(sharpe, 4) if sharpe is not None else None,
    )


def compute_performance_summary(
    returns_df: pd.DataFrame,
    current_weights: dict[str, float],
    target_weights: dict[str, float],
) -> PerformanceSummary | None:
    portfolio_returns = build_portfolio_return_series(returns_df, current_weights)
    if portfolio_returns is None or portfolio_returns.empty:
        return None

    target_returns = build_portfolio_return_series(returns_df, target_weights)
    history_index = pd.DatetimeIndex(portfolio_returns.index)
    if len(history_index) == 0:
        return None

    as_of = cast(pd.Timestamp, history_index[-1])
    history_start = cast(pd.Timestamp, history_index[0])

    return PerformanceSummary(
        as_of=as_of.date().isoformat(),
        history_start=history_start.date().isoformat(),
        current=compute_trailing_returns(portfolio_returns),
        target=compute_trailing_returns(target_returns),
        monthly_returns=compute_monthly_returns(portfolio_returns),
        symbols=compute_symbol_performance(returns_df, current_weights),
    )


def build_portfolio_return_series(
    returns_df: pd.DataFrame, weights: dict[str, float]
) -> pd.Series | None:
    if returns_df.empty:
        return None

    available = [
        symbol
        for symbol in weights
        if symbol in returns_df.columns and weights.get(symbol, 0.0) > 0
    ]
    if not available:
        return None

    df = returns_df[available].copy()
    if len(df) < 21:
        return None

    weight_array = np.array(
        [weights.get(symbol, 0.0) for symbol in available], dtype=float
    )
    weighted_returns = df.mul(weight_array, axis=1)
    available_weight = df.notna().mul(weight_array, axis=1).sum(axis=1)
    portfolio_returns = weighted_returns.sum(axis=1, skipna=True).div(available_weight)
    portfolio_returns = portfolio_returns.replace([np.inf, -np.inf], np.nan).dropna()
    if len(portfolio_returns) < 21:
        return None

    portfolio_returns.index = pd.to_datetime(portfolio_returns.index)
    return cast(pd.Series, portfolio_returns)


def compute_trailing_returns(series: pd.Series | None) -> TrailingReturns:
    if series is None or series.empty:
        return TrailingReturns()

    year_start = pd.Timestamp(year=date.today().year, month=1, day=1)

    ytd_series = cast(pd.Series, series[series.index >= year_start])

    return TrailingReturns(
        one_month=trailing_return(series, 21),
        three_month=trailing_return(series, 63),
        six_month=trailing_return(series, 126),
        ytd=period_return(ytd_series),
        one_year=trailing_return(series, 252),
        two_year_annualized=annualized_return(series, 504),
    )


def compute_monthly_returns(series: pd.Series) -> list[MonthlyReturn]:
    current_year = date.today().year
    index = pd.DatetimeIndex(series.index)
    year_mask = [cast(pd.Timestamp, ts).year == current_year for ts in index]
    year_series = cast(pd.Series, series[year_mask])
    if year_series.empty:
        return []

    monthlies = cast(pd.Series, year_series.resample("ME").apply(period_return))
    results = []
    for idx, value in monthlies.items():
        if pd.isna(value):
            continue
        month = cast(pd.Timestamp, idx)
        results.append(
            MonthlyReturn(
                period=month.strftime("%Y-%m"), return_pct=round(float(value), 4)
            )
        )
    return results


def compute_symbol_performance(
    returns_df: pd.DataFrame, weights: dict[str, float]
) -> list[SymbolPerformance]:
    if returns_df.empty:
        return []

    df = returns_df.copy()
    df.index = pd.to_datetime(df.index)

    rows = []
    for symbol, weight in weights.items():
        if weight <= 0 or symbol not in df.columns:
            continue

        series = cast(pd.Series, df[symbol].dropna())
        if len(series) < 21:
            continue

        one_year = trailing_return(series, 252)
        ytd_series = cast(
            pd.Series,
            series[
                series.index >= pd.Timestamp(year=date.today().year, month=1, day=1)
            ],
        )
        rows.append(
            SymbolPerformance(
                symbol=symbol,
                weight=round(weight, 4),
                one_month=trailing_return(series, 21),
                three_month=trailing_return(series, 63),
                ytd=period_return(ytd_series),
                one_year=one_year,
                contribution_one_year=round(weight * one_year, 4)
                if one_year is not None
                else None,
            )
        )

    rows.sort(
        key=lambda row: (
            row.one_year is None,
            -(row.one_year or 0.0),
            -(row.ytd or 0.0),
            -row.weight,
        )
    )
    return rows


def trailing_return(series: pd.Series, periods: int) -> float | None:
    if len(series) < periods:
        return None
    window = series.iloc[-periods:]
    return period_return(window)


def annualized_return(series: pd.Series, periods: int) -> float | None:
    if len(series) < periods:
        return None
    window = series.iloc[-periods:]
    total_return = period_return(window)
    if total_return is None:
        return None
    years = len(window) / TRADING_DAYS_PER_YEAR
    if years <= 0:
        return None
    return round((1 + total_return) ** (1 / years) - 1, 4)


def period_return(series: pd.Series) -> float | None:
    if series.empty:
        return None
    return round(float((1 + series).prod() - 1), 4)


def find_correlation_clusters(
    returns_df: pd.DataFrame, weights: dict[str, float], threshold: float = 0.7
) -> list[CorrelationCluster]:
    if returns_df.empty or len(returns_df.columns) < 2:
        return []

    corr = returns_df.corr()
    symbols = list(corr.columns)

    parent = {s: s for s in symbols}

    def find(x):
        while parent[x] != x:
            parent[x] = parent[parent[x]]
            x = parent[x]
        return x

    def union(a, b):
        ra, rb = find(a), find(b)
        if ra != rb:
            parent[ra] = rb

    for i, s1 in enumerate(symbols):
        for s2 in symbols[i + 1 :]:
            if abs(corr.loc[s1, s2]) >= threshold:
                union(s1, s2)

    groups: dict[str, list[str]] = defaultdict(list)
    for s in symbols:
        groups[find(s)].append(s)

    clusters = []
    for members in groups.values():
        if len(members) < 2:
            continue

        pair_corrs = []
        for i, s1 in enumerate(members):
            for s2 in members[i + 1 :]:
                pair_corrs.append(abs(corr.loc[s1, s2]))

        avg_corr = float(np.mean(pair_corrs)) if pair_corrs else 0.0
        combined = sum(weights.get(s, 0.0) for s in members)

        clusters.append(
            CorrelationCluster(
                symbols=sorted(members),
                avg_correlation=round(avg_corr, 3),
                combined_weight=round(combined, 4),
            )
        )

    clusters.sort(key=lambda c: c.combined_weight, reverse=True)
    return clusters


def find_constraint_violations(
    holdings: list, total_value: float, constraints
) -> list[ConstraintViolation]:
    if total_value <= 0:
        return []

    violations = []

    for h in holdings:
        w = h.value / total_value
        if w > constraints.max_position_weight:
            violations.append(
                ConstraintViolation(
                    violation_type="position",
                    name=h.symbol,
                    current=round(w, 4),
                    limit=constraints.max_position_weight,
                    excess=round(w - constraints.max_position_weight, 4),
                )
            )

    sector_values: dict[str, float] = defaultdict(float)
    for h in holdings:
        sector_values[h.sector] += h.value
    for sector, val in sector_values.items():
        w = val / total_value
        if w > constraints.max_sector_weight:
            violations.append(
                ConstraintViolation(
                    violation_type="sector",
                    name=sector,
                    current=round(w, 4),
                    limit=constraints.max_sector_weight,
                    excess=round(w - constraints.max_sector_weight, 4),
                )
            )

    class_values: dict[str, float] = defaultdict(float)
    for h in holdings:
        class_values[h.asset_class] += h.value
    class_ranges = {
        "equity": constraints.target_equity_range,
        "fixed_income": constraints.target_fixed_income_range,
        "cash": constraints.target_cash_range,
    }
    for cls, rng in class_ranges.items():
        w = class_values.get(cls, 0.0) / total_value
        if w > rng[1]:
            violations.append(
                ConstraintViolation(
                    violation_type="asset_class",
                    name=cls,
                    current=round(w, 4),
                    limit=rng[1],
                    excess=round(w - rng[1], 4),
                )
            )
        elif w < rng[0]:
            violations.append(
                ConstraintViolation(
                    violation_type="asset_class",
                    name=cls,
                    current=round(w, 4),
                    limit=rng[0],
                    excess=round(w - rng[0], 4),
                )
            )

    violations.sort(key=lambda v: abs(v.excess), reverse=True)
    return violations


def find_location_issues(holdings: list) -> list[LocationIssue]:
    issues = []
    ira_types = {"traditional_ira", "sep_ira"}
    tax_deferred = ira_types | {"roth_ira"}

    for h in holdings:
        if h.is_muni and h.account_type in tax_deferred:
            issues.append(
                LocationIssue(
                    symbol=h.symbol,
                    name=h.symbol,
                    current_account_type=h.account_type,
                    recommended_account_type="taxable",
                    reason="Muni interest is tax-free only in taxable accounts — wasted inside any IRA",
                    value=h.value,
                )
            )
            continue

        if (
            h.asset_class == "fixed_income"
            and not h.is_muni
            and h.account_type == "roth_ira"
        ):
            issues.append(
                LocationIssue(
                    symbol=h.symbol,
                    name=h.symbol,
                    current_account_type="roth_ira",
                    recommended_account_type="traditional_ira",
                    reason="Taxable bond interest is ordinary income — defer in traditional IRA, save Roth for growth",
                    value=h.value,
                )
            )
            continue

        if h.security_type == "reit" and h.account_type == "taxable":
            issues.append(
                LocationIssue(
                    symbol=h.symbol,
                    name=h.symbol,
                    current_account_type="taxable",
                    recommended_account_type="traditional_ira",
                    reason="REIT distributions are ordinary income — shelter in IRA",
                    value=h.value,
                )
            )
            continue

        if (
            h.asset_class == "equity"
            and h.security_type == "equity"
            and h.account_type in ira_types
        ):
            issues.append(
                LocationIssue(
                    symbol=h.symbol,
                    name=h.symbol,
                    current_account_type=h.account_type,
                    recommended_account_type="roth_ira",
                    reason="High-growth individual stock — maximize tax-free compounding in Roth",
                    value=h.value,
                )
            )
            continue

    issues.sort(key=lambda i: i.value, reverse=True)
    return issues


def compute_target_weights(
    holdings: list, total_value: float, constraints
) -> dict[str, float]:
    if total_value <= 0:
        return {}

    weights = {h.symbol: h.value / total_value for h in holdings}
    symbols_by_class: dict[str, list[str]] = defaultdict(list)
    symbol_sector: dict[str, str] = {}
    symbol_class: dict[str, str] = {}

    for h in holdings:
        symbols_by_class[h.asset_class].append(h.symbol)
        symbol_sector[h.symbol] = h.sector
        symbol_class[h.symbol] = h.asset_class

    for _ in range(3):
        weights = _cap_positions(
            weights, constraints.max_position_weight, symbol_class, symbols_by_class
        )
        weights = _cap_sectors(weights, constraints.max_sector_weight, symbol_sector)
        weights = _cap_asset_classes(weights, symbols_by_class, constraints)

    total = sum(weights.values())
    if total > 0:
        weights = {s: w / total for s, w in weights.items()}

    return {s: round(w, 6) for s, w in weights.items()}


def compute_rebalance_deltas(
    holdings: list, target_weights: dict[str, float], total_value: float
) -> list[RebalanceDelta]:
    deltas = []
    holding_map = {h.symbol: h for h in holdings}

    for h in holdings:
        current_w = h.value / total_value if total_value > 0 else 0
        target_w = target_weights.get(h.symbol, 0.0)
        delta = (target_w - current_w) * total_value

        if abs(delta) < 100:
            continue

        action = "BUY" if delta > 0 else "SELL"
        best_acct = _best_account_for_trade(h, action)

        deltas.append(
            RebalanceDelta(
                symbol=h.symbol,
                name=h.symbol,
                sector=h.sector,
                current_weight=round(current_w, 4),
                target_weight=round(target_w, 4),
                delta_dollars=round(delta, 2),
                action=action,
                best_account=best_acct,
            )
        )

    deltas.sort(key=lambda d: d.delta_dollars)
    return deltas


DEFAULT_GAP_FILL_INSTRUMENTS = {
    "fixed_income": ["BND", "AGG"],
    "muni": ["VTEB", "MUB"],
    "us_large_cap": ["VOO", "FXAIX"],
    "international": ["VXUS", "IXUS"],
    "tips": ["SCHP", "VTIP"],
    "small_cap": ["VB", "AVUV"],
    "dividend": ["SCHD", "VIG"],
}


def compute_conviction_target_weights(
    holdings: list,
    total_value: float,
    constraints,
    convictions: Convictions,
) -> dict[str, float]:
    if total_value <= 0:
        return {}

    weights = {h.symbol: h.value / total_value for h in holdings}
    symbols_by_class: dict[str, list[str]] = defaultdict(list)
    symbol_sector: dict[str, str] = {}
    symbol_class: dict[str, str] = {}

    for h in holdings:
        symbols_by_class[h.asset_class].append(h.symbol)
        symbol_sector[h.symbol] = h.sector
        symbol_class[h.symbol] = h.asset_class

    locked: set[str] = set()
    for sym, pc in convictions.positions.items():
        if sym in weights:
            clamped = max(pc.range[0], min(weights[sym], pc.range[1]))
            weights[sym] = clamped
            locked.add(sym)

    ac_overrides = {}
    for cls, override in convictions.asset_classes.items():
        ac_overrides[cls] = override.target_range

    effective_constraints = _EffectiveConstraints(constraints, ac_overrides)

    for _ in range(3):
        weights = _cap_positions(
            weights,
            constraints.max_position_weight,
            symbol_class,
            symbols_by_class,
            locked,
        )
        weights = _cap_sectors(
            weights, constraints.max_sector_weight, symbol_sector, locked
        )
        weights = _cap_asset_classes_with_overrides(
            weights,
            symbols_by_class,
            effective_constraints,
            locked,
        )

    total = sum(weights.values())
    if total > 0:
        weights = {s: w / total for s, w in weights.items()}

    return {s: round(w, 6) for s, w in weights.items()}


def compute_conviction_statuses(
    holdings: list,
    total_value: float,
    convictions: Convictions,
) -> list[ConvictionStatus]:
    if total_value <= 0:
        return []

    symbol_weights: dict[str, float] = defaultdict(float)
    sector_weights: dict[str, float] = defaultdict(float)
    class_weights: dict[str, float] = defaultdict(float)

    for h in holdings:
        w = h.value / total_value
        symbol_weights[h.symbol] += w
        sector_weights[h.sector] += w
        class_weights[h.asset_class] += w

    statuses = []

    for sym, pc in convictions.positions.items():
        cw = symbol_weights.get(sym, 0.0)
        mid = (pc.range[0] + pc.range[1]) / 2
        status = "in_range"
        if cw < pc.range[0]:
            status = "underweight"
        elif cw > pc.range[1]:
            status = "overweight"
        statuses.append(
            ConvictionStatus(
                type="position",
                name=sym,
                thesis=pc.thesis,
                target_weight=mid,
                current_weight=round(cw, 4),
                status=status,
            )
        )

    for name, sc in convictions.sectors.items():
        cw = sector_weights.get(name, 0.0)
        status = "in_range"
        if cw < sc.target_weight * 0.8:
            status = "underweight"
        elif cw > sc.target_weight * 1.2:
            status = "overweight"
        statuses.append(
            ConvictionStatus(
                type="sector",
                name=name,
                thesis=sc.thesis,
                target_weight=sc.target_weight,
                current_weight=round(cw, 4),
                status=status,
            )
        )

    for name, strat in convictions.strategies.items():
        held_syms = set(symbol_weights.keys())
        strat_syms = set(strat.instruments) & held_syms
        cw = sum(symbol_weights.get(s, 0.0) for s in strat_syms)
        for h in holdings:
            if (
                h.symbol not in strat_syms
                and strat.min_yield > 0
                and h.dividend_yield >= strat.min_yield
            ):
                cw += h.value / total_value
        status = "in_range"
        if cw < strat.target_weight * 0.8:
            status = "underweight"
        elif cw > strat.target_weight * 1.2:
            status = "overweight"
        statuses.append(
            ConvictionStatus(
                type="strategy",
                name=name,
                thesis=strat.thesis,
                target_weight=strat.target_weight,
                current_weight=round(cw, 4),
                status=status,
            )
        )

    for cls, override in convictions.asset_classes.items():
        cw = class_weights.get(cls, 0.0)
        mid = (override.target_range[0] + override.target_range[1]) / 2
        status = "in_range"
        if cw < override.target_range[0]:
            status = "underweight"
        elif cw > override.target_range[1]:
            status = "overweight"
        statuses.append(
            ConvictionStatus(
                type="asset_class",
                name=cls,
                thesis=override.thesis,
                target_weight=mid,
                current_weight=round(cw, 4),
                status=status,
            )
        )

    return statuses


def find_consolidation_candidates(
    returns_df: pd.DataFrame,
    weights: dict[str, float],
    holdings: list,
    convictions: Convictions,
) -> list[ConsolidationRec]:
    if returns_df.empty or len(returns_df.columns) < 2:
        return []

    conviction_syms = set(convictions.positions.keys())
    holding_map = {h.symbol: h for h in holdings}
    corr = returns_df.corr()
    recs = []

    checked = set()
    for s1 in corr.columns:
        for s2 in corr.columns:
            if s1 >= s2:
                continue
            pair = (s1, s2)
            if pair in checked:
                continue
            checked.add(pair)

            h1 = holding_map.get(s1)
            h2 = holding_map.get(s2)
            if not h1 or not h2:
                continue
            if h1.asset_class != h2.asset_class:
                continue

            c = corr.loc[s1, s2]
            if abs(c) < 0.85:
                continue

            w1 = weights.get(s1, 0.0)
            w2 = weights.get(s2, 0.0)

            if s1 in conviction_syms and s2 in conviction_syms:
                continue
            if s1 in conviction_syms:
                smaller, larger = s2, s1
            elif s2 in conviction_syms:
                smaller, larger = s1, s2
            elif w1 >= w2:
                smaller, larger = s2, s1
            else:
                smaller, larger = s1, s2

            recs.append(
                ConsolidationRec(
                    symbol=smaller,
                    name=smaller,
                    current_weight=round(weights.get(smaller, 0.0), 4),
                    reason=f"Overlapping exposure (r={abs(c):.2f})",
                    merge_into=larger,
                )
            )

    recs.sort(key=lambda r: r.current_weight, reverse=True)
    return recs


def compute_gap_fill_deltas(
    holdings: list,
    target_weights: dict[str, float],
    total_value: float,
    convictions: Convictions,
) -> list[RebalanceDelta]:
    held = {h.symbol for h in holdings}
    deltas = []

    for name, sc in convictions.sectors.items():
        sector_weight = sum(
            target_weights.get(h.symbol, 0.0) for h in holdings if h.sector == name
        )
        gap = sc.target_weight - sector_weight
        if gap <= 0.01:
            continue
        candidates = [s for s in sc.instruments if s not in held]
        if not candidates:
            continue
        per_instrument = gap / len(candidates)
        for sym in candidates:
            dollars = per_instrument * total_value
            if dollars < 100:
                continue
            deltas.append(
                RebalanceDelta(
                    symbol=sym,
                    name=sym,
                    sector=name,
                    current_weight=0.0,
                    target_weight=round(per_instrument, 6),
                    delta_dollars=round(dollars, 2),
                    action="BUY",
                    best_account="taxable",
                    note="Gap Fill",
                )
            )

    for name, strat in convictions.strategies.items():
        strat_weight = sum(
            target_weights.get(h.symbol, 0.0)
            for h in holdings
            if h.symbol in strat.instruments
            or (strat.min_yield > 0 and h.dividend_yield >= strat.min_yield)
        )
        gap = strat.target_weight - strat_weight
        if gap <= 0.01:
            continue
        candidates = [s for s in strat.instruments if s not in held]
        if not candidates:
            continue
        per_instrument = gap / len(candidates)
        best_acct = strat.location if strat.location else "taxable"
        for sym in candidates:
            dollars = per_instrument * total_value
            if dollars < 100:
                continue
            deltas.append(
                RebalanceDelta(
                    symbol=sym,
                    name=sym,
                    sector=name,
                    current_weight=0.0,
                    target_weight=round(per_instrument, 6),
                    delta_dollars=round(dollars, 2),
                    action="BUY",
                    best_account=best_acct,
                    note="Gap Fill",
                )
            )

    for cls, override in convictions.asset_classes.items():
        class_weight = sum(
            target_weights.get(h.symbol, 0.0) for h in holdings if h.asset_class == cls
        )
        gap = override.target_range[0] - class_weight
        if gap <= 0.01:
            continue
        instruments = DEFAULT_GAP_FILL_INSTRUMENTS.get(cls, [])
        candidates = [s for s in instruments if s not in held]
        if not candidates:
            continue
        per_instrument = gap / len(candidates)
        for sym in candidates:
            dollars = per_instrument * total_value
            if dollars < 100:
                continue
            deltas.append(
                RebalanceDelta(
                    symbol=sym,
                    name=sym,
                    sector=cls,
                    current_weight=0.0,
                    target_weight=round(per_instrument, 6),
                    delta_dollars=round(dollars, 2),
                    action="BUY",
                    best_account="taxable",
                    note="Gap Fill",
                )
            )

    return deltas


def annotate_conviction_deltas(
    deltas: list[RebalanceDelta],
    convictions: Convictions,
    consolidation_recs: list[ConsolidationRec],
) -> None:
    conviction_syms = set(convictions.positions.keys())
    merge_targets = {r.symbol for r in consolidation_recs}

    for d in deltas:
        if d.note:
            continue
        if d.symbol in conviction_syms:
            d.note = "Conviction"
        elif d.symbol in merge_targets:
            d.note = "Consolidation"


class _EffectiveConstraints:
    """Wraps constraints with asset class overrides from convictions."""

    def __init__(self, base, ac_overrides: dict[str, list[float]]):
        self._base = base
        self._overrides = ac_overrides

    @property
    def max_position_weight(self):
        return self._base.max_position_weight

    @property
    def max_sector_weight(self):
        return self._base.max_sector_weight

    @property
    def target_equity_range(self):
        return self._overrides.get("equity", self._base.target_equity_range)

    @property
    def target_fixed_income_range(self):
        return self._overrides.get("fixed_income", self._base.target_fixed_income_range)

    @property
    def target_cash_range(self):
        return self._overrides.get("cash", self._base.target_cash_range)


def _cap_asset_classes_with_overrides(
    weights: dict[str, float],
    symbols_by_class: dict[str, list[str]],
    constraints,
    locked: set[str],
) -> dict[str, float]:
    class_ranges = {
        "equity": constraints.target_equity_range,
        "fixed_income": constraints.target_fixed_income_range,
        "cash": constraints.target_cash_range,
    }

    for cls, rng in class_ranges.items():
        members = symbols_by_class.get(cls, [])
        cls_weight = sum(weights.get(s, 0.0) for s in members)

        if cls_weight > rng[1] and cls_weight > 0:
            unlocked = [s for s in members if s not in locked]
            unlocked_weight = sum(weights.get(s, 0.0) for s in unlocked)
            locked_weight = cls_weight - unlocked_weight
            target_unlocked = max(rng[1] - locked_weight, 0)

            if unlocked_weight > 0:
                ratio = target_unlocked / unlocked_weight
                freed = 0.0
                for s in unlocked:
                    old = weights.get(s, 0.0)
                    weights[s] = old * ratio
                    freed += old - weights[s]

                others = [s for s in weights if s not in members]
                total_other = sum(weights.get(s, 0.0) for s in others)
                if total_other > 0:
                    for s in others:
                        weights[s] = weights[s] + freed * (weights[s] / total_other)

    return weights


async def _fetch_returns(market: "MarketService", symbols: list[str]) -> pd.DataFrame:
    end = date.today()
    start = end - timedelta(days=730)

    sem = asyncio.Semaphore(10)

    async def fetch_one(sym: str) -> tuple[str, list[dict]]:
        async with sem:
            try:
                points = await market.get_daily_chart(sym, start, end)
                return sym, points
            except Exception:
                logger.warning("failed to fetch history for %s", sym)
                return sym, []

    tasks = [fetch_one(s) for s in symbols]
    results = await asyncio.gather(*tasks)

    series = {}
    for sym, points in results:
        if len(points) < 30:
            continue
        prices = pd.Series(
            {p["timestamp"]: p["close"] for p in points if p.get("close") is not None}
        )
        if len(prices) < 30:
            continue
        prices = prices.sort_index()
        series[sym] = prices.pct_change().dropna()

    if not series:
        return pd.DataFrame()

    return pd.DataFrame(series).dropna()


def _cap_positions(
    weights: dict[str, float],
    cap: float,
    symbol_class: dict[str, str],
    symbols_by_class: dict[str, list[str]],
    locked: set[str] | None = None,
) -> dict[str, float]:
    locked = locked or set()
    excess = 0.0
    for sym, w in list(weights.items()):
        if sym in locked:
            continue
        if w > cap:
            excess += w - cap
            weights[sym] = cap

    if excess <= 0:
        return weights

    under_cap = {s: w for s, w in weights.items() if w < cap and s not in locked}
    total_under = sum(under_cap.values())
    if total_under > 0:
        for s, w in under_cap.items():
            weights[s] = w + excess * (w / total_under)

    return weights


def _cap_sectors(
    weights: dict[str, float],
    cap: float,
    symbol_sector: dict[str, str],
    locked: set[str] | None = None,
) -> dict[str, float]:
    locked = locked or set()
    sector_weights: dict[str, float] = defaultdict(float)
    for s, w in weights.items():
        sector_weights[symbol_sector.get(s, "")] += w

    for sector, sw in sector_weights.items():
        if sw <= cap or sw <= 0:
            continue
        members = [s for s, w in weights.items() if symbol_sector.get(s) == sector]
        unlocked_members = [s for s in members if s not in locked]
        unlocked_weight = sum(weights.get(s, 0.0) for s in unlocked_members)
        locked_weight = sw - unlocked_weight
        target_unlocked = max(cap - locked_weight, 0)

        if unlocked_weight > 0:
            ratio = target_unlocked / unlocked_weight
            freed = 0.0
            for s in unlocked_members:
                old = weights[s]
                weights[s] = old * ratio
                freed += old - weights[s]

            non_members = {
                s: w
                for s, w in weights.items()
                if symbol_sector.get(s) != sector and s not in locked
            }
            total_nm = sum(non_members.values())
            if total_nm > 0:
                for s, w in non_members.items():
                    weights[s] = w + freed * (w / total_nm)

    return weights


def _cap_asset_classes(
    weights: dict[str, float],
    symbols_by_class: dict[str, list[str]],
    constraints,
) -> dict[str, float]:
    class_ranges = {
        "equity": constraints.target_equity_range,
        "fixed_income": constraints.target_fixed_income_range,
        "cash": constraints.target_cash_range,
    }

    for cls, rng in class_ranges.items():
        members = symbols_by_class.get(cls, [])
        cls_weight = sum(weights.get(s, 0.0) for s in members)

        if cls_weight > rng[1] and cls_weight > 0:
            ratio = rng[1] / cls_weight
            freed = 0.0
            for s in members:
                old = weights.get(s, 0.0)
                weights[s] = old * ratio
                freed += old - weights[s]

            others = [s for s in weights if s not in members]
            total_other = sum(weights.get(s, 0.0) for s in others)
            if total_other > 0:
                for s in others:
                    weights[s] = weights[s] + freed * (weights[s] / total_other)

    return weights


def _best_account_for_trade(holding, action: str) -> str:
    ira_types = {"traditional_ira", "sep_ira"}
    if action == "SELL":
        if holding.account_type in ira_types:
            return "IRA (tax-free sell)"
        if holding.account_type == "roth_ira":
            return "Roth (tax-free)"
        return "Taxable (evaluate tax impact)"
    return holding.account_type
