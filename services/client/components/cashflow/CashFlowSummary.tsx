"use client";

import { formatCurrency } from "../../lib/formatters";
import type { AnnualCashFlow } from "../../types/stock";

interface CashFlowSummaryProps {
  annualCashFlow: AnnualCashFlow;
}

export function CashFlowSummary({ annualCashFlow }: CashFlowSummaryProps) {
  const { totalProjected, byType } = annualCashFlow;

  const dividendPercent = (byType.dividend / totalProjected) * 100;
  const interestPercent = (byType.interest / totalProjected) * 100;
  const mmfPercent = (byType.mmf / totalProjected) * 100;

  return (
    <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mt-4">
      <div className="bg-paper-cream/50 border border-paper-gray rounded-md p-4">
        <div className="text-sm text-foreground-secondary mb-1">
          Total Projected
        </div>
        <div className="font-serif text-2xl font-semibold text-ink">
          {formatCurrency(totalProjected)}
        </div>
        <div className="text-sm text-foreground-secondary">Annual income</div>
      </div>

      <div className="bg-blue-50 border border-blue-200 rounded-md p-4">
        <div className="flex items-center gap-2 mb-1">
          <div className="w-2 h-2 rounded-full bg-blue-500" />
          <span className="text-sm text-foreground-secondary">Dividends</span>
        </div>
        <div className="font-serif text-2xl font-semibold text-ink">
          {formatCurrency(byType.dividend)}
        </div>
        <div className="text-sm text-foreground-secondary">
          {dividendPercent.toFixed(1)}% of total
        </div>
      </div>

      <div className="bg-green-50 border border-green-200 rounded-md p-4">
        <div className="flex items-center gap-2 mb-1">
          <div className="w-2 h-2 rounded-full bg-green-500" />
          <span className="text-sm text-foreground-secondary">Interest</span>
        </div>
        <div className="font-serif text-2xl font-semibold text-ink">
          {formatCurrency(byType.interest)}
        </div>
        <div className="text-sm text-foreground-secondary">
          {interestPercent.toFixed(1)}% of total
        </div>
      </div>

      <div className="bg-gray-50 border border-gray-200 rounded-md p-4">
        <div className="flex items-center gap-2 mb-1">
          <div className="w-2 h-2 rounded-full bg-gray-500" />
          <span className="text-sm text-foreground-secondary">MMF</span>
        </div>
        <div className="font-serif text-2xl font-semibold text-ink">
          {formatCurrency(byType.mmf)}
        </div>
        <div className="text-sm text-foreground-secondary">
          {mmfPercent.toFixed(1)}% of total
        </div>
      </div>
    </div>
  );
}
