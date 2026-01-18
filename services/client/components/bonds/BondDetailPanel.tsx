"use client";

import { SlideOver } from "../ui/SlideOver";
import { formatCurrency } from "../../lib/formatters";
import type { BondHolding, BondType } from "../../types/stock";

interface BondDetailPanelProps {
  bond: BondHolding | null;
  isOpen: boolean;
  onClose: () => void;
}

const bondTypeLabels: Record<BondType, string> = {
  corporate: "Corporate Bond",
  treasury: "Treasury",
  municipal: "Municipal Bond",
  agency: "Agency MBS",
  "bond-etf": "Bond ETF",
};

const bondTypeColors: Record<BondType, string> = {
  corporate: "bg-blue-100 text-blue-800",
  treasury: "bg-green-100 text-green-800",
  municipal: "bg-purple-100 text-purple-800",
  agency: "bg-amber-100 text-amber-800",
  "bond-etf": "bg-gray-100 text-gray-800",
};

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString("en-US", {
    month: "long",
    day: "numeric",
    year: "numeric",
  });
}

function calculateDaysToMaturity(maturityDate: string): number {
  const now = new Date();
  const maturity = new Date(maturityDate);
  return Math.floor((maturity.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
}

function formatTimeToMaturity(maturityDate: string): string {
  const days = calculateDaysToMaturity(maturityDate);
  if (days < 0) return "Matured";
  if (days < 30) return `${days} days`;
  if (days < 365) return `${Math.floor(days / 30)} months`;
  const years = days / 365;
  if (years < 2) return `${years.toFixed(1)} years`;
  return `${Math.floor(years)} years`;
}

export function BondDetailPanel({ bond, isOpen, onClose }: BondDetailPanelProps) {
  if (!bond) return null;

  const gainLoss = bond.marketValue - bond.parValue;
  const gainLossPercent = (gainLoss / bond.parValue) * 100;
  const daysToMaturity = calculateDaysToMaturity(bond.maturityDate);
  const annualIncome = bond.parValue * (bond.couponRate / 100);

  return (
    <SlideOver isOpen={isOpen} onClose={onClose}>
      {/* Header */}
      <div className="mb-6">
        <div className="flex items-start justify-between mb-2">
          <div>
            <h2 className="font-serif text-2xl font-semibold text-ink">
              {bond.name}
            </h2>
            <span className="text-sm text-foreground-secondary">
              {bond.symbol}
              {bond.cusip && <> · CUSIP: {bond.cusip}</>}
            </span>
          </div>
          <span className={`px-2 py-1 text-xs font-medium rounded ${bondTypeColors[bond.bondType]}`}>
            {bondTypeLabels[bond.bondType]}
          </span>
        </div>
        <div className="flex items-baseline gap-3">
          <span className="font-serif text-3xl font-semibold text-ink">
            {bond.price.toFixed(2)}
          </span>
          <span className="text-foreground-secondary">
            {bond.bondType === "bond-etf" ? "per share" : "% of par"}
          </span>
          <span
            className={`text-lg font-medium ${
              bond.priceChange >= 0 ? "text-green-700" : "text-red-700"
            }`}
          >
            {bond.priceChange >= 0 ? "+" : ""}{bond.priceChange.toFixed(2)} ({bond.priceChangePercent >= 0 ? "+" : ""}{bond.priceChangePercent.toFixed(2)}%)
          </span>
        </div>
      </div>

      {/* Yield Metrics */}
      <div className="bg-green-50 border border-green-200 rounded-md p-4 mb-6">
        <h3 className="font-serif font-semibold text-ink mb-3">Yield Metrics</h3>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <div className="text-sm text-foreground-secondary">Yield to Worst</div>
            <div className="font-serif text-2xl font-semibold text-green-700">
              {bond.yieldToWorst.toFixed(2)}%
            </div>
          </div>
          {bond.yieldToMaturity && (
            <div>
              <div className="text-sm text-foreground-secondary">Yield to Maturity</div>
              <div className="font-serif text-2xl font-semibold text-ink">
                {bond.yieldToMaturity.toFixed(2)}%
              </div>
            </div>
          )}
          <div>
            <div className="text-sm text-foreground-secondary">Coupon Rate</div>
            <div className="font-serif text-xl font-semibold text-ink">
              {bond.couponRate.toFixed(3)}%
            </div>
          </div>
          <div>
            <div className="text-sm text-foreground-secondary">Annual Income</div>
            <div className="font-serif text-xl font-semibold text-ink">
              {formatCurrency(annualIncome)}
            </div>
          </div>
        </div>
      </div>

      {/* Position Info */}
      <div className="bg-paper-cream/50 border border-paper-gray rounded-md p-4 mb-6">
        <h3 className="font-serif font-semibold text-ink mb-3">Your Position</h3>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <div className="text-sm text-foreground-secondary">
              {bond.bondType === "bond-etf" ? "Shares" : "Bonds"}
            </div>
            <div className="font-serif font-semibold text-ink">
              {bond.quantity.toLocaleString()}
            </div>
          </div>
          <div>
            <div className="text-sm text-foreground-secondary">Market Value</div>
            <div className="font-serif font-semibold text-ink">
              {formatCurrency(bond.marketValue)}
            </div>
          </div>
          <div>
            <div className="text-sm text-foreground-secondary">Par Value</div>
            <div className="font-serif font-semibold text-ink">
              {formatCurrency(bond.parValue)}
            </div>
          </div>
          <div>
            <div className="text-sm text-foreground-secondary">Gain/Loss</div>
            <div
              className={`font-serif font-semibold ${
                gainLoss >= 0 ? "text-green-700" : "text-red-700"
              }`}
            >
              {gainLoss >= 0 ? "+" : ""}{formatCurrency(gainLoss)} ({gainLossPercent >= 0 ? "+" : ""}{gainLossPercent.toFixed(1)}%)
            </div>
          </div>
          <div className="col-span-2">
            <div className="text-sm text-foreground-secondary">Accrued Interest</div>
            <div className="font-serif font-semibold text-ink">
              {formatCurrency(bond.accruedInterest)}
            </div>
          </div>
        </div>
      </div>

      {/* Maturity & Duration */}
      <div className="bg-blue-50 border border-blue-200 rounded-md p-4 mb-6">
        <h3 className="font-serif font-semibold text-ink mb-3">Maturity & Duration</h3>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <div className="text-sm text-foreground-secondary">Maturity Date</div>
            <div className="font-serif font-semibold text-ink">
              {formatDate(bond.maturityDate)}
            </div>
          </div>
          <div>
            <div className="text-sm text-foreground-secondary">Time to Maturity</div>
            <div className="font-serif font-semibold text-ink">
              {formatTimeToMaturity(bond.maturityDate)}
            </div>
          </div>
          <div>
            <div className="text-sm text-foreground-secondary">Modified Duration</div>
            <div className="font-serif font-semibold text-ink">
              {bond.duration.toFixed(2)} years
            </div>
          </div>
          <div>
            <div className="text-sm text-foreground-secondary">Interest Rate Sensitivity</div>
            <div className="font-serif font-semibold text-ink">
              {bond.duration < 3 ? "Low" : bond.duration < 6 ? "Moderate" : "High"}
            </div>
          </div>
        </div>

        {/* Maturity Progress Bar */}
        {daysToMaturity > 0 && (
          <div className="mt-4 pt-4 border-t border-blue-200">
            <div className="flex items-center justify-between text-sm mb-2">
              <span className="text-foreground-secondary">Maturity Progress</span>
              <span className="text-ink">{daysToMaturity.toLocaleString()} days remaining</span>
            </div>
            <div className="h-2 bg-blue-100 rounded-full overflow-hidden">
              <div
                className="h-full bg-blue-500 transition-all"
                style={{ width: `${Math.max(0, 100 - (daysToMaturity / 3650) * 100)}%` }}
              />
            </div>
          </div>
        )}
      </div>

      {/* Credit Rating */}
      {bond.creditRating && (
        <div className="bg-paper-cream/50 border border-paper-gray rounded-md p-4 mb-6">
          <h3 className="font-serif font-semibold text-ink mb-3">Credit Quality</h3>
          <div className="flex items-center gap-4">
            <div className="w-16 h-16 rounded-full bg-green-100 flex items-center justify-center">
              <span className="font-serif text-xl font-bold text-green-800">
                {bond.creditRating}
              </span>
            </div>
            <div>
              <div className="font-medium text-ink">
                {bond.creditRating.startsWith("AAA") ? "Highest Quality" :
                 bond.creditRating.startsWith("AA") ? "High Quality" :
                 bond.creditRating.startsWith("A") ? "Upper Medium Grade" :
                 bond.creditRating.startsWith("BBB") ? "Medium Grade" :
                 "Below Investment Grade"}
              </div>
              {bond.ratingAgency && (
                <div className="text-sm text-foreground-secondary">
                  Rated by {bond.ratingAgency}
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Next Coupon Payment */}
      {bond.nextCouponDate && (
        <div className="bg-amber-50 border border-amber-200 rounded-md p-4">
          <h3 className="font-serif font-semibold text-ink mb-3">Next Coupon Payment</h3>
          <div className="flex items-center justify-between">
            <div>
              <div className="font-medium text-ink">
                {formatDate(bond.nextCouponDate)}
              </div>
              <div className="text-sm text-foreground-secondary">
                {calculateDaysToMaturity(bond.nextCouponDate)} days away
              </div>
            </div>
            <div className="text-right">
              <div className="font-serif text-xl font-semibold text-ink">
                {formatCurrency(annualIncome / 2)}
              </div>
              <div className="text-sm text-foreground-secondary">
                Semi-annual payment
              </div>
            </div>
          </div>
        </div>
      )}
    </SlideOver>
  );
}
