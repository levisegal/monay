"use client";

import { useState, useRef, useEffect } from "react";

interface Account {
  id: string;
  name: string;
  institution: string;
  accountNumber: string;
  type: string;
  balance: number;
}

interface AccountSelectorProps {
  accounts: Account[];
  selectedId: string | null;
  onSelect: (id: string | null) => void;
}

export function AccountSelector({
  accounts,
  selectedId,
  onSelect,
}: AccountSelectorProps) {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const selectedAccount = accounts.find((a) => a.id === selectedId);
  const displayText = selectedId === null
    ? `All (${accounts.length} accounts)`
    : selectedAccount?.name || "Select account";

  return (
    <div className="flex items-center gap-2 text-sm" ref={dropdownRef}>
      <span className="text-foreground-secondary">View:</span>

      <div className="relative">
        <button
          onClick={() => setIsOpen(!isOpen)}
          className="flex items-center gap-2 px-3 py-1.5 bg-white border border-paper-gray rounded-md hover:border-ink transition-colors"
        >
          <span className="text-ink font-medium">{displayText}</span>
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
          <div className="absolute top-full left-0 mt-1 w-64 bg-white border border-paper-gray rounded-md shadow-lg z-50 max-h-80 overflow-y-auto">
            <button
              onClick={() => {
                onSelect(null);
                setIsOpen(false);
              }}
              className={`w-full text-left px-3 py-2 hover:bg-paper-cream transition-colors ${
                selectedId === null ? "bg-paper-cream font-medium" : ""
              }`}
            >
              All ({accounts.length} accounts)
            </button>

            <div className="border-t border-paper-gray" />

            {accounts.map((account) => (
              <button
                key={account.id}
                onClick={() => {
                  onSelect(account.id);
                  setIsOpen(false);
                }}
                className={`w-full text-left px-3 py-2 hover:bg-paper-cream transition-colors ${
                  selectedId === account.id ? "bg-paper-cream font-medium" : ""
                }`}
              >
                <div className="text-ink">{account.name}</div>
                <div className="text-xs text-foreground-secondary">{account.institution}</div>
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
