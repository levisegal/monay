"use client";

import { formatCurrency } from "../../lib/formatters";
import type { CashFlowPeriod } from "../../types/stock";

interface CashFlowCalendarProps {
  monthlyData: CashFlowPeriod[];
}

export function CashFlowCalendar({ monthlyData }: CashFlowCalendarProps) {
  const maxIncome = Math.max(...monthlyData.map((m) => m.totalIncome));
  const totalIncome = monthlyData.reduce((sum, m) => sum + m.totalIncome, 0);

  return (
    <div className="space-y-6">
      <div className="text-sm text-foreground-secondary">
        Monthly income projection showing dividends, interest, and MMF distributions.
      </div>

      {/* Stacked Bar Chart */}
      <div className="space-y-3">
        {monthlyData.map((month) => {
          const dividendWidth =
            maxIncome > 0 ? (month.dividendIncome / maxIncome) * 100 : 0;
          const interestWidth =
            maxIncome > 0 ? (month.interestIncome / maxIncome) * 100 : 0;
          const mmfWidth =
            maxIncome > 0 ? (month.mmfIncome / maxIncome) * 100 : 0;
          const totalWidth =
            maxIncome > 0 ? (month.totalIncome / maxIncome) * 100 : 0;

          return (
            <div key={month.periodLabel} className="group">
              <div className="flex items-center justify-between text-sm mb-1">
                <div className="font-medium text-ink w-20">
                  {month.periodLabel.split(" ")[0]}
                </div>
                <div className="text-foreground-secondary">
                  {formatCurrency(month.totalIncome)}
                </div>
              </div>
              <div className="h-8 bg-paper-cream rounded-md overflow-hidden flex">
                {month.dividendIncome > 0 && (
                  <div
                    className="h-full bg-blue-500 transition-all duration-300"
                    style={{ width: `${dividendWidth}%` }}
                    title={`Dividends: ${formatCurrency(month.dividendIncome)}`}
                  />
                )}
                {month.interestIncome > 0 && (
                  <div
                    className="h-full bg-green-500 transition-all duration-300"
                    style={{ width: `${interestWidth}%` }}
                    title={`Interest: ${formatCurrency(month.interestIncome)}`}
                  />
                )}
                {month.mmfIncome > 0 && (
                  <div
                    className="h-full bg-gray-400 transition-all duration-300"
                    style={{ width: `${mmfWidth}%` }}
                    title={`MMF: ${formatCurrency(month.mmfIncome)}`}
                  />
                )}
              </div>

              {/* Hover Detail */}
              <div className="hidden group-hover:flex gap-4 mt-1 text-xs text-foreground-secondary">
                {month.dividendIncome > 0 && (
                  <span className="flex items-center gap-1">
                    <span className="w-2 h-2 rounded-full bg-blue-500" />
                    Div: {formatCurrency(month.dividendIncome)}
                  </span>
                )}
                {month.interestIncome > 0 && (
                  <span className="flex items-center gap-1">
                    <span className="w-2 h-2 rounded-full bg-green-500" />
                    Int: {formatCurrency(month.interestIncome)}
                  </span>
                )}
                {month.mmfIncome > 0 && (
                  <span className="flex items-center gap-1">
                    <span className="w-2 h-2 rounded-full bg-gray-400" />
                    MMF: {formatCurrency(month.mmfIncome)}
                  </span>
                )}
              </div>
            </div>
          );
        })}
      </div>

      {/* Legend */}
      <div className="flex items-center justify-center gap-6 pt-4 border-t border-paper-gray">
        <div className="flex items-center gap-2 text-sm">
          <div className="w-3 h-3 rounded bg-blue-500" />
          <span className="text-foreground-secondary">Dividends</span>
        </div>
        <div className="flex items-center gap-2 text-sm">
          <div className="w-3 h-3 rounded bg-green-500" />
          <span className="text-foreground-secondary">Interest</span>
        </div>
        <div className="flex items-center gap-2 text-sm">
          <div className="w-3 h-3 rounded bg-gray-400" />
          <span className="text-foreground-secondary">MMF</span>
        </div>
      </div>

      {/* Quarterly Summary */}
      <div className="border-t border-paper-gray pt-4">
        <h4 className="font-medium text-ink mb-3">Quarterly Summary</h4>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          {[0, 1, 2, 3].map((q) => {
            const startIdx = q * 3;
            const quarterMonths = monthlyData.slice(startIdx, startIdx + 3);
            const quarterTotal = quarterMonths.reduce(
              (sum, m) => sum + m.totalIncome,
              0
            );
            return (
              <div
                key={q}
                className="bg-paper-cream/50 border border-paper-gray rounded-md p-3"
              >
                <div className="text-sm text-foreground-secondary">Q{q + 1}</div>
                <div className="font-semibold text-ink">
                  {formatCurrency(quarterTotal)}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
