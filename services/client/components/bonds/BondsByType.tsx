"use client";

import { formatCurrency } from "../../lib/formatters";
import type { BondHolding, BondType } from "../../types/stock";

interface BondsByTypeProps {
  holdings: BondHolding[];
  onSelectBond?: (bond: BondHolding) => void;
}

const bondTypeLabels: Record<BondType, string> = {
  corporate: "Corporate",
  treasury: "Treasury",
  municipal: "Municipal",
  agency: "Agency",
  "bond-etf": "Bond ETF",
};

const bondTypeColors: Record<BondType, { bg: string; fill: string }> = {
  corporate: { bg: "bg-blue-100", fill: "#3b82f6" },
  treasury: { bg: "bg-green-100", fill: "#10b981" },
  municipal: { bg: "bg-purple-100", fill: "#8b5cf6" },
  agency: { bg: "bg-amber-100", fill: "#f59e0b" },
  "bond-etf": { bg: "bg-gray-100", fill: "#6b7280" },
};

interface TypeBreakdown {
  type: BondType;
  marketValue: number;
  parValue: number;
  avgYield: number;
  avgDuration: number;
  count: number;
  percentage: number;
}

function calculateBreakdown(holdings: BondHolding[]): TypeBreakdown[] {
  const totalValue = holdings.reduce((sum, h) => sum + h.marketValue, 0);
  const byType = new Map<BondType, BondHolding[]>();

  for (const h of holdings) {
    const list = byType.get(h.bondType) || [];
    list.push(h);
    byType.set(h.bondType, list);
  }

  const breakdown: TypeBreakdown[] = [];
  for (const [type, bonds] of byType) {
    const marketValue = bonds.reduce((sum, b) => sum + b.marketValue, 0);
    const parValue = bonds.reduce((sum, b) => sum + b.parValue, 0);
    const weightedYield =
      bonds.reduce((sum, b) => sum + b.yieldToWorst * b.marketValue, 0) /
      marketValue;
    const weightedDuration =
      bonds.reduce((sum, b) => sum + b.duration * b.marketValue, 0) /
      marketValue;

    breakdown.push({
      type,
      marketValue,
      parValue,
      avgYield: weightedYield,
      avgDuration: weightedDuration,
      count: bonds.length,
      percentage: (marketValue / totalValue) * 100,
    });
  }

  return breakdown.sort((a, b) => b.marketValue - a.marketValue);
}

export function BondsByType({ holdings, onSelectBond }: BondsByTypeProps) {
  const breakdown = calculateBreakdown(holdings);
  const totalValue = holdings.reduce((sum, h) => sum + h.marketValue, 0);

  let cumulativePercent = 0;
  const segments = breakdown.map((b) => {
    const startPercent = cumulativePercent;
    cumulativePercent += b.percentage;
    return { ...b, startPercent };
  });

  return (
    <div className="space-y-6">
      {/* Donut Chart */}
      <div className="flex flex-col md:flex-row items-center gap-8">
        <div className="relative w-48 h-48">
          <svg viewBox="0 0 100 100" className="w-full h-full -rotate-90">
            {segments.map((seg, i) => {
              const radius = 40;
              const circumference = 2 * Math.PI * radius;
              const strokeDasharray = `${(seg.percentage / 100) * circumference} ${circumference}`;
              const strokeDashoffset =
                -(seg.startPercent / 100) * circumference;

              return (
                <circle
                  key={seg.type}
                  cx="50"
                  cy="50"
                  r={radius}
                  fill="none"
                  stroke={bondTypeColors[seg.type].fill}
                  strokeWidth="16"
                  strokeDasharray={strokeDasharray}
                  strokeDashoffset={strokeDashoffset}
                />
              );
            })}
          </svg>
          <div className="absolute inset-0 flex flex-col items-center justify-center">
            <div className="text-2xl font-serif font-semibold text-ink">
              {breakdown.length}
            </div>
            <div className="text-sm text-foreground-secondary">Types</div>
          </div>
        </div>

        {/* Legend */}
        <div className="flex-1 space-y-2">
          {breakdown.map((b) => (
            <div key={b.type} className="flex items-center gap-3">
              <div
                className="w-3 h-3 rounded-full"
                style={{ backgroundColor: bondTypeColors[b.type].fill }}
              />
              <div className="flex-1">
                <div className="font-medium text-ink">
                  {bondTypeLabels[b.type]}
                </div>
                <div className="text-sm text-foreground-secondary">
                  {b.count} holding{b.count !== 1 ? "s" : ""}
                </div>
              </div>
              <div className="text-right">
                <div className="font-medium text-ink">
                  {formatCurrency(b.marketValue)}
                </div>
                <div className="text-sm text-foreground-secondary">
                  {b.percentage.toFixed(1)}%
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Detailed Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {breakdown.map((b) => (
          <div
            key={b.type}
            className={`${bondTypeColors[b.type].bg} border border-paper-gray rounded-md p-4`}
          >
            <div className="flex items-center justify-between mb-3">
              <div className="font-semibold text-ink">
                {bondTypeLabels[b.type]}
              </div>
              <div className="text-sm text-foreground-secondary">
                {b.percentage.toFixed(1)}%
              </div>
            </div>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-foreground-secondary">Market Value</span>
                <span className="font-medium text-ink">
                  {formatCurrency(b.marketValue)}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-foreground-secondary">Avg Yield</span>
                <span className="font-medium text-ink">
                  {b.avgYield.toFixed(2)}%
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-foreground-secondary">Avg Duration</span>
                <span className="font-medium text-ink">
                  {b.avgDuration.toFixed(1)} yrs
                </span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
