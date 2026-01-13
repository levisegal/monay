"use client";

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
  return (
    <div className="flex items-center gap-1 text-sm">
      <span className="text-foreground-secondary mr-1">View:</span>

      {/* All */}
      <button
        onClick={() => onSelect(null)}
        className={`px-2 py-0.5 rounded transition-colors ${
          selectedId === null
            ? "bg-ink text-white"
            : "text-foreground-secondary hover:text-ink"
        }`}
      >
        All
      </button>

      <span className="text-paper-gray">·</span>

      {/* Individual accounts */}
      {accounts.map((account, idx) => (
        <span key={account.id} className="flex items-center">
          <button
            onClick={() => onSelect(account.id)}
            className={`px-2 py-0.5 rounded transition-colors ${
              selectedId === account.id
                ? "bg-ink text-white"
                : "text-foreground-secondary hover:text-ink"
            }`}
            title={`${account.type} ••${account.accountNumber}`}
          >
            {account.name}
            <span className={selectedId === account.id ? "text-white/60" : "text-foreground-secondary/60"}> ({account.institution.split(" ")[0]})</span>
          </button>
          {idx < accounts.length - 1 && (
            <span className="text-paper-gray ml-1">·</span>
          )}
        </span>
      ))}

    </div>
  );
}
