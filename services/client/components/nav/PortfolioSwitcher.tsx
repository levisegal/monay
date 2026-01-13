"use client";

import { useState, useRef, useEffect } from "react";

interface Portfolio {
  id: string;
  name: string;
  totalValue: number;
}

interface PortfolioSwitcherProps {
  portfolios: Portfolio[];
  selectedId: string;
  onSelect: (id: string) => void;
}

export function PortfolioSwitcher({
  portfolios,
  selectedId,
  onSelect,
}: PortfolioSwitcherProps) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const selectedPortfolio = portfolios.find((p) => p.id === selectedId);

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const formatValue = (value: number) =>
    new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: "USD",
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(value);

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-3 px-3 py-2 rounded-md hover:bg-paper-cream transition-colors"
      >
        {/* Avatar/Icon */}
        <div className="w-8 h-8 rounded-full bg-accent-blue/10 flex items-center justify-center">
          <span className="text-accent-blue font-semibold text-sm">
            {selectedPortfolio?.name.charAt(0).toUpperCase()}
          </span>
        </div>
        <div className="text-left hidden sm:block">
          <div className="text-sm font-semibold text-ink">
            {selectedPortfolio?.name}
          </div>
          <div className="text-xs text-foreground-secondary">
            {formatValue(selectedPortfolio?.totalValue || 0)}
          </div>
        </div>
        <svg
          className={`w-4 h-4 text-foreground-secondary transition-transform ${isOpen ? "rotate-180" : ""}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-2 w-64 bg-white border border-paper-gray rounded-md shadow-paper-hover z-50">
          <div className="p-2">
            <div className="text-xs font-medium text-foreground-secondary px-2 py-1">
              PORTFOLIOS
            </div>
            {portfolios.map((portfolio) => (
              <button
                key={portfolio.id}
                onClick={() => {
                  onSelect(portfolio.id);
                  setIsOpen(false);
                }}
                className={`w-full flex items-center gap-3 px-2 py-2 rounded-md text-left transition-colors ${
                  portfolio.id === selectedId
                    ? "bg-accent-blue/10"
                    : "hover:bg-paper-cream"
                }`}
              >
                <div className="w-8 h-8 rounded-full bg-accent-blue/10 flex items-center justify-center">
                  <span className="text-accent-blue font-semibold text-sm">
                    {portfolio.name.charAt(0).toUpperCase()}
                  </span>
                </div>
                <div>
                  <div className="text-sm font-semibold text-ink">
                    {portfolio.name}
                  </div>
                  <div className="text-xs text-foreground-secondary">
                    {formatValue(portfolio.totalValue)}
                  </div>
                </div>
                {portfolio.id === selectedId && (
                  <svg className="w-4 h-4 text-accent-blue ml-auto" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                  </svg>
                )}
              </button>
            ))}
          </div>
          <div className="border-t border-paper-gray p-2">
            <button className="w-full flex items-center gap-2 px-2 py-2 text-sm text-foreground-secondary hover:text-ink hover:bg-paper-cream rounded-md transition-colors">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              Add Portfolio
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
