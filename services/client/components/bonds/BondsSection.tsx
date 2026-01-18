"use client";

import { useState } from "react";
import { BondsSummary } from "./BondsSummary";
import { BondsTable } from "./BondsTable";
import { BondsByType } from "./BondsByType";
import { BondsByMaturity } from "./BondsByMaturity";
import { BondDetailPanel } from "./BondDetailPanel";
import type { BondHolding, BondSummary } from "../../types/stock";

type TabId = "summary" | "by-type" | "maturity";

interface Tab {
  id: TabId;
  label: string;
  description: string;
}

const tabs: Tab[] = [
  { id: "summary", label: "Holdings", description: "All bonds" },
  { id: "by-type", label: "By Type", description: "Bond categories" },
  { id: "maturity", label: "Maturity", description: "Ladder view" },
];

const mockBondHoldings: BondHolding[] = [
  {
    symbol: "AGG",
    name: "iShares Core U.S. Aggregate Bond ETF",
    bondType: "bond-etf",
    parValue: 42000,
    marketValue: 41349,
    quantity: 420,
    price: 98.45,
    couponRate: 3.2,
    yieldToWorst: 4.85,
    duration: 6.1,
    maturityDate: "2030-12-31",
    creditRating: "AA",
    accruedInterest: 112.5,
    accountId: "vanguard-1",
    priceChange: -0.32,
    priceChangePercent: -0.3,
  },
  {
    symbol: "VTIP",
    name: "Vanguard Short-Term Inflation-Protected Securities ETF",
    bondType: "treasury",
    parValue: 15000,
    marketValue: 14850,
    quantity: 300,
    price: 49.5,
    couponRate: 0.125,
    yieldToWorst: 2.45,
    duration: 2.4,
    maturityDate: "2027-06-15",
    creditRating: "AAA",
    accruedInterest: 8.25,
    accountId: "fidelity-1",
    priceChange: 0.15,
    priceChangePercent: 0.3,
  },
  {
    symbol: "AAPL-2030",
    name: "Apple Inc. 2.65% 2030",
    cusip: "037833DV1",
    bondType: "corporate",
    parValue: 10000,
    marketValue: 9450,
    quantity: 10,
    price: 94.5,
    couponRate: 2.65,
    yieldToWorst: 4.92,
    yieldToMaturity: 4.92,
    duration: 5.8,
    maturityDate: "2030-05-11",
    nextCouponDate: "2026-05-11",
    creditRating: "AA+",
    ratingAgency: "S&P",
    accruedInterest: 176.39,
    accountId: "schwab-1",
    priceChange: -0.25,
    priceChangePercent: -0.26,
  },
  {
    symbol: "NYC-GO-2035",
    name: "New York City GO 4.0% 2035",
    cusip: "64966QHT3",
    bondType: "municipal",
    parValue: 25000,
    marketValue: 24125,
    quantity: 25,
    price: 96.5,
    couponRate: 4.0,
    yieldToWorst: 3.85,
    duration: 7.2,
    maturityDate: "2035-08-01",
    nextCouponDate: "2026-02-01",
    creditRating: "AA",
    ratingAgency: "Moody's",
    accruedInterest: 458.33,
    accountId: "schwab-1",
    priceChange: 0.5,
    priceChangePercent: 0.52,
  },
  {
    symbol: "FNMA-MBS",
    name: "Fannie Mae 3.5% MBS",
    bondType: "agency",
    parValue: 20000,
    marketValue: 18600,
    quantity: 20,
    price: 93.0,
    couponRate: 3.5,
    yieldToWorst: 5.15,
    duration: 4.5,
    maturityDate: "2052-07-01",
    nextCouponDate: "2026-02-01",
    creditRating: "AA+",
    accruedInterest: 291.67,
    accountId: "fidelity-1",
    priceChange: -0.45,
    priceChangePercent: -0.48,
  },
  {
    symbol: "BND",
    name: "Vanguard Total Bond Market ETF",
    bondType: "bond-etf",
    parValue: 30000,
    marketValue: 28950,
    quantity: 400,
    price: 72.38,
    couponRate: 3.1,
    yieldToWorst: 4.72,
    duration: 6.3,
    maturityDate: "2031-06-30",
    creditRating: "AA",
    accruedInterest: 77.5,
    accountId: "schwab-1",
    priceChange: -0.18,
    priceChangePercent: -0.25,
  },
];

const mockSummary: BondSummary = {
  totalParValue: 142000,
  totalMarketValue: 137324,
  totalAccruedInterest: 1124.64,
  weightedAvgYield: 4.42,
  weightedAvgDuration: 5.5,
  weightedAvgCoupon: 2.98,
  bondAllocationPercent: 48.3,
};

export function BondsSection() {
  const [activeTab, setActiveTab] = useState<TabId>("summary");
  const [selectedBond, setSelectedBond] = useState<BondHolding | null>(null);

  const handleSelectBond = (bond: BondHolding) => {
    setSelectedBond(bond);
  };

  return (
    <div className="bg-white border border-paper-gray rounded-md shadow-paper">
      <div className="p-4 sm:p-6 border-b border-paper-gray">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h2 className="font-serif text-lg font-semibold text-ink">
              Fixed Income
            </h2>
            <p className="text-sm text-foreground-secondary">
              Bond holdings and metrics
            </p>
          </div>
        </div>

        <BondsSummary summary={mockSummary} />
      </div>

      <div className="border-b border-paper-gray">
        <div className="flex overflow-x-auto">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`flex-1 min-w-[120px] px-4 py-3 text-sm font-medium transition-colors border-b-2 ${
                activeTab === tab.id
                  ? "border-accent-blue text-accent-blue bg-paper-cream/30"
                  : "border-transparent text-foreground-secondary hover:text-ink hover:bg-paper-cream/20"
              }`}
            >
              <div className="text-center">
                <div>{tab.label}</div>
                <div className="text-xs opacity-70 hidden sm:block">
                  {tab.description}
                </div>
              </div>
            </button>
          ))}
        </div>
      </div>

      <div className="p-4 sm:p-6">
        {activeTab === "summary" && (
          <BondsTable holdings={mockBondHoldings} onSelectBond={handleSelectBond} />
        )}
        {activeTab === "by-type" && (
          <BondsByType holdings={mockBondHoldings} onSelectBond={handleSelectBond} />
        )}
        {activeTab === "maturity" && (
          <BondsByMaturity holdings={mockBondHoldings} onSelectBond={handleSelectBond} />
        )}
      </div>

      <BondDetailPanel
        bond={selectedBond}
        isOpen={selectedBond !== null}
        onClose={() => setSelectedBond(null)}
      />
    </div>
  );
}
