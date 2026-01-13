"use client";

import { formatCurrency } from "../../lib/formatters";

interface TargetProfile {
  stockAllocation: number;
  bondAllocation: number;
}

interface AllocationComparisonProps {
  targetProfile: TargetProfile;
}

interface AssetAllocation {
  category: string;
  subcategory?: string;
  symbol?: string;
  name: string;
  currentValue: number;
  currentPercent: number;
  targetPercent: number;
  difference: number;
  color: string;
}

const totalPortfolioValue = 284523;

const mockAllocations: AssetAllocation[] = [
  {
    category: "Stocks",
    subcategory: "US Large Cap",
    symbol: "AAPL",
    name: "Apple Inc.",
    currentValue: 42840,
    currentPercent: 15.1,
    targetPercent: 10.0,
    difference: -5.1,
    color: "bg-blue-500",
  },
  {
    category: "Stocks",
    subcategory: "US Large Cap",
    symbol: "MSFT",
    name: "Microsoft Corp.",
    currentValue: 68204,
    currentPercent: 24.0,
    targetPercent: 15.0,
    difference: -9.0,
    color: "bg-blue-400",
  },
  {
    category: "Stocks",
    subcategory: "US Large Cap",
    symbol: "AMZN",
    name: "Amazon.com",
    currentValue: 16934,
    currentPercent: 6.0,
    targetPercent: 5.0,
    difference: -1.0,
    color: "bg-blue-300",
  },
  {
    category: "Stocks",
    subcategory: "US Large Cap",
    symbol: "GOOGL",
    name: "Alphabet Inc.",
    currentValue: 15598,
    currentPercent: 5.5,
    targetPercent: 5.0,
    difference: -0.5,
    color: "bg-blue-200",
  },
  {
    category: "Stocks",
    subcategory: "US Total Market",
    symbol: "VTI",
    name: "Vanguard Total Stock",
    currentValue: 84753,
    currentPercent: 29.8,
    targetPercent: 25.0,
    difference: -4.8,
    color: "bg-indigo-400",
  },
  {
    category: "Bonds",
    subcategory: "US Aggregate",
    symbol: "AGG",
    name: "iShares US Agg Bond",
    currentValue: 41349,
    currentPercent: 14.5,
    targetPercent: 35.0,
    difference: 20.5,
    color: "bg-emerald-400",
  },
  {
    category: "Cash",
    name: "Cash & Equivalents",
    currentValue: 12840,
    currentPercent: 4.5,
    targetPercent: 5.0,
    difference: 0.5,
    color: "bg-gray-400",
  },
];

export function AllocationComparison({ targetProfile }: AllocationComparisonProps) {
  const stocksTotal = mockAllocations
    .filter((a) => a.category === "Stocks")
    .reduce((sum, a) => sum + a.currentPercent, 0);
  const bondsTotal = mockAllocations
    .filter((a) => a.category === "Bonds")
    .reduce((sum, a) => sum + a.currentPercent, 0);
  const cashTotal = mockAllocations
    .filter((a) => a.category === "Cash")
    .reduce((sum, a) => sum + a.currentPercent, 0);

  const categories = [
    {
      name: "Stocks",
      current: stocksTotal,
      target: targetProfile.stockAllocation,
      color: "bg-blue-500",
      bgColor: "bg-blue-100",
    },
    {
      name: "Bonds",
      current: bondsTotal,
      target: targetProfile.bondAllocation,
      color: "bg-emerald-500",
      bgColor: "bg-emerald-100",
    },
    {
      name: "Cash",
      current: cashTotal,
      target: 100 - targetProfile.stockAllocation - targetProfile.bondAllocation || 5,
      color: "bg-gray-500",
      bgColor: "bg-gray-100",
    },
  ];

  return (
    <div className="space-y-6">
      {/* Category Overview */}
      <div>
        <h3 className="font-serif font-semibold text-ink mb-4">
          Asset Class Breakdown
        </h3>
        <div className="space-y-4">
          {categories.map((cat) => {
            const diff = cat.current - cat.target;
            const isOverweight = diff > 0;
            return (
              <div key={cat.name}>
                <div className="flex items-center justify-between mb-2">
                  <span className="font-medium text-ink">{cat.name}</span>
                  <div className="flex items-center gap-4 text-sm">
                    <span className="text-foreground-secondary">
                      Current: <span className="font-medium text-ink">{cat.current.toFixed(1)}%</span>
                    </span>
                    <span className="text-foreground-secondary">
                      Target: <span className="font-medium text-ink">{cat.target}%</span>
                    </span>
                    <span
                      className={`font-medium ${
                        Math.abs(diff) < 1
                          ? "text-foreground-secondary"
                          : isOverweight
                          ? "text-amber-600"
                          : "text-green-600"
                      }`}
                    >
                      {diff > 0 ? "+" : ""}
                      {diff.toFixed(1)}%
                      {Math.abs(diff) >= 1 && (
                        <span className="text-xs ml-1">
                          ({isOverweight ? "overweight" : "underweight"})
                        </span>
                      )}
                    </span>
                  </div>
                </div>
                {/* Comparison Bar */}
                <div className="relative h-8 rounded-md overflow-hidden">
                  {/* Background */}
                  <div className={`absolute inset-0 ${cat.bgColor}`} />
                  {/* Current allocation */}
                  <div
                    className={`absolute top-0 left-0 h-full ${cat.color} opacity-80 transition-all duration-500`}
                    style={{ width: `${cat.current}%` }}
                  />
                  {/* Target marker */}
                  <div
                    className="absolute top-0 h-full w-0.5 bg-ink"
                    style={{ left: `${cat.target}%` }}
                  >
                    <div className="absolute -top-5 left-1/2 -translate-x-1/2 text-xs text-ink whitespace-nowrap">
                      Target
                    </div>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Detailed Position Breakdown */}
      <div className="border-t border-paper-gray pt-6">
        <h3 className="font-serif font-semibold text-ink mb-4">
          Position Details
        </h3>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="border-b border-paper-gray">
              <tr className="text-left">
                <th className="pb-2 font-medium text-foreground-secondary">Position</th>
                <th className="pb-2 font-medium text-foreground-secondary text-right">Value</th>
                <th className="pb-2 font-medium text-foreground-secondary text-right">Current %</th>
                <th className="pb-2 font-medium text-foreground-secondary text-right">Target %</th>
                <th className="pb-2 font-medium text-foreground-secondary text-right">Difference</th>
                <th className="pb-2 font-medium text-foreground-secondary text-right">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-paper-gray">
              {mockAllocations.map((alloc, idx) => {
                const diff = alloc.currentPercent - alloc.targetPercent;
                const diffValue = (diff / 100) * totalPortfolioValue;
                const action =
                  Math.abs(diff) < 0.5
                    ? "Hold"
                    : diff > 0
                    ? "Reduce"
                    : "Increase";
                const actionColor =
                  action === "Hold"
                    ? "text-foreground-secondary"
                    : action === "Reduce"
                    ? "text-red-600"
                    : "text-green-600";

                return (
                  <tr key={idx} className="hover:bg-paper-cream/30">
                    <td className="py-3">
                      <div className="flex items-center gap-2">
                        <div className={`w-2 h-2 rounded-full ${alloc.color}`} />
                        <div>
                          <div className="font-medium text-ink">
                            {alloc.symbol || alloc.name}
                          </div>
                          <div className="text-xs text-foreground-secondary">
                            {alloc.subcategory || alloc.category}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td className="py-3 text-right font-medium text-ink">
                      {formatCurrency(alloc.currentValue)}
                    </td>
                    <td className="py-3 text-right text-foreground-secondary">
                      {alloc.currentPercent.toFixed(1)}%
                    </td>
                    <td className="py-3 text-right text-foreground-secondary">
                      {alloc.targetPercent.toFixed(1)}%
                    </td>
                    <td
                      className={`py-3 text-right font-medium ${
                        Math.abs(diff) < 0.5
                          ? "text-foreground-secondary"
                          : diff > 0
                          ? "text-amber-600"
                          : "text-green-600"
                      }`}
                    >
                      {diff > 0 ? "+" : ""}
                      {diff.toFixed(1)}%
                      <div className="text-xs opacity-75">
                        {diffValue > 0 ? "+" : ""}
                        {formatCurrency(Math.abs(diffValue))}
                      </div>
                    </td>
                    <td className={`py-3 text-right font-medium ${actionColor}`}>
                      {action}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>

      {/* Visual Allocation Pie Comparison */}
      <div className="border-t border-paper-gray pt-6">
        <h3 className="font-serif font-semibold text-ink mb-4">
          Visual Comparison
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Current Allocation */}
          <div className="text-center">
            <div className="text-sm text-foreground-secondary mb-3">Current Allocation</div>
            <div className="relative w-40 h-40 mx-auto">
              <svg viewBox="0 0 100 100" className="transform -rotate-90">
                <circle
                  cx="50"
                  cy="50"
                  r="40"
                  fill="none"
                  stroke="#e5e7eb"
                  strokeWidth="20"
                />
                <circle
                  cx="50"
                  cy="50"
                  r="40"
                  fill="none"
                  stroke="#3b82f6"
                  strokeWidth="20"
                  strokeDasharray={`${stocksTotal * 2.51} 251`}
                />
                <circle
                  cx="50"
                  cy="50"
                  r="40"
                  fill="none"
                  stroke="#10b981"
                  strokeWidth="20"
                  strokeDasharray={`${bondsTotal * 2.51} 251`}
                  strokeDashoffset={`${-stocksTotal * 2.51}`}
                />
                <circle
                  cx="50"
                  cy="50"
                  r="40"
                  fill="none"
                  stroke="#9ca3af"
                  strokeWidth="20"
                  strokeDasharray={`${cashTotal * 2.51} 251`}
                  strokeDashoffset={`${-(stocksTotal + bondsTotal) * 2.51}`}
                />
              </svg>
            </div>
            <div className="flex justify-center gap-4 mt-3 text-xs">
              <span className="flex items-center gap-1">
                <span className="w-2 h-2 bg-blue-500 rounded-full" />
                Stocks {stocksTotal.toFixed(0)}%
              </span>
              <span className="flex items-center gap-1">
                <span className="w-2 h-2 bg-emerald-500 rounded-full" />
                Bonds {bondsTotal.toFixed(0)}%
              </span>
              <span className="flex items-center gap-1">
                <span className="w-2 h-2 bg-gray-400 rounded-full" />
                Cash {cashTotal.toFixed(0)}%
              </span>
            </div>
          </div>

          {/* Target Allocation */}
          <div className="text-center">
            <div className="text-sm text-foreground-secondary mb-3">Target Allocation</div>
            <div className="relative w-40 h-40 mx-auto">
              <svg viewBox="0 0 100 100" className="transform -rotate-90">
                <circle
                  cx="50"
                  cy="50"
                  r="40"
                  fill="none"
                  stroke="#e5e7eb"
                  strokeWidth="20"
                />
                <circle
                  cx="50"
                  cy="50"
                  r="40"
                  fill="none"
                  stroke="#3b82f6"
                  strokeWidth="20"
                  strokeDasharray={`${targetProfile.stockAllocation * 2.51} 251`}
                />
                <circle
                  cx="50"
                  cy="50"
                  r="40"
                  fill="none"
                  stroke="#10b981"
                  strokeWidth="20"
                  strokeDasharray={`${targetProfile.bondAllocation * 2.51} 251`}
                  strokeDashoffset={`${-targetProfile.stockAllocation * 2.51}`}
                />
                <circle
                  cx="50"
                  cy="50"
                  r="40"
                  fill="none"
                  stroke="#9ca3af"
                  strokeWidth="20"
                  strokeDasharray={`${5 * 2.51} 251`}
                  strokeDashoffset={`${-(targetProfile.stockAllocation + targetProfile.bondAllocation) * 2.51}`}
                />
              </svg>
            </div>
            <div className="flex justify-center gap-4 mt-3 text-xs">
              <span className="flex items-center gap-1">
                <span className="w-2 h-2 bg-blue-500 rounded-full" />
                Stocks {targetProfile.stockAllocation}%
              </span>
              <span className="flex items-center gap-1">
                <span className="w-2 h-2 bg-emerald-500 rounded-full" />
                Bonds {targetProfile.bondAllocation}%
              </span>
              <span className="flex items-center gap-1">
                <span className="w-2 h-2 bg-gray-400 rounded-full" />
                Cash 5%
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
