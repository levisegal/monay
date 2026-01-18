"use client";

import { formatCurrency } from "../../lib/formatters";
import type { BondHolding, BondType } from "../../types/stock";

interface BondsTableProps {
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

const bondTypeColors: Record<BondType, string> = {
  corporate: "bg-blue-100 text-blue-800",
  treasury: "bg-green-100 text-green-800",
  municipal: "bg-purple-100 text-purple-800",
  agency: "bg-amber-100 text-amber-800",
  "bond-etf": "bg-gray-100 text-gray-800",
};

function formatMaturityDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString("en-US", {
    month: "short",
    year: "numeric",
  });
}

export function BondsTable({ holdings, onSelectBond }: BondsTableProps) {
  return (
    <div>
      {/* Mobile Cards */}
      <div className="md:hidden space-y-3">
        {holdings.map((bond) => (
          <div
            key={bond.symbol}
            onClick={() => onSelectBond?.(bond)}
            className="border border-paper-gray rounded-md p-4 bg-paper-cream/20 cursor-pointer hover:shadow-paper-hover transition-all"
          >
            <div className="flex items-start justify-between mb-3">
              <div>
                <div className="font-semibold text-ink">{bond.symbol}</div>
                <div className="text-sm text-foreground-secondary line-clamp-1">
                  {bond.name}
                </div>
              </div>
              <span
                className={`text-xs px-2 py-0.5 rounded ${bondTypeColors[bond.bondType]}`}
              >
                {bondTypeLabels[bond.bondType]}
              </span>
            </div>

            <div className="grid grid-cols-2 gap-3 text-sm">
              <div>
                <div className="text-foreground-secondary">YTW</div>
                <div className="font-medium text-ink">
                  {bond.yieldToWorst.toFixed(2)}%
                </div>
              </div>
              <div>
                <div className="text-foreground-secondary">Coupon</div>
                <div className="font-medium text-ink">
                  {bond.couponRate.toFixed(2)}%
                </div>
              </div>
              <div>
                <div className="text-foreground-secondary">Duration</div>
                <div className="font-medium text-ink">
                  {bond.duration.toFixed(1)} yrs
                </div>
              </div>
              <div>
                <div className="text-foreground-secondary">Rating</div>
                <div className="font-medium text-ink">
                  {bond.creditRating || "—"}
                </div>
              </div>
              <div>
                <div className="text-foreground-secondary">Maturity</div>
                <div className="font-medium text-ink">
                  {formatMaturityDate(bond.maturityDate)}
                </div>
              </div>
              <div>
                <div className="text-foreground-secondary">Market Value</div>
                <div className="font-medium text-ink">
                  {formatCurrency(bond.marketValue)}
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Desktop Table */}
      <div className="hidden md:block overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-paper-gray">
              <th className="text-left py-3 px-2 font-medium text-foreground-secondary">
                Security
              </th>
              <th className="text-right py-3 px-2 font-medium text-foreground-secondary">
                YTW
              </th>
              <th className="text-right py-3 px-2 font-medium text-foreground-secondary">
                Coupon
              </th>
              <th className="text-right py-3 px-2 font-medium text-foreground-secondary">
                Duration
              </th>
              <th className="text-center py-3 px-2 font-medium text-foreground-secondary">
                Rating
              </th>
              <th className="text-right py-3 px-2 font-medium text-foreground-secondary">
                Maturity
              </th>
              <th className="text-right py-3 px-2 font-medium text-foreground-secondary">
                Par Value
              </th>
              <th className="text-right py-3 px-2 font-medium text-foreground-secondary">
                Market Value
              </th>
            </tr>
          </thead>
          <tbody>
            {holdings.map((bond) => (
              <tr
                key={bond.symbol}
                onClick={() => onSelectBond?.(bond)}
                className="border-b border-paper-gray hover:bg-paper-cream/30 transition-colors cursor-pointer"
              >
                <td className="py-3 px-2">
                  <div className="flex items-center gap-2">
                    <div>
                      <div className="font-medium text-ink">{bond.symbol}</div>
                      <div className="text-xs text-foreground-secondary line-clamp-1 max-w-[200px]">
                        {bond.name}
                      </div>
                    </div>
                    <span
                      className={`text-xs px-1.5 py-0.5 rounded whitespace-nowrap ${bondTypeColors[bond.bondType]}`}
                    >
                      {bondTypeLabels[bond.bondType]}
                    </span>
                  </div>
                </td>
                <td className="py-3 px-2 text-right font-medium text-ink">
                  {bond.yieldToWorst.toFixed(2)}%
                </td>
                <td className="py-3 px-2 text-right text-ink">
                  {bond.couponRate.toFixed(2)}%
                </td>
                <td className="py-3 px-2 text-right text-ink">
                  {bond.duration.toFixed(1)} yrs
                </td>
                <td className="py-3 px-2 text-center">
                  <span className="font-medium text-ink">
                    {bond.creditRating || "—"}
                  </span>
                </td>
                <td className="py-3 px-2 text-right text-ink">
                  {formatMaturityDate(bond.maturityDate)}
                </td>
                <td className="py-3 px-2 text-right text-foreground-secondary">
                  {formatCurrency(bond.parValue)}
                </td>
                <td className="py-3 px-2 text-right font-medium text-ink">
                  {formatCurrency(bond.marketValue)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
