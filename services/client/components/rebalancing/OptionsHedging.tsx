"use client";

import { useState } from "react";
import { formatCurrency } from "../../lib/formatters";

type HedgeStrategy = "protective-put" | "collar" | "covered-call" | "put-spread";

interface StrategyInfo {
  id: HedgeStrategy;
  name: string;
  description: string;
  riskProfile: string;
  costType: "debit" | "credit" | "neutral";
  maxLoss: string;
  maxGain: string;
  bestFor: string;
}

interface OptionContract {
  type: "PUT" | "CALL";
  strike: number;
  expiration: string;
  premium: number;
  delta: number;
  impliedVol: number;
  openInterest: number;
  action: "BUY" | "SELL";
}

interface HedgeProposal {
  symbol: string;
  currentPrice: number;
  sharesHeld: number;
  positionValue: number;
  strategy: HedgeStrategy;
  contracts: OptionContract[];
  totalCost: number;
  protectionLevel: number;
  breakeven: number;
  maxProtectedLoss: number;
}

const strategies: StrategyInfo[] = [
  {
    id: "protective-put",
    name: "Protective Put",
    description: "Buy put options to protect against downside while maintaining unlimited upside.",
    riskProfile: "Conservative",
    costType: "debit",
    maxLoss: "Premium paid + (Stock price - Strike)",
    maxGain: "Unlimited",
    bestFor: "Long-term holders wanting insurance against crashes",
  },
  {
    id: "collar",
    name: "Collar Strategy",
    description: "Buy protective put and sell covered call to reduce or eliminate hedge cost.",
    riskProfile: "Moderate",
    costType: "neutral",
    maxLoss: "Limited (to put strike)",
    maxGain: "Limited (to call strike)",
    bestFor: "Reducing hedge cost while accepting capped upside",
  },
  {
    id: "covered-call",
    name: "Covered Call",
    description: "Sell call options against shares you own to generate income.",
    riskProfile: "Income-focused",
    costType: "credit",
    maxLoss: "Full downside exposure",
    maxGain: "Premium + (Strike - Current price)",
    bestFor: "Generating income in sideways or slightly bullish markets",
  },
  {
    id: "put-spread",
    name: "Put Spread",
    description: "Buy put and sell lower strike put to reduce cost of downside protection.",
    riskProfile: "Moderate",
    costType: "debit",
    maxLoss: "Limited (between strikes)",
    maxGain: "Limited to spread width - premium",
    bestFor: "Cost-effective protection against moderate declines",
  },
];

const mockHedgeProposals: HedgeProposal[] = [
  {
    symbol: "MSFT",
    currentPrice: 378.91,
    sharesHeld: 180,
    positionValue: 68204,
    strategy: "protective-put",
    contracts: [
      {
        type: "PUT",
        strike: 360,
        expiration: "2025-03-21",
        premium: 8.50,
        delta: -0.32,
        impliedVol: 24.5,
        openInterest: 12500,
        action: "BUY",
      },
    ],
    totalCost: 1530,
    protectionLevel: 95,
    breakeven: 387.41,
    maxProtectedLoss: 4932,
  },
  {
    symbol: "AAPL",
    currentPrice: 178.50,
    sharesHeld: 240,
    positionValue: 42840,
    strategy: "collar",
    contracts: [
      {
        type: "PUT",
        strike: 170,
        expiration: "2025-03-21",
        premium: 4.20,
        delta: -0.28,
        impliedVol: 22.1,
        openInterest: 45000,
        action: "BUY",
      },
      {
        type: "CALL",
        strike: 190,
        expiration: "2025-03-21",
        premium: 3.80,
        delta: 0.35,
        impliedVol: 21.5,
        openInterest: 38000,
        action: "SELL",
      },
    ],
    totalCost: 96,
    protectionLevel: 95.3,
    breakeven: 178.90,
    maxProtectedLoss: 2040,
  },
  {
    symbol: "VTI",
    currentPrice: 242.15,
    sharesHeld: 350,
    positionValue: 84753,
    strategy: "put-spread",
    contracts: [
      {
        type: "PUT",
        strike: 235,
        expiration: "2025-03-21",
        premium: 5.80,
        delta: -0.35,
        impliedVol: 18.2,
        openInterest: 8500,
        action: "BUY",
      },
      {
        type: "PUT",
        strike: 220,
        expiration: "2025-03-21",
        premium: 2.40,
        delta: -0.18,
        impliedVol: 19.5,
        openInterest: 6200,
        action: "SELL",
      },
    ],
    totalCost: 1190,
    protectionLevel: 97,
    breakeven: 245.55,
    maxProtectedLoss: 2450,
  },
];

export function OptionsHedging() {
  const [selectedStrategy, setSelectedStrategy] = useState<HedgeStrategy>("protective-put");
  const [selectedProposals, setSelectedProposals] = useState<Set<string>>(new Set());
  const [protectionTarget, setProtectionTarget] = useState(10);

  const filteredProposals = mockHedgeProposals;
  const totalHedgeCost = filteredProposals
    .filter((p) => selectedProposals.has(p.symbol))
    .reduce((sum, p) => sum + p.totalCost, 0);
  const totalProtectedValue = filteredProposals
    .filter((p) => selectedProposals.has(p.symbol))
    .reduce((sum, p) => sum + p.positionValue, 0);

  const toggleProposal = (symbol: string) => {
    const newSelected = new Set(selectedProposals);
    if (newSelected.has(symbol)) {
      newSelected.delete(symbol);
    } else {
      newSelected.add(symbol);
    }
    setSelectedProposals(newSelected);
  };

  return (
    <div className="space-y-6">
      {/* Introduction */}
      <div className="bg-blue-50 border border-blue-200 rounded-md p-4">
        <h3 className="font-serif font-semibold text-ink mb-2">
          Options-Based Portfolio Protection
        </h3>
        <p className="text-sm text-foreground-secondary">
          Use options strategies to hedge your concentrated positions against market downturns.
          Options provide insurance-like protection while allowing you to maintain your positions
          for long-term gains and avoid taxable events.
        </p>
      </div>

      {/* Protection Target Slider */}
      <div>
        <div className="flex items-center justify-between mb-2">
          <label className="font-medium text-ink">
            Target Protection Level
          </label>
          <span className="font-serif font-semibold text-accent-blue">
            {protectionTarget}% downside protection
          </span>
        </div>
        <input
          type="range"
          min="5"
          max="25"
          value={protectionTarget}
          onChange={(e) => setProtectionTarget(Number(e.target.value))}
          className="w-full h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer accent-accent-blue"
        />
        <div className="flex justify-between text-xs text-foreground-secondary mt-1">
          <span>5% (cheaper)</span>
          <span>15%</span>
          <span>25% (more protection)</span>
        </div>
      </div>

      {/* Strategy Selector */}
      <div>
        <h3 className="font-serif font-semibold text-ink mb-3">
          Select Hedging Strategy
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {strategies.map((strategy) => (
            <button
              key={strategy.id}
              onClick={() => setSelectedStrategy(strategy.id)}
              className={`p-4 rounded-md border text-left transition-all ${
                selectedStrategy === strategy.id
                  ? "border-accent-blue bg-accent-blue/5"
                  : "border-paper-gray hover:border-accent-blue/50"
              }`}
            >
              <div className="flex items-center justify-between mb-2">
                <span className="font-semibold text-ink">{strategy.name}</span>
                <span
                  className={`text-xs px-2 py-0.5 rounded ${
                    strategy.costType === "debit"
                      ? "bg-red-100 text-red-700"
                      : strategy.costType === "credit"
                      ? "bg-green-100 text-green-700"
                      : "bg-gray-100 text-gray-700"
                  }`}
                >
                  {strategy.costType === "debit"
                    ? "Pay premium"
                    : strategy.costType === "credit"
                    ? "Receive premium"
                    : "Low/no cost"}
                </span>
              </div>
              <p className="text-sm text-foreground-secondary mb-3">
                {strategy.description}
              </p>
              <div className="grid grid-cols-2 gap-2 text-xs">
                <div>
                  <span className="text-foreground-secondary">Max Loss:</span>
                  <span className="text-ink ml-1">{strategy.maxLoss}</span>
                </div>
                <div>
                  <span className="text-foreground-secondary">Max Gain:</span>
                  <span className="text-ink ml-1">{strategy.maxGain}</span>
                </div>
              </div>
              <div className="text-xs text-accent-blue mt-2">
                Best for: {strategy.bestFor}
              </div>
            </button>
          ))}
        </div>
      </div>

      {/* Hedge Proposals */}
      <div className="border-t border-paper-gray pt-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="font-serif font-semibold text-ink">
            Recommended Hedges
          </h3>
          <div className="text-sm text-foreground-secondary">
            Expiration: March 21, 2025 (67 days)
          </div>
        </div>

        <div className="space-y-4">
          {filteredProposals.map((proposal) => (
            <HedgeProposalCard
              key={proposal.symbol}
              proposal={proposal}
              isSelected={selectedProposals.has(proposal.symbol)}
              onToggle={() => toggleProposal(proposal.symbol)}
            />
          ))}
        </div>
      </div>

      {/* Summary */}
      {selectedProposals.size > 0 && (
        <div className="border-t border-paper-gray pt-4">
          <div className="bg-paper-cream/50 rounded-md p-4">
            <div className="flex items-center justify-between mb-4">
              <div>
                <span className="font-medium text-ink">
                  {selectedProposals.size} position{selectedProposals.size !== 1 ? "s" : ""} selected
                </span>
                <span className="text-foreground-secondary ml-2">
                  ({formatCurrency(totalProtectedValue)} protected)
                </span>
              </div>
              <button className="px-4 py-2 bg-accent-blue text-white rounded-md hover:bg-accent-blue/90 transition-colors font-medium text-sm">
                Preview Orders
              </button>
            </div>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
              <div>
                <div className="text-foreground-secondary">Total Cost</div>
                <div className="font-serif font-semibold text-ink">
                  {formatCurrency(totalHedgeCost)}
                </div>
              </div>
              <div>
                <div className="text-foreground-secondary">% of Protected Value</div>
                <div className="font-serif font-semibold text-ink">
                  {((totalHedgeCost / totalProtectedValue) * 100).toFixed(2)}%
                </div>
              </div>
              <div>
                <div className="text-foreground-secondary">Contracts</div>
                <div className="font-serif font-semibold text-ink">
                  {filteredProposals
                    .filter((p) => selectedProposals.has(p.symbol))
                    .reduce((sum, p) => sum + p.contracts.length * Math.ceil(p.sharesHeld / 100), 0)}
                </div>
              </div>
              <div>
                <div className="text-foreground-secondary">Max Protected Loss</div>
                <div className="font-serif font-semibold text-green-600">
                  {formatCurrency(
                    filteredProposals
                      .filter((p) => selectedProposals.has(p.symbol))
                      .reduce((sum, p) => sum + p.maxProtectedLoss, 0)
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Educational Section */}
      <div className="border-t border-paper-gray pt-6">
        <h3 className="font-serif font-semibold text-ink mb-4">
          Understanding Options Greeks
        </h3>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
          <div className="p-3 bg-paper-cream/50 rounded-md border border-paper-gray">
            <div className="font-medium text-ink mb-1">Delta (Δ)</div>
            <div className="text-xs text-foreground-secondary">
              Price sensitivity to underlying. -0.30 delta means option gains $0.30 for every $1 stock drops.
            </div>
          </div>
          <div className="p-3 bg-paper-cream/50 rounded-md border border-paper-gray">
            <div className="font-medium text-ink mb-1">Implied Vol (IV)</div>
            <div className="text-xs text-foreground-secondary">
              Market&apos;s expected volatility. Higher IV = more expensive options.
            </div>
          </div>
          <div className="p-3 bg-paper-cream/50 rounded-md border border-paper-gray">
            <div className="font-medium text-ink mb-1">Time Decay (Θ)</div>
            <div className="text-xs text-foreground-secondary">
              Options lose value as expiration approaches. Longer expirations cost more.
            </div>
          </div>
          <div className="p-3 bg-paper-cream/50 rounded-md border border-paper-gray">
            <div className="font-medium text-ink mb-1">Open Interest</div>
            <div className="text-xs text-foreground-secondary">
              Number of outstanding contracts. Higher = better liquidity and tighter spreads.
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function HedgeProposalCard({
  proposal,
  isSelected,
  onToggle,
}: {
  proposal: HedgeProposal;
  isSelected: boolean;
  onToggle: () => void;
}) {
  const [expanded, setExpanded] = useState(false);
  const strategyInfo = strategies.find((s) => s.id === proposal.strategy);

  return (
    <div
      className={`border rounded-md overflow-hidden transition-all ${
        isSelected
          ? "border-accent-blue bg-accent-blue/5"
          : "border-paper-gray"
      }`}
    >
      {/* Header */}
      <div
        onClick={onToggle}
        className="p-4 cursor-pointer hover:bg-paper-cream/30 transition-colors"
      >
        <div className="flex items-start justify-between">
          <div className="flex items-start gap-3">
            <div
              className={`mt-1 w-5 h-5 rounded border-2 flex items-center justify-center ${
                isSelected
                  ? "border-accent-blue bg-accent-blue"
                  : "border-paper-gray"
              }`}
            >
              {isSelected && (
                <svg className="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
                  <path
                    fillRule="evenodd"
                    d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                    clipRule="evenodd"
                  />
                </svg>
              )}
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-semibold text-ink text-lg">{proposal.symbol}</span>
                <span className="text-xs px-2 py-0.5 bg-blue-100 text-blue-700 rounded">
                  {strategyInfo?.name}
                </span>
              </div>
              <div className="text-sm text-foreground-secondary">
                {proposal.sharesHeld} shares @ {formatCurrency(proposal.currentPrice)}
              </div>
              <div className="text-sm text-foreground-secondary">
                Position value: {formatCurrency(proposal.positionValue)}
              </div>
            </div>
          </div>

          <div className="text-right">
            <div className="font-serif text-lg font-semibold text-ink">
              {formatCurrency(proposal.totalCost)}
            </div>
            <div className="text-xs text-foreground-secondary">
              ({((proposal.totalCost / proposal.positionValue) * 100).toFixed(1)}% of position)
            </div>
            <div className="text-sm text-green-600 font-medium mt-1">
              Max loss: {formatCurrency(proposal.maxProtectedLoss)}
            </div>
          </div>
        </div>

        {/* Quick Stats */}
        <div className="grid grid-cols-3 gap-4 mt-4 pt-3 border-t border-paper-gray/50">
          <div>
            <div className="text-xs text-foreground-secondary">Protection</div>
            <div className="font-medium text-ink">{proposal.protectionLevel}%</div>
          </div>
          <div>
            <div className="text-xs text-foreground-secondary">Breakeven</div>
            <div className="font-medium text-ink">{formatCurrency(proposal.breakeven)}</div>
          </div>
          <div>
            <div className="text-xs text-foreground-secondary">Contracts</div>
            <div className="font-medium text-ink">
              {Math.ceil(proposal.sharesHeld / 100)} × {proposal.contracts.length}
            </div>
          </div>
        </div>
      </div>

      {/* Expand Toggle */}
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full px-4 py-2 bg-paper-cream/50 text-sm text-foreground-secondary hover:text-ink transition-colors flex items-center justify-center gap-1"
      >
        {expanded ? "Hide" : "Show"} contract details
        <svg
          className={`w-4 h-4 transition-transform ${expanded ? "rotate-180" : ""}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {/* Contract Details */}
      {expanded && (
        <div className="px-4 pb-4 bg-paper-cream/30">
          <div className="space-y-3">
            {proposal.contracts.map((contract, idx) => (
              <div
                key={idx}
                className="p-3 bg-white rounded-md border border-paper-gray"
              >
                <div className="flex items-center justify-between mb-2">
                  <div className="flex items-center gap-2">
                    <span
                      className={`px-2 py-0.5 text-xs font-medium rounded ${
                        contract.action === "BUY"
                          ? "bg-green-100 text-green-700"
                          : "bg-red-100 text-red-700"
                      }`}
                    >
                      {contract.action}
                    </span>
                    <span
                      className={`px-2 py-0.5 text-xs font-medium rounded ${
                        contract.type === "PUT"
                          ? "bg-purple-100 text-purple-700"
                          : "bg-blue-100 text-blue-700"
                      }`}
                    >
                      {contract.type}
                    </span>
                    <span className="font-semibold text-ink">
                      ${contract.strike} Strike
                    </span>
                  </div>
                  <div className="font-medium text-ink">
                    ${contract.premium.toFixed(2)}/share
                  </div>
                </div>
                <div className="grid grid-cols-4 gap-2 text-xs">
                  <div>
                    <span className="text-foreground-secondary">Exp:</span>
                    <span className="text-ink ml-1">{contract.expiration}</span>
                  </div>
                  <div>
                    <span className="text-foreground-secondary">Delta:</span>
                    <span className="text-ink ml-1">{contract.delta.toFixed(2)}</span>
                  </div>
                  <div>
                    <span className="text-foreground-secondary">IV:</span>
                    <span className="text-ink ml-1">{contract.impliedVol}%</span>
                  </div>
                  <div>
                    <span className="text-foreground-secondary">OI:</span>
                    <span className="text-ink ml-1">{contract.openInterest.toLocaleString()}</span>
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* P&L Diagram placeholder */}
          <div className="mt-4 p-4 bg-white rounded-md border border-paper-gray">
            <div className="text-sm font-medium text-ink mb-2">Payoff at Expiration</div>
            <div className="h-32 flex items-center justify-center text-foreground-secondary text-sm">
              [P&L Diagram Visualization]
            </div>
            <div className="flex justify-between text-xs text-foreground-secondary mt-2">
              <span>Stock at ${(proposal.currentPrice * 0.8).toFixed(0)}</span>
              <span>Current: ${proposal.currentPrice.toFixed(0)}</span>
              <span>Stock at ${(proposal.currentPrice * 1.2).toFixed(0)}</span>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
