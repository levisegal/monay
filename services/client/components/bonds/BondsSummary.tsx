"use client";

import { formatCurrency } from "../../lib/formatters";
import type { BondSummary } from "../../types/stock";

interface BondsSummaryProps {
  summary: BondSummary;
}

export function BondsSummary({ summary }: BondsSummaryProps) {
  return (
    <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mt-4">
      <div className="bg-paper-cream/50 border border-paper-gray rounded-md p-4">
        <div className="text-sm text-foreground-secondary mb-1">
          Bond Allocation
        </div>
        <div className="font-serif text-2xl font-semibold text-ink">
          {formatCurrency(summary.totalMarketValue)}
        </div>
        <div className="text-sm text-foreground-secondary">
          {summary.bondAllocationPercent.toFixed(1)}% of portfolio
        </div>
      </div>

      <div className="bg-paper-cream/50 border border-paper-gray rounded-md p-4">
        <div className="text-sm text-foreground-secondary mb-1">
          Weighted Avg Yield
        </div>
        <div className="font-serif text-2xl font-semibold text-ink">
          {summary.weightedAvgYield.toFixed(2)}%
        </div>
        <div className="text-sm text-foreground-secondary">YTW</div>
      </div>

      <div className="bg-paper-cream/50 border border-paper-gray rounded-md p-4">
        <div className="text-sm text-foreground-secondary mb-1">
          Avg Duration
        </div>
        <div className="font-serif text-2xl font-semibold text-ink">
          {summary.weightedAvgDuration.toFixed(1)} yrs
        </div>
        <div className="text-sm text-foreground-secondary">Modified</div>
      </div>

      <div className="bg-paper-cream/50 border border-paper-gray rounded-md p-4">
        <div className="text-sm text-foreground-secondary mb-1">
          Accrued Interest
        </div>
        <div className="font-serif text-2xl font-semibold text-ink">
          {formatCurrency(summary.totalAccruedInterest)}
        </div>
        <div className="text-sm text-foreground-secondary">Pending</div>
      </div>
    </div>
  );
}
