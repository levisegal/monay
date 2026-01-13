"use client";

import { PortfolioSwitcher } from "./PortfolioSwitcher";

interface Portfolio {
  id: string;
  name: string;
  totalValue: number;
}

interface HeaderProps {
  portfolios: Portfolio[];
  selectedPortfolioId: string;
  onPortfolioSelect: (id: string) => void;
}

export function Header({
  portfolios,
  selectedPortfolioId,
  onPortfolioSelect,
}: HeaderProps) {
  return (
    <header className="bg-white border-b border-paper-gray">
      <div className="max-w-7xl mx-auto px-4 sm:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo / Brand */}
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-md bg-ink flex items-center justify-center">
              <span className="text-white font-serif font-bold text-lg">M</span>
            </div>
            <span className="font-serif text-xl font-semibold text-ink hidden sm:block">
              Monay
            </span>
          </div>

          {/* Right side - Portfolio Switcher */}
          <PortfolioSwitcher
            portfolios={portfolios}
            selectedId={selectedPortfolioId}
            onSelect={onPortfolioSelect}
          />
        </div>
      </div>
    </header>
  );
}
