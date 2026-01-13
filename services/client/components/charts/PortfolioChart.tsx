"use client";

import { useState } from "react";
import { StockChart } from "./StockChart";
import { TimeRangeSelector } from "./TimeRangeSelector";
import type { TimeRange, PortfolioHistoryPoint } from "../../types/stock";
import { filterByTimeRange } from "../../lib/chartUtils";
import { formatCurrency, formatPercent } from "../../lib/formatters";

interface PortfolioChartProps {
  data: PortfolioHistoryPoint[];
}

export function PortfolioChart({ data }: PortfolioChartProps) {
  const [timeRange, setTimeRange] = useState<TimeRange>("1Y");

  // Convert to chart format and filter by range
  const chartPoints = data.map((d) => ({
    timestamp: new Date(d.date).getTime(),
    date: new Date(d.date).toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
    }),
    close: d.totalValue,
    open: d.totalValue,
    high: d.totalValue,
    low: d.totalValue,
    volume: 0,
  }));

  const filteredData = filterByTimeRange(chartPoints, timeRange);
  const displayData = filteredData.map((d) => ({
    date: d.date,
    value: d.close,
  }));

  // Calculate change for the period
  const first = filteredData[0]?.close ?? 0;
  const last = filteredData[filteredData.length - 1]?.close ?? 0;
  const change = last - first;
  const changePercent = first > 0 ? (change / first) * 100 : 0;
  const isPositive = change >= 0;

  return (
    <div className="bg-white border border-paper-gray rounded-md shadow-paper p-4 sm:p-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
        <div>
          <h2 className="font-serif text-lg font-semibold text-ink mb-1">
            Portfolio Performance
          </h2>
          <div className="flex items-baseline gap-2">
            <span
              className={`text-sm font-medium ${
                isPositive ? "text-green-700" : "text-red-700"
              }`}
            >
              {isPositive ? "+" : ""}
              {formatCurrency(change)} ({formatPercent(changePercent)})
            </span>
            <span className="text-sm text-foreground-secondary">
              {timeRange === "1Y" ? "Past year" : `Past ${timeRange}`}
            </span>
          </div>
        </div>
        <TimeRangeSelector value={timeRange} onChange={setTimeRange} />
      </div>

      <StockChart data={displayData} height={240} isPositive={isPositive} />
    </div>
  );
}
