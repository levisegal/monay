"use client";

import { useState } from "react";
import { AllocationComparison } from "./AllocationComparison";
import { SuggestedTrades } from "./SuggestedTrades";
import { RiskAnalysis } from "./RiskAnalysis";
import { OptionsHedging } from "./OptionsHedging";

type TabId = "allocation" | "trades" | "risk" | "hedging";

interface Tab {
  id: TabId;
  label: string;
  description: string;
}

const tabs: Tab[] = [
  { id: "allocation", label: "Allocation", description: "Current vs target" },
  { id: "trades", label: "Trades", description: "Suggested actions" },
  { id: "risk", label: "Risk", description: "Exposure analysis" },
  { id: "hedging", label: "Hedging", description: "Options strategies" },
];

export type RiskProfile = "conservative" | "moderate" | "aggressive";

interface RiskProfileOption {
  id: RiskProfile;
  label: string;
  description: string;
  stockAllocation: number;
  bondAllocation: number;
}

const riskProfiles: RiskProfileOption[] = [
  {
    id: "conservative",
    label: "Conservative",
    description: "Lower risk, stable returns",
    stockAllocation: 40,
    bondAllocation: 60,
  },
  {
    id: "moderate",
    label: "Moderate",
    description: "Balanced growth & stability",
    stockAllocation: 60,
    bondAllocation: 40,
  },
  {
    id: "aggressive",
    label: "Aggressive",
    description: "Higher risk, growth focus",
    stockAllocation: 80,
    bondAllocation: 20,
  },
];

export function RebalancingSection() {
  const [activeTab, setActiveTab] = useState<TabId>("allocation");
  const [targetProfile, setTargetProfile] = useState<RiskProfile>("moderate");

  const selectedProfile = riskProfiles.find((p) => p.id === targetProfile)!;

  return (
    <div className="bg-white border border-paper-gray rounded-md shadow-paper">
      {/* Header */}
      <div className="p-4 sm:p-6 border-b border-paper-gray">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h2 className="font-serif text-lg font-semibold text-ink">
              Portfolio Rebalancing
            </h2>
            <p className="text-sm text-foreground-secondary">
              Analyze and optimize your portfolio allocation
            </p>
          </div>

          {/* Risk Profile Selector */}
          <div className="flex items-center gap-2">
            <span className="text-sm text-foreground-secondary">Target:</span>
            <div className="flex rounded-md border border-paper-gray overflow-hidden">
              {riskProfiles.map((profile) => (
                <button
                  key={profile.id}
                  onClick={() => setTargetProfile(profile.id)}
                  className={`px-3 py-1.5 text-sm font-medium transition-colors ${
                    targetProfile === profile.id
                      ? "bg-accent-blue text-white"
                      : "bg-white text-foreground-secondary hover:bg-paper-cream"
                  }`}
                >
                  {profile.label}
                </button>
              ))}
            </div>
          </div>
        </div>

        {/* Target Profile Summary */}
        <div className="mt-4 p-3 bg-paper-cream/50 rounded-md border border-paper-gray">
          <div className="flex items-center justify-between">
            <div>
              <span className="font-medium text-ink">{selectedProfile.label}</span>
              <span className="text-foreground-secondary mx-2">·</span>
              <span className="text-sm text-foreground-secondary">
                {selectedProfile.description}
              </span>
            </div>
            <div className="flex items-center gap-4 text-sm">
              <span className="text-foreground-secondary">
                <span className="font-medium text-ink">{selectedProfile.stockAllocation}%</span> Stocks
              </span>
              <span className="text-foreground-secondary">
                <span className="font-medium text-ink">{selectedProfile.bondAllocation}%</span> Bonds
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Tab Navigation */}
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
                <div className="text-xs opacity-70 hidden sm:block">{tab.description}</div>
              </div>
            </button>
          ))}
        </div>
      </div>

      {/* Tab Content */}
      <div className="p-4 sm:p-6">
        {activeTab === "allocation" && (
          <AllocationComparison targetProfile={selectedProfile} />
        )}
        {activeTab === "trades" && (
          <SuggestedTrades targetProfile={selectedProfile} />
        )}
        {activeTab === "risk" && <RiskAnalysis />}
        {activeTab === "hedging" && <OptionsHedging />}
      </div>
    </div>
  );
}
