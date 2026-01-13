"use client";

import { useState } from "react";
import { formatCurrency } from "../../lib/formatters";

interface TargetProfile {
  stockAllocation: number;
  bondAllocation: number;
}

interface SuggestedTradesProps {
  targetProfile: TargetProfile;
}

type TradeAction = "BUY" | "SELL" | "HOLD";
type TaxStrategy = "tax-efficient" | "fastest" | "tax-loss-harvest";

interface SuggestedTrade {
  symbol: string;
  name: string;
  action: TradeAction;
  shares: number;
  currentPrice: number;
  estimatedValue: number;
  currentAllocation: number;
  targetAllocation: number;
  taxImpact: number;
  isLongTerm: boolean;
  priority: "high" | "medium" | "low";
  reason: string;
}

const mockTrades: SuggestedTrade[] = [
  {
    symbol: "MSFT",
    name: "Microsoft Corp.",
    action: "SELL",
    shares: 45,
    currentPrice: 378.91,
    estimatedValue: 17051,
    currentAllocation: 24.0,
    targetAllocation: 15.0,
    taxImpact: 3500,
    isLongTerm: true,
    priority: "high",
    reason: "Significantly overweight. Selling reduces concentration risk.",
  },
  {
    symbol: "AAPL",
    name: "Apple Inc.",
    action: "SELL",
    shares: 30,
    currentPrice: 178.50,
    estimatedValue: 5355,
    currentAllocation: 15.1,
    targetAllocation: 10.0,
    taxImpact: 1200,
    isLongTerm: true,
    priority: "medium",
    reason: "Moderately overweight. Consider reducing position.",
  },
  {
    symbol: "VTI",
    name: "Vanguard Total Stock",
    action: "SELL",
    shares: 20,
    currentPrice: 242.15,
    estimatedValue: 4843,
    currentAllocation: 29.8,
    targetAllocation: 25.0,
    taxImpact: 950,
    isLongTerm: true,
    priority: "low",
    reason: "Slightly overweight. Broad market exposure is generally good.",
  },
  {
    symbol: "AGG",
    name: "iShares US Agg Bond",
    action: "BUY",
    shares: 580,
    currentPrice: 98.45,
    estimatedValue: 57101,
    currentAllocation: 14.5,
    targetAllocation: 35.0,
    taxImpact: 0,
    isLongTerm: false,
    priority: "high",
    reason: "Significantly underweight. Adding bonds reduces portfolio volatility.",
  },
  {
    symbol: "BND",
    name: "Vanguard Total Bond",
    action: "BUY",
    shares: 150,
    currentPrice: 72.50,
    estimatedValue: 10875,
    currentAllocation: 0,
    targetAllocation: 10.0,
    taxImpact: 0,
    isLongTerm: false,
    priority: "medium",
    reason: "Consider adding for additional bond diversification.",
  },
];

const taxStrategies: { id: TaxStrategy; label: string; description: string }[] = [
  {
    id: "tax-efficient",
    label: "Tax Efficient",
    description: "Minimize tax impact by selling long-term holdings first",
  },
  {
    id: "fastest",
    label: "Fastest",
    description: "Rebalance immediately regardless of tax consequences",
  },
  {
    id: "tax-loss-harvest",
    label: "Tax-Loss Harvest",
    description: "Prioritize selling losers to offset gains",
  },
];

export function SuggestedTrades({ targetProfile }: SuggestedTradesProps) {
  const [selectedStrategy, setSelectedStrategy] = useState<TaxStrategy>("tax-efficient");
  const [selectedTrades, setSelectedTrades] = useState<Set<string>>(new Set());

  const sellTrades = mockTrades.filter((t) => t.action === "SELL");
  const buyTrades = mockTrades.filter((t) => t.action === "BUY");

  const targetStockPercent = targetProfile.stockAllocation;
  const targetBondPercent = targetProfile.bondAllocation;

  const totalSellValue = sellTrades.reduce((sum, t) => sum + t.estimatedValue, 0);
  const totalBuyValue = buyTrades.reduce((sum, t) => sum + t.estimatedValue, 0);
  const totalTaxImpact = mockTrades.reduce((sum, t) => sum + t.taxImpact, 0);
  const netCashFlow = totalSellValue - totalBuyValue;

  const toggleTrade = (symbol: string) => {
    const newSelected = new Set(selectedTrades);
    if (newSelected.has(symbol)) {
      newSelected.delete(symbol);
    } else {
      newSelected.add(symbol);
    }
    setSelectedTrades(newSelected);
  };

  const selectAll = () => {
    setSelectedTrades(new Set(mockTrades.map((t) => t.symbol)));
  };

  const selectedSellValue = sellTrades
    .filter((t) => selectedTrades.has(t.symbol))
    .reduce((sum, t) => sum + t.estimatedValue, 0);
  const selectedBuyValue = buyTrades
    .filter((t) => selectedTrades.has(t.symbol))
    .reduce((sum, t) => sum + t.estimatedValue, 0);
  const selectedTaxImpact = mockTrades
    .filter((t) => selectedTrades.has(t.symbol))
    .reduce((sum, t) => sum + t.taxImpact, 0);

  return (
    <div className="space-y-6">
      {/* Target Summary */}
      <div className="text-sm text-foreground-secondary">
        Rebalancing to target: <span className="font-medium text-ink">{targetStockPercent}% stocks</span>,{" "}
        <span className="font-medium text-ink">{targetBondPercent}% bonds</span>
      </div>

      {/* Tax Strategy Selector */}
      <div>
        <h3 className="font-serif font-semibold text-ink mb-3">
          Rebalancing Strategy
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
          {taxStrategies.map((strategy) => (
            <button
              key={strategy.id}
              onClick={() => setSelectedStrategy(strategy.id)}
              className={`p-3 rounded-md border text-left transition-all ${
                selectedStrategy === strategy.id
                  ? "border-accent-blue bg-accent-blue/5"
                  : "border-paper-gray hover:border-accent-blue/50"
              }`}
            >
              <div className="flex items-center gap-2">
                <div
                  className={`w-3 h-3 rounded-full border-2 ${
                    selectedStrategy === strategy.id
                      ? "border-accent-blue bg-accent-blue"
                      : "border-paper-gray"
                  }`}
                />
                <span className="font-medium text-ink">{strategy.label}</span>
              </div>
              <p className="text-xs text-foreground-secondary mt-1 ml-5">
                {strategy.description}
              </p>
            </button>
          ))}
        </div>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="bg-red-50 border border-red-200 rounded-md p-3">
          <div className="text-xs text-red-600 mb-1">Total to Sell</div>
          <div className="font-serif text-lg font-semibold text-red-700">
            {formatCurrency(totalSellValue)}
          </div>
        </div>
        <div className="bg-green-50 border border-green-200 rounded-md p-3">
          <div className="text-xs text-green-600 mb-1">Total to Buy</div>
          <div className="font-serif text-lg font-semibold text-green-700">
            {formatCurrency(totalBuyValue)}
          </div>
        </div>
        <div className="bg-amber-50 border border-amber-200 rounded-md p-3">
          <div className="text-xs text-amber-600 mb-1">Est. Tax Impact</div>
          <div className="font-serif text-lg font-semibold text-amber-700">
            {formatCurrency(totalTaxImpact)}
          </div>
        </div>
        <div className={`border rounded-md p-3 ${netCashFlow >= 0 ? "bg-blue-50 border-blue-200" : "bg-purple-50 border-purple-200"}`}>
          <div className={`text-xs mb-1 ${netCashFlow >= 0 ? "text-blue-600" : "text-purple-600"}`}>
            Net Cash {netCashFlow >= 0 ? "Generated" : "Needed"}
          </div>
          <div className={`font-serif text-lg font-semibold ${netCashFlow >= 0 ? "text-blue-700" : "text-purple-700"}`}>
            {formatCurrency(Math.abs(netCashFlow))}
          </div>
        </div>
      </div>

      {/* Trade Selection Header */}
      <div className="flex items-center justify-between border-t border-paper-gray pt-4">
        <h3 className="font-serif font-semibold text-ink">
          Suggested Trades
        </h3>
        <div className="flex items-center gap-4">
          <button
            onClick={selectAll}
            className="text-sm text-accent-blue hover:text-ink transition-colors"
          >
            Select All
          </button>
          <button
            onClick={() => setSelectedTrades(new Set())}
            className="text-sm text-foreground-secondary hover:text-ink transition-colors"
          >
            Clear
          </button>
        </div>
      </div>

      {/* Sell Trades */}
      <div>
        <div className="flex items-center gap-2 mb-3">
          <span className="px-2 py-0.5 bg-red-100 text-red-700 text-xs font-medium rounded">
            SELL
          </span>
          <span className="text-sm text-foreground-secondary">
            Reduce overweight positions
          </span>
        </div>
        <div className="space-y-2">
          {sellTrades.map((trade) => (
            <TradeCard
              key={trade.symbol}
              trade={trade}
              isSelected={selectedTrades.has(trade.symbol)}
              onToggle={() => toggleTrade(trade.symbol)}
            />
          ))}
        </div>
      </div>

      {/* Buy Trades */}
      <div>
        <div className="flex items-center gap-2 mb-3">
          <span className="px-2 py-0.5 bg-green-100 text-green-700 text-xs font-medium rounded">
            BUY
          </span>
          <span className="text-sm text-foreground-secondary">
            Increase underweight positions
          </span>
        </div>
        <div className="space-y-2">
          {buyTrades.map((trade) => (
            <TradeCard
              key={trade.symbol}
              trade={trade}
              isSelected={selectedTrades.has(trade.symbol)}
              onToggle={() => toggleTrade(trade.symbol)}
            />
          ))}
        </div>
      </div>

      {/* Selected Summary */}
      {selectedTrades.size > 0 && (
        <div className="border-t border-paper-gray pt-4">
          <div className="bg-paper-cream/50 rounded-md p-4">
            <div className="flex items-center justify-between mb-3">
              <span className="font-medium text-ink">
                Selected: {selectedTrades.size} trade{selectedTrades.size !== 1 ? "s" : ""}
              </span>
              <button className="px-4 py-2 bg-accent-blue text-white rounded-md hover:bg-accent-blue/90 transition-colors font-medium text-sm">
                Preview Orders
              </button>
            </div>
            <div className="grid grid-cols-3 gap-4 text-sm">
              <div>
                <span className="text-foreground-secondary">Sell Total:</span>
                <span className="font-medium text-red-600 ml-2">
                  {formatCurrency(selectedSellValue)}
                </span>
              </div>
              <div>
                <span className="text-foreground-secondary">Buy Total:</span>
                <span className="font-medium text-green-600 ml-2">
                  {formatCurrency(selectedBuyValue)}
                </span>
              </div>
              <div>
                <span className="text-foreground-secondary">Tax Impact:</span>
                <span className="font-medium text-amber-600 ml-2">
                  {formatCurrency(selectedTaxImpact)}
                </span>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function TradeCard({
  trade,
  isSelected,
  onToggle,
}: {
  trade: SuggestedTrade;
  isSelected: boolean;
  onToggle: () => void;
}) {
  const isSell = trade.action === "SELL";

  return (
    <div
      onClick={onToggle}
      className={`border rounded-md p-4 cursor-pointer transition-all ${
        isSelected
          ? "border-accent-blue bg-accent-blue/5"
          : "border-paper-gray hover:border-accent-blue/50"
      }`}
    >
      <div className="flex items-start justify-between">
        <div className="flex items-start gap-3">
          <div
            className={`mt-1 w-4 h-4 rounded border-2 flex items-center justify-center ${
              isSelected
                ? "border-accent-blue bg-accent-blue"
                : "border-paper-gray"
            }`}
          >
            {isSelected && (
              <svg className="w-2.5 h-2.5 text-white" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fillRule="evenodd"
                  d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                  clipRule="evenodd"
                />
              </svg>
            )}
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="font-semibold text-ink">{trade.symbol}</span>
              <span
                className={`text-xs px-1.5 py-0.5 rounded ${
                  trade.priority === "high"
                    ? "bg-red-100 text-red-700"
                    : trade.priority === "medium"
                    ? "bg-amber-100 text-amber-700"
                    : "bg-gray-100 text-gray-700"
                }`}
              >
                {trade.priority} priority
              </span>
              {trade.isLongTerm && (
                <span className="text-xs px-1.5 py-0.5 bg-green-100 text-green-700 rounded">
                  Long-term
                </span>
              )}
            </div>
            <div className="text-sm text-foreground-secondary">{trade.name}</div>
            <div className="text-xs text-foreground-secondary mt-1">
              {trade.reason}
            </div>
          </div>
        </div>

        <div className="text-right">
          <div className={`font-serif font-semibold ${isSell ? "text-red-600" : "text-green-600"}`}>
            {isSell ? "-" : "+"}{trade.shares} shares
          </div>
          <div className="text-sm text-foreground-secondary">
            @ {formatCurrency(trade.currentPrice)}
          </div>
          <div className={`text-sm font-medium ${isSell ? "text-red-600" : "text-green-600"}`}>
            {formatCurrency(trade.estimatedValue)}
          </div>
          {trade.taxImpact > 0 && (
            <div className="text-xs text-amber-600 mt-1">
              ~{formatCurrency(trade.taxImpact)} tax
            </div>
          )}
        </div>
      </div>

      {/* Allocation Change */}
      <div className="mt-3 pt-3 border-t border-paper-gray/50">
        <div className="flex items-center justify-between text-xs">
          <span className="text-foreground-secondary">Allocation change:</span>
          <span className="text-ink">
            {trade.currentAllocation}% → {trade.targetAllocation}%
          </span>
        </div>
        <div className="mt-1.5 h-1.5 bg-gray-100 rounded-full overflow-hidden">
          <div className="h-full flex">
            <div
              className={`${isSell ? "bg-red-400" : "bg-green-400"}`}
              style={{
                width: `${Math.min(trade.currentAllocation, trade.targetAllocation)}%`,
              }}
            />
            <div
              className={`${isSell ? "bg-red-200" : "bg-green-200"}`}
              style={{
                width: `${Math.abs(trade.currentAllocation - trade.targetAllocation)}%`,
              }}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
