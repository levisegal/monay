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

// Bond types
export type BondType =
  | "corporate"
  | "treasury"
  | "municipal"
  | "agency"
  | "bond-etf";

export interface BondHolding {
  symbol: string;
  name: string;
  cusip?: string;
  bondType: BondType;
  parValue: number;
  marketValue: number;
  quantity: number;
  price: number;
  couponRate: number;
  yieldToWorst: number;
  yieldToMaturity?: number;
  duration: number;
  maturityDate: string;
  nextCouponDate?: string;
  creditRating?: string;
  ratingAgency?: string;
  accruedInterest: number;
  accountId: string;
  priceChange: number;
  priceChangePercent: number;
}

export interface BondSummary {
  totalParValue: number;
  totalMarketValue: number;
  totalAccruedInterest: number;
  weightedAvgYield: number;
  weightedAvgDuration: number;
  weightedAvgCoupon: number;
  bondAllocationPercent: number;
}

// Cash flow types
export type IncomeType = "dividend" | "interest" | "mmf-distribution";

export type PaymentFrequency =
  | "monthly"
  | "quarterly"
  | "semi-annual"
  | "annual"
  | "variable";

export interface IncomePayment {
  id: string;
  symbol: string;
  name: string;
  incomeType: IncomeType;
  paymentDate: string;
  exDate?: string;
  amount: number;
  totalAmount: number;
  shares: number;
  frequency: PaymentFrequency;
  accountId: string;
}

export interface CashFlowPeriod {
  periodLabel: string;
  startDate: string;
  endDate: string;
  dividendIncome: number;
  interestIncome: number;
  mmfIncome: number;
  totalIncome: number;
  payments?: IncomePayment[];
}

export interface AnnualCashFlow {
  year: number;
  totalProjected: number;
  byType: {
    dividend: number;
    interest: number;
    mmf: number;
  };
  byMonth: CashFlowPeriod[];
}
