"use client";

import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { ChartTooltip } from "./ChartTooltip";

interface StockChartProps {
  data: Array<{ date: string; value: number }>;
  height?: number;
  isPositive?: boolean;
}

export function StockChart({
  data,
  height = 280,
  isPositive = true,
}: StockChartProps) {
  const color = isPositive ? "#2C5F77" : "#dc2626";
  const gradientId = `gradient-${isPositive ? "positive" : "negative"}`;

  // Calculate Y axis domain with some padding
  const values = data.map((d) => d.value);
  const min = Math.min(...values);
  const max = Math.max(...values);
  const padding = (max - min) * 0.1;

  return (
    <ResponsiveContainer width="100%" height={height}>
      <AreaChart data={data} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
        <defs>
          <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor={color} stopOpacity={0.3} />
            <stop offset="95%" stopColor={color} stopOpacity={0} />
          </linearGradient>
        </defs>
        <CartesianGrid
          strokeDasharray="3 3"
          stroke="#E8E4DD"
          vertical={false}
        />
        <XAxis
          dataKey="date"
          tick={{ fill: "#6B6456", fontSize: 12 }}
          tickLine={false}
          axisLine={{ stroke: "#E8E4DD" }}
          tickMargin={8}
        />
        <YAxis
          tick={{ fill: "#6B6456", fontSize: 12 }}
          tickLine={false}
          axisLine={false}
          tickFormatter={(value) => `$${value.toLocaleString()}`}
          domain={[min - padding, max + padding]}
          width={70}
        />
        <Tooltip content={<ChartTooltip />} />
        <Area
          type="monotone"
          dataKey="value"
          stroke={color}
          strokeWidth={2}
          fillOpacity={1}
          fill={`url(#${gradientId})`}
        />
      </AreaChart>
    </ResponsiveContainer>
  );
}
