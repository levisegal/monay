"use client";

import type { TimeRange } from "../../types/stock";

interface TimeRangeSelectorProps {
  value: TimeRange;
  onChange: (range: TimeRange) => void;
  ranges?: TimeRange[];
}

const defaultRanges: TimeRange[] = ["1D", "1W", "1M", "3M", "1Y", "5Y"];

export function TimeRangeSelector({
  value,
  onChange,
  ranges = defaultRanges,
}: TimeRangeSelectorProps) {
  return (
    <div className="inline-flex border border-paper-gray rounded-md overflow-hidden">
      {ranges.map((range) => (
        <button
          key={range}
          onClick={() => onChange(range)}
          className={`px-3 py-1.5 text-sm font-medium transition-colors ${
            value === range
              ? "bg-paper-gray text-ink"
              : "text-foreground-secondary hover:text-ink hover:bg-paper-cream"
          }`}
        >
          {range}
        </button>
      ))}
    </div>
  );
}
