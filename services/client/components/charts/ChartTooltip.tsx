"use client";

import { formatCurrency } from "../../lib/formatters";

interface ChartTooltipProps {
  active?: boolean;
  payload?: Array<{ value: number }>;
  label?: string;
}

export function ChartTooltip({ active, payload, label }: ChartTooltipProps) {
  if (active && payload && payload.length) {
    return (
      <div className="bg-white border border-paper-gray rounded-md shadow-paper p-3">
        <p className="text-sm text-foreground-secondary mb-1">{label}</p>
        <p className="font-serif font-semibold text-ink">
          {formatCurrency(payload[0].value)}
        </p>
      </div>
    );
  }
  return null;
}
