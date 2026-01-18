"use client";

import { useState } from "react";
import { CashFlowSummary } from "./CashFlowSummary";
import { CashFlowCalendar } from "./CashFlowCalendar";
import { CashFlowByType } from "./CashFlowByType";
import { CashFlowTable } from "./CashFlowTable";
import type { CashFlowPeriod, IncomePayment, AnnualCashFlow } from "../../types/stock";

type TabId = "timeline" | "by-source" | "payments";
type TimeHorizon = "calendar-year" | "12-month";

interface Tab {
  id: TabId;
  label: string;
  description: string;
}

const tabs: Tab[] = [
  { id: "timeline", label: "Timeline", description: "Monthly view" },
  { id: "by-source", label: "By Source", description: "Income types" },
  { id: "payments", label: "All Payments", description: "Schedule" },
];

const mockPayments: IncomePayment[] = [
  { id: "mmf-jan", symbol: "SWVXX", name: "Schwab Value Advantage MMF", incomeType: "mmf-distribution", paymentDate: "2026-01-31", amount: 0.004, totalAmount: 50.0, shares: 12500, frequency: "monthly", accountId: "schwab-1" },
  { id: "int-nyc-feb", symbol: "NYC-GO-2035", name: "New York City GO 4.0% 2035", incomeType: "interest", paymentDate: "2026-02-01", amount: 500.0, totalAmount: 500.0, shares: 25, frequency: "semi-annual", accountId: "schwab-1" },
  { id: "int-fnma-feb", symbol: "FNMA-MBS", name: "Fannie Mae 3.5% MBS", incomeType: "interest", paymentDate: "2026-02-01", amount: 350.0, totalAmount: 350.0, shares: 20, frequency: "monthly", accountId: "fidelity-1" },
  { id: "div-aapl-feb", symbol: "AAPL", name: "Apple Inc.", incomeType: "dividend", paymentDate: "2026-02-15", exDate: "2026-02-10", amount: 0.25, totalAmount: 60.0, shares: 240, frequency: "quarterly", accountId: "schwab-1" },
  { id: "mmf-feb", symbol: "SWVXX", name: "Schwab Value Advantage MMF", incomeType: "mmf-distribution", paymentDate: "2026-02-28", amount: 0.004, totalAmount: 50.0, shares: 12500, frequency: "monthly", accountId: "schwab-1" },
  { id: "div-msft-mar", symbol: "MSFT", name: "Microsoft Corp.", incomeType: "dividend", paymentDate: "2026-03-14", exDate: "2026-02-20", amount: 0.75, totalAmount: 112.5, shares: 150, frequency: "quarterly", accountId: "schwab-1" },
  { id: "div-vti-mar", symbol: "VTI", name: "Vanguard Total Stock Market ETF", incomeType: "dividend", paymentDate: "2026-03-28", exDate: "2026-03-23", amount: 0.82, totalAmount: 164.0, shares: 200, frequency: "quarterly", accountId: "vanguard-1" },
  { id: "mmf-mar", symbol: "SWVXX", name: "Schwab Value Advantage MMF", incomeType: "mmf-distribution", paymentDate: "2026-03-31", amount: 0.004, totalAmount: 50.0, shares: 12500, frequency: "monthly", accountId: "schwab-1" },
  { id: "int-agg-apr", symbol: "AGG", name: "iShares Core U.S. Aggregate Bond ETF", incomeType: "interest", paymentDate: "2026-04-07", amount: 0.28, totalAmount: 117.6, shares: 420, frequency: "monthly", accountId: "vanguard-1" },
  { id: "mmf-apr", symbol: "SWVXX", name: "Schwab Value Advantage MMF", incomeType: "mmf-distribution", paymentDate: "2026-04-30", amount: 0.004, totalAmount: 50.0, shares: 12500, frequency: "monthly", accountId: "schwab-1" },
  { id: "int-aapl-bond-may", symbol: "AAPL-2030", name: "Apple Inc. 2.65% 2030", incomeType: "interest", paymentDate: "2026-05-11", amount: 132.5, totalAmount: 132.5, shares: 10, frequency: "semi-annual", accountId: "schwab-1" },
  { id: "div-aapl-may", symbol: "AAPL", name: "Apple Inc.", incomeType: "dividend", paymentDate: "2026-05-15", exDate: "2026-05-10", amount: 0.25, totalAmount: 60.0, shares: 240, frequency: "quarterly", accountId: "schwab-1" },
  { id: "mmf-may", symbol: "SWVXX", name: "Schwab Value Advantage MMF", incomeType: "mmf-distribution", paymentDate: "2026-05-31", amount: 0.004, totalAmount: 50.0, shares: 12500, frequency: "monthly", accountId: "schwab-1" },
  { id: "div-msft-jun", symbol: "MSFT", name: "Microsoft Corp.", incomeType: "dividend", paymentDate: "2026-06-11", exDate: "2026-05-21", amount: 0.75, totalAmount: 112.5, shares: 150, frequency: "quarterly", accountId: "schwab-1" },
  { id: "div-vti-jun", symbol: "VTI", name: "Vanguard Total Stock Market ETF", incomeType: "dividend", paymentDate: "2026-06-26", exDate: "2026-06-22", amount: 0.82, totalAmount: 164.0, shares: 200, frequency: "quarterly", accountId: "vanguard-1" },
  { id: "mmf-jun", symbol: "SWVXX", name: "Schwab Value Advantage MMF", incomeType: "mmf-distribution", paymentDate: "2026-06-30", amount: 0.004, totalAmount: 50.0, shares: 12500, frequency: "monthly", accountId: "schwab-1" },
];

const mockMonthlyBreakdown: CashFlowPeriod[] = [
  { periodLabel: "Jan 2026", startDate: "2026-01-01", endDate: "2026-01-31", dividendIncome: 0, interestIncome: 0, mmfIncome: 50.0, totalIncome: 50.0 },
  { periodLabel: "Feb 2026", startDate: "2026-02-01", endDate: "2026-02-28", dividendIncome: 60.0, interestIncome: 850.0, mmfIncome: 50.0, totalIncome: 960.0 },
  { periodLabel: "Mar 2026", startDate: "2026-03-01", endDate: "2026-03-31", dividendIncome: 276.5, interestIncome: 0, mmfIncome: 50.0, totalIncome: 326.5 },
  { periodLabel: "Apr 2026", startDate: "2026-04-01", endDate: "2026-04-30", dividendIncome: 0, interestIncome: 117.6, mmfIncome: 50.0, totalIncome: 167.6 },
  { periodLabel: "May 2026", startDate: "2026-05-01", endDate: "2026-05-31", dividendIncome: 60.0, interestIncome: 132.5, mmfIncome: 50.0, totalIncome: 242.5 },
  { periodLabel: "Jun 2026", startDate: "2026-06-01", endDate: "2026-06-30", dividendIncome: 276.5, interestIncome: 0, mmfIncome: 50.0, totalIncome: 326.5 },
  { periodLabel: "Jul 2026", startDate: "2026-07-01", endDate: "2026-07-31", dividendIncome: 0, interestIncome: 467.6, mmfIncome: 50.0, totalIncome: 517.6 },
  { periodLabel: "Aug 2026", startDate: "2026-08-01", endDate: "2026-08-31", dividendIncome: 60.0, interestIncome: 500.0, mmfIncome: 50.0, totalIncome: 610.0 },
  { periodLabel: "Sep 2026", startDate: "2026-09-01", endDate: "2026-09-30", dividendIncome: 276.5, interestIncome: 0, mmfIncome: 50.0, totalIncome: 326.5 },
  { periodLabel: "Oct 2026", startDate: "2026-10-01", endDate: "2026-10-31", dividendIncome: 0, interestIncome: 117.6, mmfIncome: 50.0, totalIncome: 167.6 },
  { periodLabel: "Nov 2026", startDate: "2026-11-01", endDate: "2026-11-30", dividendIncome: 60.0, interestIncome: 132.5, mmfIncome: 50.0, totalIncome: 242.5 },
  { periodLabel: "Dec 2026", startDate: "2026-12-01", endDate: "2026-12-31", dividendIncome: 276.5, interestIncome: 0, mmfIncome: 50.0, totalIncome: 326.5 },
];

const mockAnnualCashFlow: AnnualCashFlow = {
  year: 2026,
  totalProjected: 4263.8,
  byType: {
    dividend: 1346.0,
    interest: 2317.8,
    mmf: 600.0,
  },
  byMonth: mockMonthlyBreakdown,
};

export function CashFlowSection() {
  const [activeTab, setActiveTab] = useState<TabId>("timeline");
  const [timeHorizon, setTimeHorizon] = useState<TimeHorizon>("12-month");

  return (
    <div className="bg-white border border-paper-gray rounded-md shadow-paper">
      <div className="p-4 sm:p-6 border-b border-paper-gray">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h2 className="font-serif text-lg font-semibold text-ink">
              Cash Flow
            </h2>
            <p className="text-sm text-foreground-secondary">
              Projected income from dividends, interest & distributions
            </p>
          </div>

          {/* Time Horizon Toggle */}
          <div className="flex items-center gap-2">
            <div className="flex rounded-md border border-paper-gray overflow-hidden">
              <button
                onClick={() => setTimeHorizon("calendar-year")}
                className={`px-3 py-1.5 text-sm font-medium transition-colors ${
                  timeHorizon === "calendar-year"
                    ? "bg-accent-blue text-white"
                    : "bg-white text-foreground-secondary hover:bg-paper-cream"
                }`}
              >
                Calendar Year
              </button>
              <button
                onClick={() => setTimeHorizon("12-month")}
                className={`px-3 py-1.5 text-sm font-medium transition-colors ${
                  timeHorizon === "12-month"
                    ? "bg-accent-blue text-white"
                    : "bg-white text-foreground-secondary hover:bg-paper-cream"
                }`}
              >
                12-Month Forward
              </button>
            </div>
          </div>
        </div>

        <CashFlowSummary annualCashFlow={mockAnnualCashFlow} />
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
        {activeTab === "timeline" && (
          <CashFlowCalendar monthlyData={mockMonthlyBreakdown} />
        )}
        {activeTab === "by-source" && (
          <CashFlowByType annualCashFlow={mockAnnualCashFlow} payments={mockPayments} />
        )}
        {activeTab === "payments" && <CashFlowTable payments={mockPayments} />}
      </div>
    </div>
  );
}
