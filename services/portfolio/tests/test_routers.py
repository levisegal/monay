"""Tests for Portfolio Service HTTP routers."""
import tempfile
from datetime import date
from pathlib import Path
from unittest.mock import AsyncMock, MagicMock

import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient

import sys
sys.path.insert(0, str(Path(__file__).parent.parent))

from routers import chart, portfolio, quotes
from services.cache import PriceCache
from services.holdings_client import HoldingsClient
from services.market import MarketService


@pytest.fixture
def mock_market():
    """Create a mock market service."""
    service = MagicMock(spec=MarketService)
    service.get_quotes = AsyncMock()
    service.get_intraday_chart = AsyncMock()
    service.get_daily_chart = AsyncMock()
    return service


@pytest.fixture
def mock_holdings():
    """Create a mock holdings client."""
    client = MagicMock(spec=HoldingsClient)
    client.get_holdings = AsyncMock()
    client.get_accounts = AsyncMock()
    return client


def create_test_app(market=None, holdings=None):
    """Create a FastAPI app with test routers."""
    app = FastAPI()
    if market:
        app.include_router(quotes.create_router(market), prefix="/api/v1")
        app.include_router(chart.create_router(market), prefix="/api/v1")
    if holdings and market:
        app.include_router(portfolio.create_router(holdings, market), prefix="/api/v1")
    return app


class TestQuotesRouter:
    def test_get_quotes_single_symbol(self, mock_market):
        mock_market.get_quotes.return_value = [
            {
                "symbol": "AAPL",
                "name": "Apple Inc.",
                "price": 175.50,
                "change": 2.30,
                "change_percent": 1.33,
                "previous_close": 173.20,
                "volume": 50000000,
                "asset_type": "equity",
            }
        ]

        app = create_test_app(market=mock_market)
        client = TestClient(app)

        response = client.get("/api/v1/quotes", params={"symbols": "AAPL"})

        assert response.status_code == 200
        data = response.json()
        assert len(data["quotes"]) == 1
        assert data["quotes"][0]["symbol"] == "AAPL"
        assert data["quotes"][0]["price"] == 175.50
        mock_market.get_quotes.assert_called_once_with(["AAPL"])

    def test_get_quotes_multiple_symbols(self, mock_market):
        mock_market.get_quotes.return_value = [
            {"symbol": "AAPL", "name": "Apple Inc.", "price": 175.50, "change": None, "change_percent": None, "previous_close": None, "volume": None, "asset_type": "equity"},
            {"symbol": "MSFT", "name": "Microsoft Corp", "price": 380.00, "change": None, "change_percent": None, "previous_close": None, "volume": None, "asset_type": "equity"},
        ]

        app = create_test_app(market=mock_market)
        client = TestClient(app)

        response = client.get("/api/v1/quotes", params={"symbols": "AAPL,MSFT"})

        assert response.status_code == 200
        data = response.json()
        assert len(data["quotes"]) == 2
        symbols = {q["symbol"] for q in data["quotes"]}
        assert symbols == {"AAPL", "MSFT"}
        mock_market.get_quotes.assert_called_once_with(["AAPL", "MSFT"])

    def test_get_quotes_normalizes_symbols(self, mock_market):
        mock_market.get_quotes.return_value = []

        app = create_test_app(market=mock_market)
        client = TestClient(app)

        response = client.get("/api/v1/quotes", params={"symbols": "aapl, msft ,  googl"})

        assert response.status_code == 200
        mock_market.get_quotes.assert_called_once_with(["AAPL", "MSFT", "GOOGL"])

    def test_get_quotes_missing_symbols(self, mock_market):
        app = create_test_app(market=mock_market)
        client = TestClient(app)

        response = client.get("/api/v1/quotes")

        assert response.status_code == 422


class TestChartRouter:
    def test_get_chart_daily(self, mock_market):
        mock_market.get_daily_chart.return_value = [
            {"timestamp": "2024-01-15", "open": 170.0, "high": 175.0, "low": 169.0, "close": 174.0, "volume": 1000000},
            {"timestamp": "2024-01-16", "open": 174.0, "high": 178.0, "low": 173.0, "close": 177.0, "volume": 1200000},
        ]

        app = create_test_app(market=mock_market)
        client = TestClient(app)

        response = client.get("/api/v1/chart/AAPL", params={"range": "1y"})

        assert response.status_code == 200
        data = response.json()
        assert data["symbol"] == "AAPL"
        assert len(data["points"]) == 2
        assert data["points"][0]["close"] == 174.0
        mock_market.get_daily_chart.assert_called_once()

    def test_get_chart_intraday(self, mock_market):
        mock_market.get_intraday_chart.return_value = [
            {"timestamp": "2024-01-15T09:30:00", "open": 170.0, "high": 171.0, "low": 169.5, "close": 170.5, "volume": 100000},
        ]

        app = create_test_app(market=mock_market)
        client = TestClient(app)

        response = client.get("/api/v1/chart/AAPL", params={"range": "1d"})

        assert response.status_code == 200
        data = response.json()
        assert data["symbol"] == "AAPL"
        mock_market.get_intraday_chart.assert_called_once()
        mock_market.get_daily_chart.assert_not_called()

    def test_get_chart_normalizes_symbol(self, mock_market):
        mock_market.get_daily_chart.return_value = []

        app = create_test_app(market=mock_market)
        client = TestClient(app)

        response = client.get("/api/v1/chart/aapl", params={"range": "1y"})

        assert response.status_code == 200
        data = response.json()
        assert data["symbol"] == "AAPL"


class TestPortfolioRouter:
    def test_get_enriched_portfolio_empty(self, mock_market, mock_holdings):
        mock_holdings.get_holdings.return_value = []

        app = create_test_app(market=mock_market, holdings=mock_holdings)
        client = TestClient(app)

        response = client.get("/api/v1/portfolio/enriched")

        assert response.status_code == 200
        data = response.json()
        assert data["summary"]["total_value"] == 0
        assert data["summary"]["holding_count"] == 0
        assert data["holdings"] == []

    def test_get_enriched_portfolio_with_holdings(self, mock_market, mock_holdings):
        mock_holdings.get_holdings.return_value = [
            {
                "account_id": "acct_123",
                "account_name": "Brokerage",
                "symbol": "AAPL",
                "security_name": "Apple Inc.",
                "quantity": 100.0,
                "cost_basis_micros": 15000_000_000,
            },
        ]
        mock_market.get_quotes.return_value = [
            {
                "symbol": "AAPL",
                "name": "Apple Inc.",
                "price": 175.0,
                "change": 2.0,
                "change_percent": 1.15,
                "previous_close": 173.0,
                "volume": 50000000,
                "asset_type": "equity",
            }
        ]

        app = create_test_app(market=mock_market, holdings=mock_holdings)
        client = TestClient(app)

        response = client.get("/api/v1/portfolio/enriched")

        assert response.status_code == 200
        data = response.json()
        assert data["summary"]["total_value"] == 17500.0
        assert data["summary"]["total_cost_basis"] == 15000.0
        assert data["summary"]["total_unrealized_gain"] == 2500.0
        assert data["summary"]["holding_count"] == 1

        holding = data["holdings"][0]
        assert holding["symbol"] == "AAPL"
        assert holding["quantity"] == 100.0
        assert holding["current_price"] == 175.0
        assert holding["market_value"] == 17500.0
        assert holding["unrealized_gain"] == 2500.0
        assert holding["day_change"] == 200.0

    def test_get_enriched_portfolio_multiple_holdings(self, mock_market, mock_holdings):
        mock_holdings.get_holdings.return_value = [
            {"account_id": "acct_1", "account_name": "Account 1", "symbol": "AAPL", "quantity": 50.0, "cost_basis_micros": 7500_000_000},
            {"account_id": "acct_1", "account_name": "Account 1", "symbol": "MSFT", "quantity": 30.0, "cost_basis_micros": 10500_000_000},
        ]
        mock_market.get_quotes.return_value = [
            {"symbol": "AAPL", "name": "Apple", "price": 175.0, "change": 2.0, "asset_type": "equity"},
            {"symbol": "MSFT", "name": "Microsoft", "price": 380.0, "change": -1.5, "asset_type": "equity"},
        ]

        app = create_test_app(market=mock_market, holdings=mock_holdings)
        client = TestClient(app)

        response = client.get("/api/v1/portfolio/enriched")

        assert response.status_code == 200
        data = response.json()

        aapl_value = 50.0 * 175.0
        msft_value = 30.0 * 380.0
        total_expected = aapl_value + msft_value

        assert data["summary"]["total_value"] == total_expected
        assert data["summary"]["holding_count"] == 2

        holdings_by_symbol = {h["symbol"]: h for h in data["holdings"]}
        assert holdings_by_symbol["AAPL"]["allocation_percent"] == pytest.approx((aapl_value / total_expected) * 100, rel=0.01)
        assert holdings_by_symbol["MSFT"]["allocation_percent"] == pytest.approx((msft_value / total_expected) * 100, rel=0.01)

    def test_get_enriched_portfolio_filters_by_account(self, mock_market, mock_holdings):
        mock_holdings.get_holdings.return_value = [
            {"account_id": "acct_specific", "account_name": "Specific Account", "symbol": "AAPL", "quantity": 100.0, "cost_basis_micros": 15000_000_000},
        ]
        mock_market.get_quotes.return_value = [
            {"symbol": "AAPL", "price": 175.0, "change": 0, "asset_type": "equity"},
        ]

        app = create_test_app(market=mock_market, holdings=mock_holdings)
        client = TestClient(app)

        response = client.get("/api/v1/portfolio/enriched", params={"account_id": "acct_specific"})

        assert response.status_code == 200
        mock_holdings.get_holdings.assert_called_once_with("acct_specific")

    def test_get_enriched_portfolio_handles_missing_price(self, mock_market, mock_holdings):
        mock_holdings.get_holdings.return_value = [
            {"account_id": "acct_1", "account_name": "Account", "symbol": "UNKNOWN", "quantity": 100.0, "cost_basis_micros": 5000_000_000},
        ]
        mock_market.get_quotes.return_value = []

        app = create_test_app(market=mock_market, holdings=mock_holdings)
        client = TestClient(app)

        response = client.get("/api/v1/portfolio/enriched")

        assert response.status_code == 200
        data = response.json()
        holding = data["holdings"][0]
        assert holding["current_price"] is None
        assert holding["market_value"] is None


class TestRangeToStartDate:
    """Property-based style tests for date range calculation."""

    def test_1mo_range(self):
        from routers.chart import _range_to_start_date
        end = date(2024, 6, 15)
        start = _range_to_start_date("1mo", end)
        assert (end - start).days == 30

    def test_1y_range(self):
        from routers.chart import _range_to_start_date
        end = date(2024, 6, 15)
        start = _range_to_start_date("1y", end)
        assert (end - start).days == 365

    def test_unknown_range_defaults_to_1y(self):
        from routers.chart import _range_to_start_date
        end = date(2024, 6, 15)
        start = _range_to_start_date("unknown", end)
        assert (end - start).days == 365
