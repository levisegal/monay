"use client";

import { useState, useEffect } from "react";
import { PortfolioChart } from "../components/charts/PortfolioChart";
import { HoldingDetailPanel } from "../components/holdings/HoldingDetailPanel";
import { GainsLossSection } from "../components/gains/GainsLossSection";
import { BondsSection } from "../components/bonds/BondsSection";
import { CashFlowSection } from "../components/cashflow/CashFlowSection";
import { RebalancingSection } from "../components/rebalancing/RebalancingSection";
import { Header } from "../components/nav/Header";
import { AccountSelector } from "../components/nav/AccountSelector";

import { useGetEnrichedPortfolioApiV1PortfolioEnrichedGet } from "../lib/api/portfolio/portfolio";
import { useGetChartApiV1ChartSymbolGet } from "../lib/api/chart/chart";
import { useGetQuotesApiV1QuotesGet } from "../lib/api/quotes/quotes";
import type {
  EnrichedPortfolio,
  EnrichedHolding,
  PortfolioSummary,
  ChartResponse,
  QuotesResponse,
} from "../lib/api/portfolio.schemas";

import portfolioHistory from "../data/portfolio/history.json";
import type { StockQuote, ChartDataPoint } from "../types/stock";

const mockPortfolios = [
  { id: "personal", name: "Personal Portfolio", totalValue: 284523 },
  { id: "family", name: "Family Trust", totalValue: 1245000 },
  { id: "retirement", name: "Retirement", totalValue: 523400 },
];

interface Account {
  id: string;
  name: string;
  institution: string;
  accountNumber: string;
  type: string;
  balance: number;
}

interface AggregatedHolding {
  symbol: string;
  name: string;
  shares: number;
  price: number;
  value: number;
  costBasis: number;
  change: number;
  changePercent: number;
  allocation: number;
  accountCount: number;
  positive: boolean;
}

export default function DashboardPage() {
  const [selectedHolding, setSelectedHolding] = useState<string | null>(null);
  const [selectedPortfolio, setSelectedPortfolio] = useState("personal");
  const [selectedAccount, setSelectedAccount] = useState<string | null>(null);
  const [currentDate, setCurrentDate] = useState<string>("");

  useEffect(() => {
    setCurrentDate(
      new Date().toLocaleDateString("en-US", {
        month: "long",
        day: "numeric",
        year: "numeric",
      })
    );
  }, []);

  const { data: portfolioResponse, isLoading, error } = useGetEnrichedPortfolioApiV1PortfolioEnrichedGet(
    selectedAccount ? { account_id: selectedAccount } : undefined
  );

  const { data: chartResponse } = useGetChartApiV1ChartSymbolGet(
    selectedHolding || "",
    { range: "1y" },
    { query: { enabled: !!selectedHolding } }
  );

  const { data: quoteResponse } = useGetQuotesApiV1QuotesGet(
    { symbols: selectedHolding || "" },
    { query: { enabled: !!selectedHolding } }
  );

  const portfolio: EnrichedPortfolio | null =
    portfolioResponse?.status === 200 ? portfolioResponse.data : null;
  const summary: PortfolioSummary | undefined = portfolio?.summary;
  const holdings: EnrichedHolding[] = portfolio?.holdings || [];

  const chartData: ChartResponse | null =
    chartResponse?.status === 200 ? chartResponse.data : null;

  const quotesData: QuotesResponse | null =
    quoteResponse?.status === 200 ? quoteResponse.data : null;

  const uniqueAccounts: Account[] = Array.from(
    new Map(
      holdings.map((h: EnrichedHolding) => [
        h.account_id,
        {
          id: h.account_id,
          name: h.account_name,
          institution: "Holdings",
          accountNumber: "",
          type: "Brokerage",
          balance: 0,
        },
      ])
    ).values()
  );

  const aggregatedHoldings: AggregatedHolding[] = Object.values(
    holdings.reduce(
      (acc: Record<string, AggregatedHolding>, h: EnrichedHolding) => {
        if (!acc[h.symbol]) {
          acc[h.symbol] = {
            symbol: h.symbol,
            name: h.name || h.symbol,
            shares: h.quantity,
            price: h.current_price || 0,
            value: h.market_value || 0,
            costBasis: h.cost_basis || 0,
            change: h.day_change || 0,
            changePercent: h.unrealized_gain_percent || 0,
            allocation: h.allocation_percent || 0,
            accountCount: 1,
            positive: (h.day_change || 0) >= 0,
          };
        } else {
          acc[h.symbol].shares += h.quantity;
          acc[h.symbol].value += h.market_value || 0;
          acc[h.symbol].costBasis += h.cost_basis || 0;
          acc[h.symbol].change += h.day_change || 0;
          acc[h.symbol].accountCount += 1;
        }
        return acc;
      },
      {}
    )
  );

  const stats = [
    {
      label: "Total Value",
      value: formatCurrency(summary?.total_value || 0),
      change: `${summary?.total_unrealized_gain && summary.total_unrealized_gain >= 0 ? "+" : ""}${formatCurrency(summary?.total_unrealized_gain || 0)} total`,
      positive: (summary?.total_unrealized_gain || 0) >= 0,
    },
    {
      label: "Holdings",
      value: String(summary?.holding_count || 0),
      change: `${aggregatedHoldings.length} positions`,
      positive: true,
    },
    {
      label: "Today's Change",
      value: `${summary?.day_change && summary.day_change >= 0 ? "+" : ""}${formatCurrency(summary?.day_change || 0)}`,
      change: `${summary?.day_change_percent && summary.day_change_percent >= 0 ? "+" : ""}${(summary?.day_change_percent || 0).toFixed(2)}%`,
      positive: (summary?.day_change || 0) >= 0,
    },
    {
      label: "Cost Basis",
      value: formatCurrency(summary?.total_cost_basis || 0),
      change: "Total invested",
      positive: null,
    },
  ];

  const currentHoldings: EnrichedHolding[] = selectedHolding
    ? holdings.filter((h: EnrichedHolding) => h.symbol === selectedHolding)
    : [];

  const quote = quotesData?.quotes?.[0];
  const currentQuote: StockQuote | null = quote
    ? {
        symbol: quote.symbol,
        shortName: quote.name || quote.symbol,
        longName: quote.name || quote.symbol,
        regularMarketPrice: quote.price || 0,
        regularMarketChange: quote.change || 0,
        regularMarketChangePercent: quote.change_percent || 0,
        regularMarketPreviousClose: quote.previous_close || 0,
        regularMarketOpen: quote.price || 0,
        regularMarketDayHigh: quote.price || 0,
        regularMarketDayLow: quote.price || 0,
        regularMarketVolume: quote.volume || 0,
        fiftyTwoWeekHigh: quote.price || 0,
        fiftyTwoWeekLow: quote.price || 0,
        marketCap: 0,
        trailingPE: null,
        dividendYield: null,
        currency: "USD",
        exchange: "NASDAQ",
        quoteType: (quote.asset_type === "etf" ? "ETF" : "EQUITY") as "EQUITY" | "ETF" | "MUTUALFUND",
      }
    : null;

  const currentChartData: ChartDataPoint[] =
    chartData?.points?.map((p, idx) => ({
      timestamp: idx,
      date: p.timestamp,
      open: p.open || 0,
      high: p.high || 0,
      low: p.low || 0,
      close: p.close || 0,
      volume: p.volume || 0,
    })) || [];

  const getAccountInfo = (accountId: string): Account | undefined =>
    uniqueAccounts.find((a) => a.id === accountId);

  return (
    <div className="min-h-screen">
      <Header
        portfolios={mockPortfolios}
        selectedPortfolioId={selectedPortfolio}
        onPortfolioSelect={setSelectedPortfolio}
      />

      <div className="max-w-7xl mx-auto px-4 py-6 sm:p-8">
        <header className="mb-6">
          <h1 className="font-serif text-3xl font-semibold text-ink mb-2">
            Portfolio Statement
          </h1>
          <p className="text-base text-foreground-secondary mb-4">
            {currentDate}
          </p>
          <AccountSelector
            accounts={uniqueAccounts}
            selectedId={selectedAccount}
            onSelect={setSelectedAccount}
          />
        </header>

        {isLoading && (
          <div data-testid="loading-state" className="text-center py-8 text-foreground-secondary">
            Loading portfolio data...
          </div>
        )}

        {error && (
          <div data-testid="error-state" className="text-center py-8 text-red-700">
            Failed to load portfolio. Make sure Holdings and Portfolio services are running.
          </div>
        )}

        {!isLoading && !error && (
          <>
            <div data-testid="stats-cards" className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-section">
              {stats.map((stat) => (
                <div
                  key={stat.label}
                  className="bg-white border border-paper-gray rounded-md shadow-paper p-6"
                >
                  <div className="text-sm font-serif text-accent-blue mb-2">
                    {stat.label}
                  </div>
                  <div className="font-serif text-xl font-semibold text-ink mb-1">
                    {stat.value}
                  </div>
                  <div
                    className={`text-sm ${
                      stat.positive === true
                        ? "text-green-700"
                        : stat.positive === false
                          ? "text-red-700"
                          : "text-foreground-secondary"
                    }`}
                  >
                    {stat.change}
                  </div>
                </div>
              ))}
            </div>

            <div className="mb-section">
              <PortfolioChart data={portfolioHistory.data} />
            </div>

            <div data-testid="holdings-table" className="bg-white border border-paper-gray rounded-md shadow-paper p-4 sm:p-statement">
              <div className="flex justify-between items-center mb-6">
                <h2 className="font-serif text-lg font-semibold text-ink">
                  Holdings
                </h2>
                <button className="px-4 sm:px-6 py-2 border border-paper-gray text-ink rounded-sm hover:bg-paper-gray transition-all font-sans font-semibold text-sm">
                  Import CSV
                </button>
              </div>

              <div className="md:hidden space-y-4">
                {aggregatedHoldings.map((holding) => (
                  <div
                    key={holding.symbol}
                    onClick={() => setSelectedHolding(holding.symbol)}
                    className="border border-paper-gray rounded-md p-4 cursor-pointer hover:shadow-paper-hover transition-all"
                  >
                    <div className="flex justify-between items-start mb-3">
                      <div>
                        <div className="font-semibold text-ink">{holding.name}</div>
                        <div className="text-sm text-foreground-secondary">
                          {holding.symbol}
                          {holding.accountCount > 1 && (
                            <span className="text-foreground-secondary/60">
                              {" "}
                              · {holding.accountCount} accounts
                            </span>
                          )}
                        </div>
                      </div>
                      <div
                        className={`font-semibold ${
                          holding.positive ? "text-green-700" : "text-red-700"
                        }`}
                      >
                        {holding.positive ? "+" : ""}
                        {holding.changePercent.toFixed(2)}%
                      </div>
                    </div>
                    <div className="grid grid-cols-2 gap-3 text-sm">
                      <div>
                        <div className="text-foreground-secondary">Value</div>
                        <div className="font-serif font-semibold text-ink">
                          {formatCurrency(holding.value)}
                        </div>
                      </div>
                      <div>
                        <div className="text-foreground-secondary">Allocation</div>
                        <div className="font-semibold text-ink">
                          {holding.allocation.toFixed(1)}%
                        </div>
                      </div>
                      <div>
                        <div className="text-foreground-secondary">Shares</div>
                        <div className="text-ink">{holding.shares.toFixed(2)}</div>
                      </div>
                      <div>
                        <div className="text-foreground-secondary">Price</div>
                        <div className="text-ink">{formatCurrency(holding.price)}</div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>

              <div className="hidden md:block overflow-x-auto">
                <table className="w-full">
                  <thead className="border-b-2 border-ink">
                    <tr className="text-left">
                      <th className="pb-3 font-serif font-semibold text-ink">
                        Security
                      </th>
                      <th className="pb-3 font-serif font-semibold text-ink text-right">
                        Shares
                      </th>
                      <th className="pb-3 font-serif font-semibold text-ink text-right">
                        Price
                      </th>
                      <th className="pb-3 font-serif font-semibold text-ink text-right">
                        Value
                      </th>
                      <th className="pb-3 font-serif font-semibold text-ink text-right">
                        Allocation
                      </th>
                      <th className="pb-3 font-serif font-semibold text-ink text-right">
                        Change
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {aggregatedHoldings.map((holding) => (
                      <tr
                        key={holding.symbol}
                        data-testid="holding-row"
                        onClick={() => setSelectedHolding(holding.symbol)}
                        className="border-b border-paper-gray hover:bg-paper-cream/50 transition-all cursor-pointer"
                      >
                        <td className="py-3">
                          <div className="font-semibold text-ink">
                            {holding.name}
                          </div>
                          <div className="text-sm text-foreground-secondary">
                            {holding.symbol}
                            {holding.accountCount > 1 && (
                              <span className="text-foreground-secondary/60">
                                {" "}
                                · {holding.accountCount} accounts
                              </span>
                            )}
                          </div>
                        </td>
                        <td className="py-3 text-right text-foreground-secondary">
                          {holding.shares.toFixed(2)}
                        </td>
                        <td className="py-3 text-right text-foreground-secondary">
                          {formatCurrency(holding.price)}
                        </td>
                        <td className="py-3 text-right font-serif font-semibold text-ink">
                          {formatCurrency(holding.value)}
                        </td>
                        <td className="py-3 text-right text-foreground-secondary">
                          {holding.allocation.toFixed(1)}%
                        </td>
                        <td
                          className={`py-3 text-right font-semibold ${
                            holding.positive ? "text-green-700" : "text-red-700"
                          }`}
                        >
                          {holding.positive ? "+" : ""}
                          {holding.changePercent.toFixed(2)}%
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              {aggregatedHoldings.length === 0 && (
                <div data-testid="empty-holdings" className="text-center py-8 text-foreground-secondary">
                  No holdings found. Import a CSV to get started.
                </div>
              )}
            </div>

            <div className="mt-section">
              <GainsLossSection />
            </div>

            <div className="mt-section">
              <BondsSection />
            </div>

            <div className="mt-section">
              <CashFlowSection />
            </div>

            <div className="mt-section">
              <RebalancingSection />
            </div>
          </>
        )}
      </div>

      <HoldingDetailPanel
        holdings={currentHoldings.map((h: EnrichedHolding) => ({
          name: h.name || h.symbol,
          symbol: h.symbol,
          shares: h.quantity,
          price: h.current_price || 0,
          value: h.market_value || 0,
          costBasis: h.cost_basis || 0,
          change: h.day_change || 0,
          changePercent: h.unrealized_gain_percent || 0,
          accountName: getAccountInfo(h.account_id)?.name,
          accountInstitution: getAccountInfo(h.account_id)?.institution,
        }))}
        quote={currentQuote}
        chartData={currentChartData}
        isOpen={selectedHolding !== null}
        onClose={() => setSelectedHolding(null)}
      />
    </div>
  );
}

function formatCurrency(value: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    minimumFractionDigits: 2,
  }).format(value);
}
