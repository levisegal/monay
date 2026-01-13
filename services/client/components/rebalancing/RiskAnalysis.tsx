"use client";

import { formatCurrency } from "../../lib/formatters";

interface RiskMetric {
  label: string;
  value: string | number;
  description: string;
  status: "good" | "warning" | "danger" | "neutral";
}

interface SectorExposure {
  sector: string;
  percent: number;
  benchmark: number;
  overUnder: number;
}

interface ConcentrationRisk {
  symbol: string;
  name: string;
  percent: number;
  value: number;
  risk: "high" | "medium" | "low";
}

const riskMetrics: RiskMetric[] = [
  {
    label: "Portfolio Beta",
    value: "1.24",
    description: "Volatility relative to market. >1 means more volatile than S&P 500.",
    status: "warning",
  },
  {
    label: "Sharpe Ratio",
    value: "0.82",
    description: "Risk-adjusted return. Higher is better. >1 is considered good.",
    status: "neutral",
  },
  {
    label: "Max Drawdown",
    value: "-18.4%",
    description: "Largest peak-to-trough decline in past year.",
    status: "warning",
  },
  {
    label: "Volatility (Std Dev)",
    value: "16.2%",
    description: "Annualized standard deviation. S&P 500 historical avg is ~15%.",
    status: "warning",
  },
  {
    label: "Dividend Yield",
    value: "1.2%",
    description: "Annual dividend income as % of portfolio value.",
    status: "neutral",
  },
  {
    label: "P/E Ratio (Weighted)",
    value: "28.4",
    description: "Weighted average P/E. S&P 500 avg is ~22.",
    status: "warning",
  },
];

const sectorExposure: SectorExposure[] = [
  { sector: "Technology", percent: 52.6, benchmark: 29.5, overUnder: 23.1 },
  { sector: "Consumer Cyclical", percent: 11.5, benchmark: 10.2, overUnder: 1.3 },
  { sector: "Communication", percent: 5.5, benchmark: 8.7, overUnder: -3.2 },
  { sector: "Financial Services", percent: 0, benchmark: 12.8, overUnder: -12.8 },
  { sector: "Healthcare", percent: 0, benchmark: 13.1, overUnder: -13.1 },
  { sector: "Industrials", percent: 0, benchmark: 8.4, overUnder: -8.4 },
  { sector: "Consumer Defensive", percent: 0, benchmark: 6.8, overUnder: -6.8 },
  { sector: "Energy", percent: 0, benchmark: 4.2, overUnder: -4.2 },
  { sector: "Bonds/Fixed Income", percent: 14.5, benchmark: 0, overUnder: 14.5 },
  { sector: "Cash", percent: 4.5, benchmark: 0, overUnder: 4.5 },
];

const concentrationRisks: ConcentrationRisk[] = [
  { symbol: "MSFT", name: "Microsoft", percent: 24.0, value: 68204, risk: "high" },
  { symbol: "VTI", name: "Vanguard Total Stock", percent: 29.8, value: 84753, risk: "high" },
  { symbol: "AAPL", name: "Apple", percent: 15.1, value: 42840, risk: "medium" },
];

const scenarioAnalysis = [
  {
    scenario: "Market Crash (-30%)",
    impact: -72000,
    impactPercent: -25.3,
    description: "Severe bear market similar to 2008 or 2020",
  },
  {
    scenario: "Tech Correction (-20%)",
    impact: -38000,
    impactPercent: -13.4,
    description: "Technology sector specific downturn",
  },
  {
    scenario: "Interest Rate Hike (+1%)",
    impact: -8500,
    impactPercent: -3.0,
    description: "Impact on bond holdings from rate increases",
  },
  {
    scenario: "Moderate Recession (-15%)",
    impact: -32000,
    impactPercent: -11.2,
    description: "Typical recession-level market decline",
  },
];

export function RiskAnalysis() {
  return (
    <div className="space-y-6">
      {/* Risk Score Overview */}
      <div className="bg-amber-50 border border-amber-200 rounded-md p-4">
        <div className="flex items-center justify-between">
          <div>
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 rounded-full bg-amber-500" />
              <h3 className="font-serif font-semibold text-ink">
                Moderate-High Risk
              </h3>
            </div>
            <p className="text-sm text-foreground-secondary mt-1">
              Your portfolio is more volatile than a typical balanced portfolio due to
              high tech concentration and limited diversification.
            </p>
          </div>
          <div className="text-right">
            <div className="font-serif text-3xl font-semibold text-amber-600">68</div>
            <div className="text-xs text-foreground-secondary">Risk Score (0-100)</div>
          </div>
        </div>
      </div>

      {/* Key Risk Metrics */}
      <div>
        <h3 className="font-serif font-semibold text-ink mb-4">
          Key Risk Metrics
        </h3>
        <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
          {riskMetrics.map((metric) => (
            <div
              key={metric.label}
              className={`p-3 rounded-md border ${
                metric.status === "good"
                  ? "bg-green-50 border-green-200"
                  : metric.status === "warning"
                  ? "bg-amber-50 border-amber-200"
                  : metric.status === "danger"
                  ? "bg-red-50 border-red-200"
                  : "bg-paper-cream/50 border-paper-gray"
              }`}
            >
              <div className="text-xs text-foreground-secondary mb-1">
                {metric.label}
              </div>
              <div className="font-serif text-xl font-semibold text-ink">
                {metric.value}
              </div>
              <div className="text-xs text-foreground-secondary mt-1">
                {metric.description}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Concentration Risk */}
      <div className="border-t border-paper-gray pt-6">
        <h3 className="font-serif font-semibold text-ink mb-4">
          Concentration Risk
        </h3>
        <p className="text-sm text-foreground-secondary mb-4">
          Positions that represent significant portions of your portfolio.
          High concentration increases vulnerability to single-stock events.
        </p>
        <div className="space-y-3">
          {concentrationRisks.map((position) => (
            <div
              key={position.symbol}
              className={`p-3 rounded-md border ${
                position.risk === "high"
                  ? "border-red-200 bg-red-50"
                  : position.risk === "medium"
                  ? "border-amber-200 bg-amber-50"
                  : "border-paper-gray"
              }`}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div
                    className={`w-2 h-2 rounded-full ${
                      position.risk === "high"
                        ? "bg-red-500"
                        : position.risk === "medium"
                        ? "bg-amber-500"
                        : "bg-green-500"
                    }`}
                  />
                  <div>
                    <span className="font-semibold text-ink">{position.symbol}</span>
                    <span className="text-foreground-secondary text-sm ml-2">
                      {position.name}
                    </span>
                  </div>
                </div>
                <div className="text-right">
                  <div className="font-semibold text-ink">{position.percent}%</div>
                  <div className="text-sm text-foreground-secondary">
                    {formatCurrency(position.value)}
                  </div>
                </div>
              </div>
              <div className="mt-2">
                <div className="h-2 bg-gray-200 rounded-full overflow-hidden">
                  <div
                    className={`h-full ${
                      position.risk === "high"
                        ? "bg-red-400"
                        : position.risk === "medium"
                        ? "bg-amber-400"
                        : "bg-green-400"
                    }`}
                    style={{ width: `${position.percent}%` }}
                  />
                </div>
                <div className="flex justify-between mt-1 text-xs text-foreground-secondary">
                  <span>0%</span>
                  <span className="text-amber-600">15% recommended max</span>
                  <span>100%</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Sector Exposure */}
      <div className="border-t border-paper-gray pt-6">
        <h3 className="font-serif font-semibold text-ink mb-4">
          Sector Exposure vs S&P 500
        </h3>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="border-b border-paper-gray">
              <tr className="text-left">
                <th className="pb-2 font-medium text-foreground-secondary">Sector</th>
                <th className="pb-2 font-medium text-foreground-secondary text-right">Your %</th>
                <th className="pb-2 font-medium text-foreground-secondary text-right">S&P 500 %</th>
                <th className="pb-2 font-medium text-foreground-secondary text-right">Difference</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-paper-gray">
              {sectorExposure.map((sector) => (
                <tr key={sector.sector} className="hover:bg-paper-cream/30">
                  <td className="py-2 text-ink">{sector.sector}</td>
                  <td className="py-2 text-right font-medium text-ink">
                    {sector.percent.toFixed(1)}%
                  </td>
                  <td className="py-2 text-right text-foreground-secondary">
                    {sector.benchmark.toFixed(1)}%
                  </td>
                  <td
                    className={`py-2 text-right font-medium ${
                      Math.abs(sector.overUnder) < 2
                        ? "text-foreground-secondary"
                        : sector.overUnder > 0
                        ? "text-amber-600"
                        : "text-blue-600"
                    }`}
                  >
                    {sector.overUnder > 0 ? "+" : ""}
                    {sector.overUnder.toFixed(1)}%
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <div className="mt-3 p-3 bg-amber-50 border border-amber-200 rounded-md">
          <div className="flex items-start gap-2">
            <svg className="w-4 h-4 text-amber-600 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
              <path
                fillRule="evenodd"
                d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
                clipRule="evenodd"
              />
            </svg>
            <div className="text-sm text-amber-800">
              <strong>High Tech Concentration:</strong> Your portfolio has 52.6% in Technology
              vs 29.5% benchmark. Consider diversifying into Healthcare, Financials, and other sectors.
            </div>
          </div>
        </div>
      </div>

      {/* Scenario Analysis */}
      <div className="border-t border-paper-gray pt-6">
        <h3 className="font-serif font-semibold text-ink mb-4">
          Scenario Analysis
        </h3>
        <p className="text-sm text-foreground-secondary mb-4">
          Estimated portfolio impact under various market conditions based on current holdings.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {scenarioAnalysis.map((scenario) => (
            <div
              key={scenario.scenario}
              className="p-4 border border-paper-gray rounded-md hover:border-accent-blue/50 transition-colors"
            >
              <div className="flex items-start justify-between mb-2">
                <div className="font-medium text-ink">{scenario.scenario}</div>
                <div className="text-right">
                  <div className="font-serif font-semibold text-red-600">
                    {formatCurrency(scenario.impact)}
                  </div>
                  <div className="text-xs text-red-600">{scenario.impactPercent}%</div>
                </div>
              </div>
              <p className="text-xs text-foreground-secondary">{scenario.description}</p>
              <div className="mt-2">
                <div className="h-1.5 bg-gray-100 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-red-400"
                    style={{ width: `${Math.abs(scenario.impactPercent)}%` }}
                  />
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Risk Reduction Recommendations */}
      <div className="border-t border-paper-gray pt-6">
        <h3 className="font-serif font-semibold text-ink mb-4">
          Risk Reduction Recommendations
        </h3>
        <div className="space-y-3">
          <div className="flex items-start gap-3 p-3 bg-green-50 border border-green-200 rounded-md">
            <div className="w-6 h-6 rounded-full bg-green-100 flex items-center justify-center text-green-600 font-semibold text-sm">
              1
            </div>
            <div>
              <div className="font-medium text-ink">Reduce Technology Concentration</div>
              <div className="text-sm text-foreground-secondary">
                Sell portions of MSFT, AAPL to bring tech exposure below 35% of portfolio.
              </div>
            </div>
          </div>
          <div className="flex items-start gap-3 p-3 bg-green-50 border border-green-200 rounded-md">
            <div className="w-6 h-6 rounded-full bg-green-100 flex items-center justify-center text-green-600 font-semibold text-sm">
              2
            </div>
            <div>
              <div className="font-medium text-ink">Add Sector Diversification</div>
              <div className="text-sm text-foreground-secondary">
                Consider adding Healthcare (XLV), Financials (XLF), or Consumer Staples (XLP).
              </div>
            </div>
          </div>
          <div className="flex items-start gap-3 p-3 bg-green-50 border border-green-200 rounded-md">
            <div className="w-6 h-6 rounded-full bg-green-100 flex items-center justify-center text-green-600 font-semibold text-sm">
              3
            </div>
            <div>
              <div className="font-medium text-ink">Increase Bond Allocation</div>
              <div className="text-sm text-foreground-secondary">
                Current 14.5% is low for risk reduction. Consider increasing to 30-40%.
              </div>
            </div>
          </div>
          <div className="flex items-start gap-3 p-3 bg-blue-50 border border-blue-200 rounded-md">
            <div className="w-6 h-6 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 font-semibold text-sm">
              4
            </div>
            <div>
              <div className="font-medium text-ink">Consider Protective Options</div>
              <div className="text-sm text-foreground-secondary">
                Put options on concentrated positions can limit downside risk. See Hedging tab.
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
