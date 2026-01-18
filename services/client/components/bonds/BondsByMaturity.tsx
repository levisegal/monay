"use client";

import { formatCurrency } from "../../lib/formatters";
import type { BondHolding } from "../../types/stock";

interface BondsByMaturityProps {
  holdings: BondHolding[];
  onSelectBond?: (bond: BondHolding) => void;
}

interface MaturityBucket {
  label: string;
  range: string;
  holdings: BondHolding[];
  totalParValue: number;
  totalMarketValue: number;
  avgYield: number;
}

function categorizeByMaturity(holdings: BondHolding[]): MaturityBucket[] {
  const now = new Date();
  const buckets: MaturityBucket[] = [
    { label: "0-1 Year", range: "Short", holdings: [], totalParValue: 0, totalMarketValue: 0, avgYield: 0 },
    { label: "1-3 Years", range: "Short-Med", holdings: [], totalParValue: 0, totalMarketValue: 0, avgYield: 0 },
    { label: "3-5 Years", range: "Medium", holdings: [], totalParValue: 0, totalMarketValue: 0, avgYield: 0 },
    { label: "5-10 Years", range: "Med-Long", holdings: [], totalParValue: 0, totalMarketValue: 0, avgYield: 0 },
    { label: "10+ Years", range: "Long", holdings: [], totalParValue: 0, totalMarketValue: 0, avgYield: 0 },
  ];

  for (const h of holdings) {
    const maturity = new Date(h.maturityDate);
    const yearsToMaturity = (maturity.getTime() - now.getTime()) / (365.25 * 24 * 60 * 60 * 1000);

    let bucketIndex = 4;
    if (yearsToMaturity <= 1) bucketIndex = 0;
    else if (yearsToMaturity <= 3) bucketIndex = 1;
    else if (yearsToMaturity <= 5) bucketIndex = 2;
    else if (yearsToMaturity <= 10) bucketIndex = 3;

    buckets[bucketIndex].holdings.push(h);
    buckets[bucketIndex].totalParValue += h.parValue;
    buckets[bucketIndex].totalMarketValue += h.marketValue;
  }

  for (const bucket of buckets) {
    if (bucket.holdings.length > 0) {
      bucket.avgYield =
        bucket.holdings.reduce((sum, h) => sum + h.yieldToWorst * h.marketValue, 0) /
        bucket.totalMarketValue;
    }
  }

  return buckets;
}

export function BondsByMaturity({ holdings, onSelectBond }: BondsByMaturityProps) {
  const buckets = categorizeByMaturity(holdings);
  const maxValue = Math.max(...buckets.map((b) => b.totalMarketValue));
  const totalValue = buckets.reduce((sum, b) => sum + b.totalMarketValue, 0);

  return (
    <div className="space-y-6">
      <div className="text-sm text-foreground-secondary">
        View your bond portfolio by time to maturity. Shorter duration bonds have less interest rate risk.
      </div>

      {/* Maturity Ladder Chart */}
      <div className="space-y-4">
        {buckets.map((bucket) => {
          const width = maxValue > 0 ? (bucket.totalMarketValue / maxValue) * 100 : 0;
          const percentage = totalValue > 0 ? (bucket.totalMarketValue / totalValue) * 100 : 0;

          return (
            <div key={bucket.label} className="space-y-1">
              <div className="flex items-center justify-between text-sm">
                <div className="font-medium text-ink">{bucket.label}</div>
                <div className="text-foreground-secondary">
                  {formatCurrency(bucket.totalMarketValue)}
                  {percentage > 0 && (
                    <span className="ml-2">({percentage.toFixed(1)}%)</span>
                  )}
                </div>
              </div>
              <div className="h-8 bg-paper-cream rounded-md overflow-hidden flex items-center">
                <div
                  className="h-full bg-accent-blue transition-all duration-500 flex items-center justify-end px-2"
                  style={{ width: `${Math.max(width, bucket.holdings.length > 0 ? 5 : 0)}%` }}
                >
                  {bucket.holdings.length > 0 && (
                    <span className="text-xs text-white font-medium">
                      {bucket.holdings.length}
                    </span>
                  )}
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* Bucket Details */}
      <div className="border-t border-paper-gray pt-6">
        <h3 className="font-serif font-semibold text-ink mb-4">
          Holdings by Maturity
        </h3>
        <div className="space-y-4">
          {buckets
            .filter((b) => b.holdings.length > 0)
            .map((bucket) => (
              <div key={bucket.label}>
                <div className="flex items-center gap-2 mb-2">
                  <span className="font-medium text-ink">{bucket.label}</span>
                  <span className="text-xs px-2 py-0.5 bg-paper-cream rounded text-foreground-secondary">
                    {bucket.range}
                  </span>
                  {bucket.avgYield > 0 && (
                    <span className="text-sm text-foreground-secondary">
                      Avg YTW: {bucket.avgYield.toFixed(2)}%
                    </span>
                  )}
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-2">
                  {bucket.holdings.map((h) => (
                    <div
                      key={h.symbol}
                      onClick={() => onSelectBond?.(h)}
                      className="flex items-center justify-between p-2 bg-paper-cream/50 rounded border border-paper-gray cursor-pointer hover:bg-paper-cream transition-colors"
                    >
                      <div>
                        <div className="font-medium text-ink text-sm">
                          {h.symbol}
                        </div>
                        <div className="text-xs text-foreground-secondary">
                          {new Date(h.maturityDate).toLocaleDateString("en-US", {
                            month: "short",
                            year: "numeric",
                          })}
                        </div>
                      </div>
                      <div className="text-right">
                        <div className="font-medium text-ink text-sm">
                          {formatCurrency(h.marketValue)}
                        </div>
                        <div className="text-xs text-foreground-secondary">
                          {h.yieldToWorst.toFixed(2)}% YTW
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
        </div>
      </div>
    </div>
  );
}
