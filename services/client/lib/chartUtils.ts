import type { ChartDataPoint, TimeRange } from "../types/stock";

// Filter chart data by time range (relative to last data point, not current date)
export function filterByTimeRange(
  data: ChartDataPoint[],
  range: TimeRange
): ChartDataPoint[] {
  if (data.length === 0) return data;

  // Use the last data point's date as reference, not current date
  const lastDataDate = new Date(data[data.length - 1].date);
  let cutoffDate: Date;

  switch (range) {
    case "1D":
      cutoffDate = new Date(lastDataDate.getTime() - 24 * 60 * 60 * 1000);
      break;
    case "1W":
      cutoffDate = new Date(lastDataDate.getTime() - 7 * 24 * 60 * 60 * 1000);
      break;
    case "1M":
      cutoffDate = new Date(lastDataDate.getTime() - 30 * 24 * 60 * 60 * 1000);
      break;
    case "3M":
      cutoffDate = new Date(lastDataDate.getTime() - 90 * 24 * 60 * 60 * 1000);
      break;
    case "1Y":
      cutoffDate = new Date(lastDataDate.getTime() - 365 * 24 * 60 * 60 * 1000);
      break;
    case "5Y":
      cutoffDate = new Date(lastDataDate.getTime() - 5 * 365 * 24 * 60 * 60 * 1000);
      break;
    default:
      return data;
  }

  return data.filter((point) => new Date(point.date) >= cutoffDate);
}

// Format date for chart axis based on time range
export function formatChartDate(date: string, range: TimeRange): string {
  const d = new Date(date);

  switch (range) {
    case "1D":
      return d.toLocaleTimeString("en-US", { hour: "numeric", minute: "2-digit" });
    case "1W":
      return d.toLocaleDateString("en-US", { weekday: "short" });
    case "1M":
      return d.toLocaleDateString("en-US", { month: "short", day: "numeric" });
    case "3M":
    case "1Y":
      return d.toLocaleDateString("en-US", { month: "short", day: "numeric" });
    case "5Y":
      return d.toLocaleDateString("en-US", { month: "short", year: "2-digit" });
    default:
      return date;
  }
}

// Calculate price change for a dataset
export function calculateChange(data: ChartDataPoint[]): {
  change: number;
  changePercent: number;
  isPositive: boolean;
} {
  if (data.length < 2) {
    return { change: 0, changePercent: 0, isPositive: true };
  }

  const first = data[0].close;
  const last = data[data.length - 1].close;
  const change = last - first;
  const changePercent = (change / first) * 100;

  return {
    change,
    changePercent,
    isPositive: change >= 0,
  };
}
