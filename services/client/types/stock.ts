// Stock quote data
export interface StockQuote {
  symbol: string;
  shortName: string;
  longName: string;
  regularMarketPrice: number;
  regularMarketChange: number;
  regularMarketChangePercent: number;
  regularMarketPreviousClose: number;
  regularMarketOpen: number;
  regularMarketDayHigh: number;
  regularMarketDayLow: number;
  regularMarketVolume: number;
  fiftyTwoWeekHigh: number;
  fiftyTwoWeekLow: number;
  marketCap: number;
  trailingPE: number | null;
  dividendYield: number | null;
  currency: string;
  exchange: string;
  quoteType: "EQUITY" | "ETF" | "MUTUALFUND";
}

// Historical chart data point
export interface ChartDataPoint {
  timestamp: number;
  date: string;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

// Chart data for a stock
export interface StockChartData {
  symbol: string;
  data: ChartDataPoint[];
}

// Holding position (user's shares)
export interface HoldingPosition {
  symbol: string;
  name: string;
  shares: number;
  price: number;
  value: number;
  costBasis: number;
  allocation: number;
  change: number;
  changePercent: number;
}

// Portfolio history point
export interface PortfolioHistoryPoint {
  date: string;
  totalValue: number;
  dayChange: number;
  dayChangePercent: number;
}

// Time range options
export type TimeRange = "1D" | "1W" | "1M" | "3M" | "1Y" | "5Y";
