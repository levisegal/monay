"use client";

import { useState } from "react";
import { PortfolioChart } from "../components/charts/PortfolioChart";
import { HoldingDetailPanel } from "../components/holdings/HoldingDetailPanel";
import { GainsLossSection } from "../components/gains/GainsLossSection";

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

export default function DashboardPage() {
  const [selectedHolding, setSelectedHolding] = useState<string | null>(null);

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

  const holdings = [
    {
      name: "Apple Inc.",
      symbol: "AAPL",
      shares: 240,
      price: 178.50,
      value: 42840,
      costBasis: 36000,
      allocation: "15.1%",
      change: 2.4,
      changePercent: 2.4,
      positive: true,
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
    },
    {
      name: "Vanguard Total Stock Market ETF",
      symbol: "VTI",
      shares: 350,
      price: 242.15,
      value: 84753,
      costBasis: 70000,
      allocation: "29.8%",
      change: 1.2,
      changePercent: 1.2,
      positive: true,
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
    },
  ];

  const currentHolding = holdings.find((h) => h.symbol === selectedHolding);
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
      <div className="max-w-7xl mx-auto px-4 py-6 sm:p-8">
        {/* Page Header */}
        <header className="mb-statement">
          <h1 className="font-serif text-3xl font-semibold text-ink mb-2">
            Portfolio Statement
          </h1>
          <p className="text-base text-foreground-secondary">
            Investment Account -{" "}
            {new Date().toLocaleDateString("en-US", {
              month: "long",
              day: "numeric",
              year: "numeric",
            })}
          </p>
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
      </div>

      {/* Holding Detail Panel */}
      <HoldingDetailPanel
        holding={
          currentHolding
            ? {
                name: currentHolding.name,
                symbol: currentHolding.symbol,
                shares: currentHolding.shares,
                price: currentHolding.price,
                value: currentHolding.value,
                costBasis: currentHolding.costBasis,
                change: currentHolding.change,
                changePercent: currentHolding.changePercent,
              }
            : null
        }
        quote={currentQuote || null}
        chartData={currentChartData || []}
        isOpen={selectedHolding !== null}
        onClose={() => setSelectedHolding(null)}
      />
    </div>
  );
}
