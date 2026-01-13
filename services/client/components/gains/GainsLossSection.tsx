"use client";

import { useState } from "react";
import { formatCurrency, formatPercent } from "../../lib/formatters";

interface TaxLot {
  purchaseDate: string;
  shares: number;
  costBasis: number;
  currentValue: number;
  gainLoss: number;
  gainLossPercent: number;
  isLongTerm: boolean; // held > 1 year
}

interface PositionGains {
  symbol: string;
  name: string;
  totalShares: number;
  totalCostBasis: number;
  totalCurrentValue: number;
  shortTermGain: number;
  longTermGain: number;
  lots: TaxLot[];
}

// Mock data - in reality this would come from the backend
const mockPositionGains: PositionGains[] = [
  {
    symbol: "AAPL",
    name: "Apple Inc.",
    totalShares: 240,
    totalCostBasis: 36000,
    totalCurrentValue: 42840,
    shortTermGain: 2840,
    longTermGain: 4000,
    lots: [
      {
        purchaseDate: "2023-03-15",
        shares: 100,
        costBasis: 15000,
        currentValue: 17850,
        gainLoss: 2850,
        gainLossPercent: 19.0,
        isLongTerm: true,
      },
      {
        purchaseDate: "2023-08-20",
        shares: 80,
        costBasis: 12000,
        currentValue: 14280,
        gainLoss: 2280,
        gainLossPercent: 19.0,
        isLongTerm: true,
      },
      {
        purchaseDate: "2024-06-10",
        shares: 60,
        costBasis: 9000,
        currentValue: 10710,
        gainLoss: 1710,
        gainLossPercent: 19.0,
        isLongTerm: false,
      },
    ],
  },
  {
    symbol: "MSFT",
    name: "Microsoft Corporation",
    totalShares: 180,
    totalCostBasis: 54000,
    totalCurrentValue: 68204,
    shortTermGain: 5204,
    longTermGain: 9000,
    lots: [
      {
        purchaseDate: "2022-11-01",
        shares: 100,
        costBasis: 30000,
        currentValue: 37891,
        gainLoss: 7891,
        gainLossPercent: 26.3,
        isLongTerm: true,
      },
      {
        purchaseDate: "2024-04-15",
        shares: 80,
        costBasis: 24000,
        currentValue: 30313,
        gainLoss: 6313,
        gainLossPercent: 26.3,
        isLongTerm: false,
      },
    ],
  },
  {
    symbol: "VTI",
    name: "Vanguard Total Stock Market ETF",
    totalShares: 350,
    totalCostBasis: 70000,
    totalCurrentValue: 84753,
    shortTermGain: 0,
    longTermGain: 14753,
    lots: [
      {
        purchaseDate: "2021-06-01",
        shares: 200,
        costBasis: 40000,
        currentValue: 48430,
        gainLoss: 8430,
        gainLossPercent: 21.1,
        isLongTerm: true,
      },
      {
        purchaseDate: "2022-12-15",
        shares: 150,
        costBasis: 30000,
        currentValue: 36323,
        gainLoss: 6323,
        gainLossPercent: 21.1,
        isLongTerm: true,
      },
    ],
  },
  {
    symbol: "AGG",
    name: "iShares Core U.S. Aggregate Bond ETF",
    totalShares: 420,
    totalCostBasis: 42000,
    totalCurrentValue: 41349,
    shortTermGain: -651,
    longTermGain: 0,
    lots: [
      {
        purchaseDate: "2024-08-01",
        shares: 420,
        costBasis: 42000,
        currentValue: 41349,
        gainLoss: -651,
        gainLossPercent: -1.6,
        isLongTerm: false,
      },
    ],
  },
  {
    symbol: "AMZN",
    name: "Amazon.com Inc.",
    totalShares: 95,
    totalCostBasis: 14000,
    totalCurrentValue: 16934,
    shortTermGain: 1434,
    longTermGain: 1500,
    lots: [
      {
        purchaseDate: "2023-01-20",
        shares: 50,
        costBasis: 7500,
        currentValue: 8913,
        gainLoss: 1413,
        gainLossPercent: 18.8,
        isLongTerm: true,
      },
      {
        purchaseDate: "2024-09-05",
        shares: 45,
        costBasis: 6500,
        currentValue: 8021,
        gainLoss: 1521,
        gainLossPercent: 23.4,
        isLongTerm: false,
      },
    ],
  },
  {
    symbol: "GOOGL",
    name: "Alphabet Inc. Class A",
    totalShares: 110,
    totalCostBasis: 13000,
    totalCurrentValue: 15598,
    shortTermGain: 0,
    longTermGain: 2598,
    lots: [
      {
        purchaseDate: "2023-02-14",
        shares: 110,
        costBasis: 13000,
        currentValue: 15598,
        gainLoss: 2598,
        gainLossPercent: 20.0,
        isLongTerm: true,
      },
    ],
  },
];

export function GainsLossSection() {
  const [expandedPosition, setExpandedPosition] = useState<string | null>(null);
  const [showBreakdown, setShowBreakdown] = useState(false);

  // Calculate totals
  const totalShortTermGain = mockPositionGains.reduce(
    (sum, p) => sum + p.shortTermGain,
    0
  );
  const totalLongTermGain = mockPositionGains.reduce(
    (sum, p) => sum + p.longTermGain,
    0
  );
  const totalGain = totalShortTermGain + totalLongTermGain;

  const togglePosition = (symbol: string) => {
    setExpandedPosition(expandedPosition === symbol ? null : symbol);
  };

  return (
    <div className="bg-white border border-paper-gray rounded-md shadow-paper p-4 sm:p-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="font-serif text-lg font-semibold text-ink">
            Unrealized Gains & Losses
          </h2>
          <p className="text-sm text-foreground-secondary">
            Tax year 2025 · As of today
          </p>
        </div>
        <button
          onClick={() => setShowBreakdown(!showBreakdown)}
          className="px-4 py-2 text-sm font-medium text-accent-blue hover:text-ink transition-colors"
        >
          {showBreakdown ? "Hide Breakdown" : "Show Breakdown"}
        </button>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        {/* Total Gain */}
        <div className="bg-paper-cream/50 border border-paper-gray rounded-md p-4">
          <div className="text-sm text-foreground-secondary mb-1">
            Total Unrealized
          </div>
          <div
            className={`font-serif text-2xl font-semibold ${
              totalGain >= 0 ? "text-green-700" : "text-red-700"
            }`}
          >
            {totalGain >= 0 ? "+" : ""}
            {formatCurrency(totalGain)}
          </div>
        </div>

        {/* Short-Term */}
        <div className="bg-paper-cream/50 border border-paper-gray rounded-md p-4">
          <div className="flex items-center gap-2 mb-1">
            <span className="text-sm text-foreground-secondary">Short-Term</span>
            <span className="text-xs px-1.5 py-0.5 bg-amber-100 text-amber-800 rounded">
              &lt; 1 year
            </span>
          </div>
          <div
            className={`font-serif text-2xl font-semibold ${
              totalShortTermGain >= 0 ? "text-green-700" : "text-red-700"
            }`}
          >
            {totalShortTermGain >= 0 ? "+" : ""}
            {formatCurrency(totalShortTermGain)}
          </div>
          <div className="text-xs text-foreground-secondary mt-1">
            Taxed as ordinary income
          </div>
        </div>

        {/* Long-Term */}
        <div className="bg-paper-cream/50 border border-paper-gray rounded-md p-4">
          <div className="flex items-center gap-2 mb-1">
            <span className="text-sm text-foreground-secondary">Long-Term</span>
            <span className="text-xs px-1.5 py-0.5 bg-green-100 text-green-800 rounded">
              &gt; 1 year
            </span>
          </div>
          <div
            className={`font-serif text-2xl font-semibold ${
              totalLongTermGain >= 0 ? "text-green-700" : "text-red-700"
            }`}
          >
            {totalLongTermGain >= 0 ? "+" : ""}
            {formatCurrency(totalLongTermGain)}
          </div>
          <div className="text-xs text-foreground-secondary mt-1">
            Preferential tax rate (0-20%)
          </div>
        </div>
      </div>

      {/* Position Breakdown */}
      {showBreakdown && (
        <div className="border-t border-paper-gray pt-6">
          <h3 className="font-serif font-semibold text-ink mb-4">
            By Position
          </h3>
          <div className="space-y-2">
            {mockPositionGains.map((position) => (
              <div
                key={position.symbol}
                className="border border-paper-gray rounded-md overflow-hidden"
              >
                {/* Position Row */}
                <button
                  onClick={() => togglePosition(position.symbol)}
                  className="w-full px-4 py-3 flex items-center justify-between hover:bg-paper-cream/50 transition-colors text-left"
                >
                  <div className="flex items-center gap-4">
                    <div>
                      <div className="font-semibold text-ink">
                        {position.symbol}
                      </div>
                      <div className="text-sm text-foreground-secondary">
                        {position.name}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-6">
                    {/* Short-Term */}
                    <div className="text-right">
                      <div className="text-xs text-foreground-secondary">
                        Short-Term
                      </div>
                      <div
                        className={`font-semibold ${
                          position.shortTermGain >= 0
                            ? "text-green-700"
                            : "text-red-700"
                        }`}
                      >
                        {position.shortTermGain === 0
                          ? "—"
                          : `${position.shortTermGain >= 0 ? "+" : ""}${formatCurrency(position.shortTermGain)}`}
                      </div>
                    </div>
                    {/* Long-Term */}
                    <div className="text-right">
                      <div className="text-xs text-foreground-secondary">
                        Long-Term
                      </div>
                      <div
                        className={`font-semibold ${
                          position.longTermGain >= 0
                            ? "text-green-700"
                            : "text-red-700"
                        }`}
                      >
                        {position.longTermGain === 0
                          ? "—"
                          : `${position.longTermGain >= 0 ? "+" : ""}${formatCurrency(position.longTermGain)}`}
                      </div>
                    </div>
                    {/* Expand Icon */}
                    <svg
                      className={`w-5 h-5 text-foreground-secondary transition-transform ${
                        expandedPosition === position.symbol ? "rotate-180" : ""
                      }`}
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M19 9l-7 7-7-7"
                      />
                    </svg>
                  </div>
                </button>

                {/* Tax Lots Detail */}
                {expandedPosition === position.symbol && (
                  <div className="border-t border-paper-gray bg-paper-cream/30 px-4 py-3">
                    <div className="text-xs font-medium text-foreground-secondary mb-2">
                      TAX LOTS
                    </div>
                    <div className="space-y-2">
                      {position.lots.map((lot, idx) => (
                        <div
                          key={idx}
                          className="flex items-center justify-between text-sm py-2 border-b border-paper-gray last:border-0"
                        >
                          <div className="flex items-center gap-3">
                            <span
                              className={`text-xs px-1.5 py-0.5 rounded ${
                                lot.isLongTerm
                                  ? "bg-green-100 text-green-800"
                                  : "bg-amber-100 text-amber-800"
                              }`}
                            >
                              {lot.isLongTerm ? "LT" : "ST"}
                            </span>
                            <div>
                              <div className="text-ink">
                                {lot.shares} shares @ {formatCurrency(lot.costBasis / lot.shares)}
                              </div>
                              <div className="text-xs text-foreground-secondary">
                                Purchased{" "}
                                {new Date(lot.purchaseDate).toLocaleDateString(
                                  "en-US",
                                  { month: "short", day: "numeric", year: "numeric" }
                                )}
                              </div>
                            </div>
                          </div>
                          <div className="text-right">
                            <div
                              className={`font-semibold ${
                                lot.gainLoss >= 0 ? "text-green-700" : "text-red-700"
                              }`}
                            >
                              {lot.gainLoss >= 0 ? "+" : ""}
                              {formatCurrency(lot.gainLoss)}
                            </div>
                            <div
                              className={`text-xs ${
                                lot.gainLoss >= 0 ? "text-green-700" : "text-red-700"
                              }`}
                            >
                              {formatPercent(lot.gainLossPercent)}
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
