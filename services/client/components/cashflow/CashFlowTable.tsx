"use client";

import { useState } from "react";
import { formatCurrency } from "../../lib/formatters";
import type { IncomePayment, IncomeType } from "../../types/stock";

interface CashFlowTableProps {
  payments: IncomePayment[];
}

const incomeTypeLabels: Record<IncomeType, string> = {
  dividend: "Dividend",
  interest: "Interest",
  "mmf-distribution": "MMF",
};

const incomeTypeBadgeColors: Record<IncomeType, string> = {
  dividend: "bg-blue-100 text-blue-800",
  interest: "bg-green-100 text-green-800",
  "mmf-distribution": "bg-gray-100 text-gray-800",
};

const frequencyLabels: Record<string, string> = {
  monthly: "Monthly",
  quarterly: "Quarterly",
  "semi-annual": "Semi-Annual",
  annual: "Annual",
  variable: "Variable",
};

function formatPaymentDate(dateStr: string): string {
  const date = new Date(dateStr);
  return date.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
  });
}

export function CashFlowTable({ payments }: CashFlowTableProps) {
  const [typeFilter, setTypeFilter] = useState<IncomeType | "all">("all");

  const filteredPayments =
    typeFilter === "all"
      ? payments
      : payments.filter((p) => p.incomeType === typeFilter);

  const sortedPayments = [...filteredPayments].sort(
    (a, b) =>
      new Date(a.paymentDate).getTime() - new Date(b.paymentDate).getTime()
  );

  return (
    <div className="space-y-4">
      {/* Filter Buttons */}
      <div className="flex items-center gap-2 flex-wrap">
        <span className="text-sm text-foreground-secondary">Filter:</span>
        <button
          onClick={() => setTypeFilter("all")}
          className={`px-3 py-1 text-sm rounded-md transition-colors ${
            typeFilter === "all"
              ? "bg-ink text-white"
              : "bg-paper-cream text-foreground-secondary hover:bg-paper-gray"
          }`}
        >
          All
        </button>
        <button
          onClick={() => setTypeFilter("dividend")}
          className={`px-3 py-1 text-sm rounded-md transition-colors ${
            typeFilter === "dividend"
              ? "bg-blue-500 text-white"
              : "bg-blue-50 text-blue-800 hover:bg-blue-100"
          }`}
        >
          Dividends
        </button>
        <button
          onClick={() => setTypeFilter("interest")}
          className={`px-3 py-1 text-sm rounded-md transition-colors ${
            typeFilter === "interest"
              ? "bg-green-500 text-white"
              : "bg-green-50 text-green-800 hover:bg-green-100"
          }`}
        >
          Interest
        </button>
        <button
          onClick={() => setTypeFilter("mmf-distribution")}
          className={`px-3 py-1 text-sm rounded-md transition-colors ${
            typeFilter === "mmf-distribution"
              ? "bg-gray-500 text-white"
              : "bg-gray-50 text-gray-800 hover:bg-gray-100"
          }`}
        >
          MMF
        </button>
      </div>

      {/* Mobile Cards */}
      <div className="md:hidden space-y-3">
        {sortedPayments.map((payment) => (
          <div
            key={payment.id}
            className="border border-paper-gray rounded-md p-4 bg-paper-cream/20"
          >
            <div className="flex items-start justify-between mb-2">
              <div>
                <div className="font-semibold text-ink">{payment.symbol}</div>
                <div className="text-sm text-foreground-secondary line-clamp-1">
                  {payment.name}
                </div>
              </div>
              <span
                className={`text-xs px-2 py-0.5 rounded ${incomeTypeBadgeColors[payment.incomeType]}`}
              >
                {incomeTypeLabels[payment.incomeType]}
              </span>
            </div>
            <div className="grid grid-cols-2 gap-2 text-sm">
              <div>
                <div className="text-foreground-secondary">Payment Date</div>
                <div className="font-medium text-ink">
                  {formatPaymentDate(payment.paymentDate)}
                </div>
              </div>
              <div>
                <div className="text-foreground-secondary">Amount</div>
                <div className="font-medium text-ink">
                  {formatCurrency(payment.totalAmount)}
                </div>
              </div>
              <div>
                <div className="text-foreground-secondary">Frequency</div>
                <div className="font-medium text-ink">
                  {frequencyLabels[payment.frequency]}
                </div>
              </div>
              {payment.exDate && (
                <div>
                  <div className="text-foreground-secondary">Ex-Date</div>
                  <div className="font-medium text-ink">
                    {formatPaymentDate(payment.exDate)}
                  </div>
                </div>
              )}
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
                Date
              </th>
              <th className="text-left py-3 px-2 font-medium text-foreground-secondary">
                Security
              </th>
              <th className="text-center py-3 px-2 font-medium text-foreground-secondary">
                Type
              </th>
              <th className="text-right py-3 px-2 font-medium text-foreground-secondary">
                Amount
              </th>
              <th className="text-center py-3 px-2 font-medium text-foreground-secondary">
                Frequency
              </th>
              <th className="text-left py-3 px-2 font-medium text-foreground-secondary">
                Ex-Date
              </th>
            </tr>
          </thead>
          <tbody>
            {sortedPayments.map((payment) => (
              <tr
                key={payment.id}
                className="border-b border-paper-gray hover:bg-paper-cream/30 transition-colors"
              >
                <td className="py-3 px-2 font-medium text-ink">
                  {formatPaymentDate(payment.paymentDate)}
                </td>
                <td className="py-3 px-2">
                  <div className="font-medium text-ink">{payment.symbol}</div>
                  <div className="text-xs text-foreground-secondary line-clamp-1 max-w-[200px]">
                    {payment.name}
                  </div>
                </td>
                <td className="py-3 px-2 text-center">
                  <span
                    className={`text-xs px-2 py-0.5 rounded ${incomeTypeBadgeColors[payment.incomeType]}`}
                  >
                    {incomeTypeLabels[payment.incomeType]}
                  </span>
                </td>
                <td className="py-3 px-2 text-right font-medium text-ink">
                  {formatCurrency(payment.totalAmount)}
                </td>
                <td className="py-3 px-2 text-center text-foreground-secondary">
                  {frequencyLabels[payment.frequency]}
                </td>
                <td className="py-3 px-2 text-foreground-secondary">
                  {payment.exDate ? formatPaymentDate(payment.exDate) : "—"}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {sortedPayments.length === 0 && (
        <div className="text-center py-8 text-foreground-secondary">
          No payments found for the selected filter.
        </div>
      )}
    </div>
  );
}
