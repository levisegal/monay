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
}

interface HoldingDetailPanelProps {
  holding: HoldingData | null;
  quote: StockQuote | null;
  chartData: ChartDataPoint[];
  isOpen: boolean;
  onClose: () => void;
}

export function HoldingDetailPanel({
  holding,
  quote,
  chartData,
  isOpen,
  onClose,
}: HoldingDetailPanelProps) {
  const [timeRange, setTimeRange] = useState<TimeRange>("1M");

  if (!holding || !quote) return null;

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

  const totalGainLoss = holding.value - holding.costBasis;
  const totalGainLossPercent = (totalGainLoss / holding.costBasis) * 100;
  const dayGainLoss = holding.shares * quote.regularMarketChange;

  return (
    <SlideOver isOpen={isOpen} onClose={onClose}>
      {/* Header */}
      <div className="mb-6">
        <div className="flex items-start justify-between mb-2">
          <div>
            <h2 className="font-serif text-2xl font-semibold text-ink">
              {quote.shortName}
            </h2>
            <span className="text-sm text-foreground-secondary">
              {quote.symbol} · {quote.exchange}
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
              {holding.shares.toLocaleString()}
            </div>
          </div>
          <div>
            <div className="text-sm text-foreground-secondary">Market Value</div>
            <div className="font-serif font-semibold text-ink">
              {formatCurrency(holding.value)}
            </div>
          </div>
          <div>
            <div className="text-sm text-foreground-secondary">Cost Basis</div>
            <div className="font-serif font-semibold text-ink">
              {formatCurrency(holding.costBasis)}
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
