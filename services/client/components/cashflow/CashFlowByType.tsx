"use client";

import { formatCurrency } from "../../lib/formatters";
import type { AnnualCashFlow, IncomePayment, IncomeType } from "../../types/stock";

interface CashFlowByTypeProps {
  annualCashFlow: AnnualCashFlow;
  payments: IncomePayment[];
}

const incomeTypeLabels: Record<IncomeType, string> = {
  dividend: "Dividends",
  interest: "Interest",
  "mmf-distribution": "MMF Distributions",
};

const incomeTypeColors: Record<IncomeType, { bg: string; fill: string; text: string }> = {
  dividend: { bg: "bg-blue-50", fill: "#3b82f6", text: "text-blue-800" },
  interest: { bg: "bg-green-50", fill: "#10b981", text: "text-green-800" },
  "mmf-distribution": { bg: "bg-gray-50", fill: "#6b7280", text: "text-gray-800" },
};

interface ContributorSummary {
  symbol: string;
  name: string;
  totalAmount: number;
  paymentCount: number;
}

function getTopContributors(
  payments: IncomePayment[],
  type: IncomeType
): ContributorSummary[] {
  const bySymbol = new Map<string, ContributorSummary>();

  for (const p of payments.filter((p) => p.incomeType === type)) {
    const existing = bySymbol.get(p.symbol);
    if (existing) {
      existing.totalAmount += p.totalAmount;
      existing.paymentCount += 1;
    } else {
      bySymbol.set(p.symbol, {
        symbol: p.symbol,
        name: p.name,
        totalAmount: p.totalAmount,
        paymentCount: 1,
      });
    }
  }

  return Array.from(bySymbol.values())
    .sort((a, b) => b.totalAmount - a.totalAmount)
    .slice(0, 3);
}

export function CashFlowByType({ annualCashFlow, payments }: CashFlowByTypeProps) {
  const { totalProjected, byType } = annualCashFlow;

  const segments = [
    { type: "dividend" as IncomeType, value: byType.dividend },
    { type: "interest" as IncomeType, value: byType.interest },
    { type: "mmf-distribution" as IncomeType, value: byType.mmf },
  ].filter((s) => s.value > 0);

  let cumulativePercent = 0;
  const chartSegments = segments.map((s) => {
    const percent = (s.value / totalProjected) * 100;
    const startPercent = cumulativePercent;
    cumulativePercent += percent;
    return { ...s, percent, startPercent };
  });

  return (
    <div className="space-y-6">
      {/* Donut Chart with Legend */}
      <div className="flex flex-col md:flex-row items-center gap-8">
        <div className="relative w-48 h-48">
          <svg viewBox="0 0 100 100" className="w-full h-full -rotate-90">
            {chartSegments.map((seg) => {
              const radius = 40;
              const circumference = 2 * Math.PI * radius;
              const strokeDasharray = `${(seg.percent / 100) * circumference} ${circumference}`;
              const strokeDashoffset =
                -(seg.startPercent / 100) * circumference;

              return (
                <circle
                  key={seg.type}
                  cx="50"
                  cy="50"
                  r={radius}
                  fill="none"
                  stroke={incomeTypeColors[seg.type].fill}
                  strokeWidth="16"
                  strokeDasharray={strokeDasharray}
                  strokeDashoffset={strokeDashoffset}
                />
              );
            })}
          </svg>
          <div className="absolute inset-0 flex flex-col items-center justify-center">
            <div className="text-2xl font-serif font-semibold text-ink">
              {formatCurrency(totalProjected)}
            </div>
            <div className="text-sm text-foreground-secondary">Annual</div>
          </div>
        </div>

        {/* Legend */}
        <div className="flex-1 space-y-3">
          {chartSegments.map((seg) => (
            <div key={seg.type} className="flex items-center gap-3">
              <div
                className="w-4 h-4 rounded"
                style={{ backgroundColor: incomeTypeColors[seg.type].fill }}
              />
              <div className="flex-1">
                <div className="font-medium text-ink">
                  {incomeTypeLabels[seg.type]}
                </div>
              </div>
              <div className="text-right">
                <div className="font-medium text-ink">
                  {formatCurrency(seg.value)}
                </div>
                <div className="text-sm text-foreground-secondary">
                  {seg.percent.toFixed(1)}%
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Top Contributors by Type */}
      <div className="border-t border-paper-gray pt-6">
        <h4 className="font-medium text-ink mb-4">Top Contributors</h4>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {(["dividend", "interest", "mmf-distribution"] as IncomeType[]).map(
            (type) => {
              const contributors = getTopContributors(payments, type);
              if (contributors.length === 0) return null;

              return (
                <div
                  key={type}
                  className={`${incomeTypeColors[type].bg} border border-paper-gray rounded-md p-4`}
                >
                  <div className="flex items-center gap-2 mb-3">
                    <div
                      className="w-3 h-3 rounded"
                      style={{
                        backgroundColor: incomeTypeColors[type].fill,
                      }}
                    />
                    <span className="font-medium text-ink">
                      {incomeTypeLabels[type]}
                    </span>
                  </div>
                  <div className="space-y-2">
                    {contributors.map((c) => (
                      <div
                        key={c.symbol}
                        className="flex items-center justify-between text-sm"
                      >
                        <div>
                          <div className="font-medium text-ink">{c.symbol}</div>
                          <div className="text-xs text-foreground-secondary">
                            {c.paymentCount} payment
                            {c.paymentCount !== 1 ? "s" : ""}
                          </div>
                        </div>
                        <div className="font-medium text-ink">
                          {formatCurrency(c.totalAmount)}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              );
            }
          )}
        </div>
      </div>
    </div>
  );
}
