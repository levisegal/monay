"use client";

import { useState } from "react";
import { SlideOver } from "../ui/SlideOver";
import { StockChart } from "../charts/StockChart";
import { TimeRangeSelector } from "../charts/TimeRangeSelector";
import type { TimeRange, StockQuote, ChartDataPoint } from "../../types/stock";
import { filterByTimeRange, calculateChange } from "../../lib/chartUtils";
import {
  formatCurrency,
  formatLargeNumber,
  formatPercent,
  formatVolume,
  formatChange,
} from "../../lib/formatters";

interface HoldingData {
  name: string;
  symbol: string;
  shares: number;
  price: number;
  value: number;
  costBasis: number;
  change: number;
  changePercent: number;
  accountName?: string;
  accountInstitution?: string;
}

interface HoldingDetailPanelProps {
  holdings: HoldingData[]; // All positions for this symbol across accounts
  quote: StockQuote | null;
  chartData: ChartDataPoint[];
  isOpen: boolean;
  onClose: () => void;
}

export function HoldingDetailPanel({
  holdings,
  quote,
  chartData,
  isOpen,
  onClose,
}: HoldingDetailPanelProps) {
  const [timeRange, setTimeRange] = useState<TimeRange>("1M");

  if (!holdings || holdings.length === 0 || !quote) return null;

  // Aggregate totals across all accounts
  const totalShares = holdings.reduce((sum, h) => sum + h.shares, 0);
  const totalValue = holdings.reduce((sum, h) => sum + h.value, 0);
  const totalCostBasis = holdings.reduce((sum, h) => sum + h.costBasis, 0);
  const firstHolding = holdings[0]; // Use first for name/symbol

  const filteredData = filterByTimeRange(chartData, timeRange);
  const displayData = filteredData.map((d) => {
    const dateObj = new Date(d.date);
    // For 1D, show time (e.g., "9:30 AM"); for longer ranges show date
    const label = timeRange === "1D"
      ? dateObj.toLocaleTimeString("en-US", { hour: "numeric", minute: "2-digit" })
      : dateObj.toLocaleDateString("en-US", { month: "short", day: "numeric" });
    return { date: label, value: d.close };
  });

  const { changePercent: periodChangePercent, isPositive } =
    calculateChange(filteredData);

  const totalGainLoss = totalValue - totalCostBasis;
  const totalGainLossPercent = (totalGainLoss / totalCostBasis) * 100;
  const dayGainLoss = totalShares * quote.regularMarketChange;

  return (
    <SlideOver isOpen={isOpen} onClose={onClose} data-testid="holding-detail-panel">
      {/* Header */}
      <div className="mb-6">
        <div className="flex items-start justify-between mb-2">
          <div>
            <h2 className="font-serif text-2xl font-semibold text-ink">
              {quote.shortName}
            </h2>
            <span className="text-sm text-foreground-secondary">
              {quote.symbol} · {quote.exchange}
              {holdings.length > 1 && (
                <> · {holdings.length} accounts</>
              )}
            </span>
          </div>
          <span
            className={`px-2 py-1 text-xs font-medium rounded ${
              quote.quoteType === "ETF"
                ? "bg-accent-blue/10 text-accent-blue"
                : "bg-paper-gray text-foreground-secondary"
            }`}
          >
            {quote.quoteType}
          </span>
        </div>
        <div className="flex items-baseline gap-3">
          <span className="font-serif text-3xl font-semibold text-ink">
            {formatCurrency(quote.regularMarketPrice)}
          </span>
          <span
            className={`text-lg font-medium ${
              quote.regularMarketChange >= 0 ? "text-green-700" : "text-red-700"
            }`}
          >
            {formatChange(quote.regularMarketChange)} (
            {formatPercent(quote.regularMarketChangePercent)})
          </span>
        </div>
      </div>

      {/* Position Info */}
      <div className="bg-paper-cream/50 border border-paper-gray rounded-md p-4 mb-6">
        <h3 className="font-serif font-semibold text-ink mb-3">Your Position</h3>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <div className="text-sm text-foreground-secondary">Shares</div>
            <div className="font-serif font-semibold text-ink">
              {totalShares.toLocaleString()}
            </div>
          </div>
          <div>
            <div className="text-sm text-foreground-secondary">Market Value</div>
            <div className="font-serif font-semibold text-ink">
              {formatCurrency(totalValue)}
            </div>
          </div>
          <div>
            <div className="text-sm text-foreground-secondary">Cost Basis</div>
            <div className="font-serif font-semibold text-ink">
              {formatCurrency(totalCostBasis)}
            </div>
          </div>
          <div>
            <div className="text-sm text-foreground-secondary">Total Return</div>
            <div
              className={`font-serif font-semibold ${
                totalGainLoss >= 0 ? "text-green-700" : "text-red-700"
              }`}
            >
              {formatChange(totalGainLoss)} ({formatPercent(totalGainLossPercent)}
              )
            </div>
          </div>
          <div className="col-span-2">
            <div className="text-sm text-foreground-secondary">
              Today&apos;s Gain/Loss
            </div>
            <div
              className={`font-serif font-semibold ${
                dayGainLoss >= 0 ? "text-green-700" : "text-red-700"
              }`}
            >
              {formatChange(dayGainLoss)}
            </div>
          </div>
        </div>

        {/* Account Breakdown */}
        {holdings.length > 1 && (
          <div className="mt-4 pt-4 border-t border-paper-gray">
            <div className="text-sm text-foreground-secondary mb-2">By Account</div>
            <div className="space-y-2">
              {holdings.map((h, idx) => (
                <div key={idx} className="flex justify-between items-center text-sm">
                  <span className="text-ink">
                    {h.accountName} ({h.accountInstitution?.split(" ")[0]})
                  </span>
                  <span className="text-foreground-secondary">
                    {h.shares.toLocaleString()} shares · {formatCurrency(h.value)}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Price Chart */}
      <div className="mb-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="font-serif font-semibold text-ink">Price History</h3>
          <TimeRangeSelector value={timeRange} onChange={setTimeRange} />
        </div>
        <StockChart data={displayData} height={220} isPositive={isPositive} />
      </div>

      {/* Key Metrics */}
      <div>
        <h3 className="font-serif font-semibold text-ink mb-3">Key Metrics</h3>
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div className="flex justify-between border-b border-paper-gray pb-2">
            <span className="text-foreground-secondary">Day Range</span>
            <span className="text-ink">
              {formatCurrency(quote.regularMarketDayLow)} -{" "}
              {formatCurrency(quote.regularMarketDayHigh)}
            </span>
          </div>
          <div className="flex justify-between border-b border-paper-gray pb-2">
            <span className="text-foreground-secondary">52-Week Range</span>
            <span className="text-ink">
              {formatCurrency(quote.fiftyTwoWeekLow)} -{" "}
              {formatCurrency(quote.fiftyTwoWeekHigh)}
            </span>
          </div>
          <div className="flex justify-between border-b border-paper-gray pb-2">
            <span className="text-foreground-secondary">Volume</span>
            <span className="text-ink">
              {formatVolume(quote.regularMarketVolume)}
            </span>
          </div>
          <div className="flex justify-between border-b border-paper-gray pb-2">
            <span className="text-foreground-secondary">Market Cap</span>
            <span className="text-ink">{formatLargeNumber(quote.marketCap)}</span>
          </div>
          {quote.trailingPE && (
            <div className="flex justify-between border-b border-paper-gray pb-2">
              <span className="text-foreground-secondary">P/E Ratio</span>
              <span className="text-ink">{quote.trailingPE.toFixed(2)}</span>
            </div>
          )}
          {quote.dividendYield && (
            <div className="flex justify-between border-b border-paper-gray pb-2">
              <span className="text-foreground-secondary">Dividend Yield</span>
              <span className="text-ink">
                {(quote.dividendYield * 100).toFixed(2)}%
              </span>
            </div>
          )}
        </div>
      </div>
    </SlideOver>
  );
}
