"use client";

import { useState } from "react";
import { PortfolioChart } from "../components/charts/PortfolioChart";
import { HoldingDetailPanel } from "../components/holdings/HoldingDetailPanel";
import { GainsLossSection } from "../components/gains/GainsLossSection";
import { BondsSection } from "../components/bonds/BondsSection";
import { CashFlowSection } from "../components/cashflow/CashFlowSection";
import { RebalancingSection } from "../components/rebalancing/RebalancingSection";
import { Header } from "../components/nav/Header";
import { AccountSelector } from "../components/nav/AccountSelector";

// Import sample data
import portfolioHistory from "../data/portfolio/history.json";
import aaplQuote from "../data/quotes/AAPL.json";
import msftQuote from "../data/quotes/MSFT.json";
import vtiQuote from "../data/quotes/VTI.json";
import aaplChart from "../data/charts/AAPL.json";
import msftChart from "../data/charts/MSFT.json";
import vtiChart from "../data/charts/VTI.json";

import type { StockQuote, ChartDataPoint } from "../types/stock";

// Map symbols to data
const quotesMap: Record<string, StockQuote> = {
  AAPL: aaplQuote as StockQuote,
  MSFT: msftQuote as StockQuote,
  VTI: vtiQuote as StockQuote,
};

const chartsMap: Record<string, ChartDataPoint[]> = {
  AAPL: aaplChart.data as ChartDataPoint[],
  MSFT: msftChart.data as ChartDataPoint[],
  VTI: vtiChart.data as ChartDataPoint[],
};

// Mock portfolios
const mockPortfolios = [
  { id: "personal", name: "Personal Portfolio", totalValue: 284523 },
  { id: "family", name: "Family Trust", totalValue: 1245000 },
  { id: "retirement", name: "Retirement", totalValue: 523400 },
];

// Mock accounts with brokerage info
const mockAccounts = [
  {
    id: "schwab-1",
    name: "Individual",
    institution: "Charles Schwab",
    accountNumber: "4521",
    type: "Brokerage",
    balance: 156234,
  },
  {
    id: "fidelity-1",
    name: "Roth IRA",
    institution: "Fidelity",
    accountNumber: "8834",
    type: "IRA",
    balance: 89450,
  },
  {
    id: "vanguard-1",
    name: "401(k)",
    institution: "Vanguard",
    accountNumber: "2201",
    type: "401k",
    balance: 38839,
  },
];

// Holdings with account association (same symbol can be in multiple accounts)
const allHoldings = [
  {
    name: "Apple Inc.",
    symbol: "AAPL",
    shares: 150,
    price: 178.50,
    value: 26775,
    costBasis: 22500,
    allocation: "9.4%",
    change: 2.4,
    changePercent: 2.4,
    positive: true,
    accountId: "schwab-1",
  },
  {
    name: "Apple Inc.",
    symbol: "AAPL",
    shares: 90,
    price: 178.50,
    value: 16065,
    costBasis: 13500,
    allocation: "5.6%",
    change: 2.4,
    changePercent: 2.4,
    positive: true,
    accountId: "fidelity-1",
  },
  {
    name: "Microsoft Corporation",
    symbol: "MSFT",
    shares: 180,
    price: 378.91,
    value: 68204,
    costBasis: 54000,
    allocation: "24.0%",
    change: 1.8,
    changePercent: 1.8,
    positive: true,
    accountId: "schwab-1",
  },
  {
    name: "Vanguard Total Stock Market ETF",
    symbol: "VTI",
    shares: 200,
    price: 242.15,
    value: 48430,
    costBasis: 40000,
    allocation: "17.0%",
    change: 1.2,
    changePercent: 1.2,
    positive: true,
    accountId: "fidelity-1",
  },
  {
    name: "Vanguard Total Stock Market ETF",
    symbol: "VTI",
    shares: 150,
    price: 242.15,
    value: 36323,
    costBasis: 30000,
    allocation: "12.8%",
    change: 1.2,
    changePercent: 1.2,
    positive: true,
    accountId: "vanguard-1",
  },
  {
    name: "iShares Core U.S. Aggregate Bond ETF",
    symbol: "AGG",
    shares: 420,
    price: 98.45,
    value: 41349,
    costBasis: 42000,
    allocation: "14.5%",
    change: -0.3,
    changePercent: -0.3,
    positive: false,
    accountId: "vanguard-1",
  },
  {
    name: "Amazon.com Inc.",
    symbol: "AMZN",
    shares: 95,
    price: 178.25,
    value: 16934,
    costBasis: 14000,
    allocation: "6.0%",
    change: 3.1,
    changePercent: 3.1,
    positive: true,
    accountId: "schwab-1",
  },
  {
    name: "Alphabet Inc. Class A",
    symbol: "GOOGL",
    shares: 110,
    price: 141.80,
    value: 15598,
    costBasis: 13000,
    allocation: "5.5%",
    change: 1.5,
    changePercent: 1.5,
    positive: true,
    accountId: "fidelity-1",
  },
];

export default function DashboardPage() {
  const [selectedHolding, setSelectedHolding] = useState<string | null>(null);
  const [selectedPortfolio, setSelectedPortfolio] = useState("personal");
  const [selectedAccount, setSelectedAccount] = useState<string | null>(null);

  // Filter holdings by selected account, then aggregate by symbol
  const filteredHoldings = selectedAccount
    ? allHoldings.filter((h) => h.accountId === selectedAccount)
    : allHoldings;

  // Aggregate holdings by symbol
  const holdings = Object.values(
    filteredHoldings.reduce((acc, h) => {
      if (!acc[h.symbol]) {
        acc[h.symbol] = { ...h, accountCount: 1 };
      } else {
        acc[h.symbol].shares += h.shares;
        acc[h.symbol].value += h.value;
        acc[h.symbol].costBasis += h.costBasis;
        acc[h.symbol].accountCount += 1;
      }
      return acc;
    }, {} as Record<string, typeof filteredHoldings[0] & { accountCount: number }>)
  ).map((h) => ({
    ...h,
    // Recalculate allocation based on total
    allocation: `${((h.value / 284523) * 100).toFixed(1)}%`,
  }));

  // Get account info for display
  const getAccountInfo = (accountId: string) =>
    mockAccounts.find((a) => a.id === accountId);

  const stats = [
    {
      label: "Total Value",
      value: "$284,523",
      change: "+18.4% YTD",
      positive: true,
    },
    {
      label: "Holdings",
      value: "12",
      change: "+2 this month",
      positive: true,
    },
    {
      label: "Today's Change",
      value: "+$1,247",
      change: "+0.44%",
      positive: true,
    },
    {
      label: "Cash",
      value: "$12,840",
      change: "4.5% of portfolio",
      positive: null,
    },
  ];

  // Get all holdings with the selected symbol (across all accounts)
  const currentHoldings = selectedHolding
    ? allHoldings.filter((h) => h.symbol === selectedHolding)
    : [];
  const currentQuote = selectedHolding ? quotesMap[selectedHolding] : null;
  const currentChartData = selectedHolding ? chartsMap[selectedHolding] : [];

  const formatCurrency = (value: number) =>
    new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: "USD",
      minimumFractionDigits: 2,
    }).format(value);

  return (
    <div className="min-h-screen">
      {/* Top Header with Portfolio Switcher */}
      <Header
        portfolios={mockPortfolios}
        selectedPortfolioId={selectedPortfolio}
        onPortfolioSelect={setSelectedPortfolio}
      />

      <div className="max-w-7xl mx-auto px-4 py-6 sm:p-8">
        {/* Page Header */}
        <header className="mb-6">
          <h1 className="font-serif text-3xl font-semibold text-ink mb-2">
            Portfolio Statement
          </h1>
          <p className="text-base text-foreground-secondary mb-4">
            {new Date().toLocaleDateString("en-US", {
              month: "long",
              day: "numeric",
              year: "numeric",
            })}
          </p>
          {/* Account Selector */}
          <AccountSelector
            accounts={mockAccounts}
            selectedId={selectedAccount}
            onSelect={setSelectedAccount}
          />
        </header>

        {/* Stat Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-section">
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

        {/* Portfolio Chart */}
        <div className="mb-section">
          <PortfolioChart data={portfolioHistory.data} />
        </div>

        {/* Holdings */}
        <div className="bg-white border border-paper-gray rounded-md shadow-paper p-4 sm:p-statement">
          <div className="flex justify-between items-center mb-6">
            <h2 className="font-serif text-lg font-semibold text-ink">
              Holdings
            </h2>
            <button className="px-4 sm:px-6 py-2 border border-paper-gray text-ink rounded-sm hover:bg-paper-gray transition-all font-sans font-semibold text-sm">
              Import CSV
            </button>
          </div>

          {/* Mobile Card Layout */}
          <div className="md:hidden space-y-4">
            {holdings.map((holding) => (
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
                        <span className="text-foreground-secondary/60"> · {holding.accountCount} accounts</span>
                      )}
                    </div>
                  </div>
                  <div
                    className={`font-semibold ${
                      holding.positive ? "text-green-700" : "text-red-700"
                    }`}
                  >
                    {holding.positive ? "+" : ""}
                    {holding.changePercent}%
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
                      {holding.allocation}
                    </div>
                  </div>
                  <div>
                    <div className="text-foreground-secondary">Shares</div>
                    <div className="text-ink">{holding.shares}</div>
                  </div>
                  <div>
                    <div className="text-foreground-secondary">Price</div>
                    <div className="text-ink">
                      {formatCurrency(holding.price)}
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* Desktop Table Layout */}
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
                {holdings.map((holding) => (
                  <tr
                    key={holding.symbol}
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
                          <span className="text-foreground-secondary/60"> · {holding.accountCount} accounts</span>
                        )}
                      </div>
                    </td>
                    <td className="py-3 text-right text-foreground-secondary">
                      {holding.shares}
                    </td>
                    <td className="py-3 text-right text-foreground-secondary">
                      {formatCurrency(holding.price)}
                    </td>
                    <td className="py-3 text-right font-serif font-semibold text-ink">
                      {formatCurrency(holding.value)}
                    </td>
                    <td className="py-3 text-right text-foreground-secondary">
                      {holding.allocation}
                    </td>
                    <td
                      className={`py-3 text-right font-semibold ${
                        holding.positive ? "text-green-700" : "text-red-700"
                      }`}
                    >
                      {holding.positive ? "+" : ""}
                      {holding.changePercent}%
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Gains & Losses Section */}
        <div className="mt-section">
          <GainsLossSection />
        </div>

        {/* Fixed Income Section */}
        <div className="mt-section">
          <BondsSection />
        </div>

        {/* Cash Flow Section */}
        <div className="mt-section">
          <CashFlowSection />
        </div>

        {/* Rebalancing Section */}
        <div className="mt-section">
          <RebalancingSection />
        </div>
      </div>

      {/* Holding Detail Panel */}
      <HoldingDetailPanel
        holdings={currentHoldings.map((h) => ({
          name: h.name,
          symbol: h.symbol,
          shares: h.shares,
          price: h.price,
          value: h.value,
          costBasis: h.costBasis,
          change: h.change,
          changePercent: h.changePercent,
          accountName: getAccountInfo(h.accountId)?.name,
          accountInstitution: getAccountInfo(h.accountId)?.institution,
        }))}
        quote={currentQuote || null}
        chartData={currentChartData || []}
        isOpen={selectedHolding !== null}
        onClose={() => setSelectedHolding(null)}
      />
    </div>
  );
}
