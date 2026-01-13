export default function DashboardPage() {
  const stats = [
    {
      label: "Total Value",
      value: "$284,523",
      change: "+18.4% YTD",
      positive: true,
    },
    {
      label: "Holdings",
      value: "12",
      change: "+2 this month",
      positive: true,
    },
    {
      label: "Today's Change",
      value: "+$1,247",
      change: "+0.44%",
      positive: true,
    },
    {
      label: "Cash",
      value: "$12,840",
      change: "4.5% of portfolio",
      positive: null,
    },
  ];

  const holdings = [
    {
      name: "Apple Inc.",
      symbol: "AAPL",
      shares: "240",
      price: "$178.50",
      value: "$42,840",
      allocation: "15.1%",
      change: "+2.4%",
      positive: true,
    },
    {
      name: "Microsoft Corporation",
      symbol: "MSFT",
      shares: "180",
      price: "$378.91",
      value: "$68,204",
      allocation: "24.0%",
      change: "+1.8%",
      positive: true,
    },
    {
      name: "Vanguard Total Stock Market ETF",
      symbol: "VTI",
      shares: "350",
      price: "$242.15",
      value: "$84,753",
      allocation: "29.8%",
      change: "+1.2%",
      positive: true,
    },
    {
      name: "iShares Core U.S. Aggregate Bond ETF",
      symbol: "AGG",
      shares: "420",
      price: "$98.45",
      value: "$41,349",
      allocation: "14.5%",
      change: "-0.3%",
      positive: false,
    },
    {
      name: "Amazon.com Inc.",
      symbol: "AMZN",
      shares: "95",
      price: "$178.25",
      value: "$16,934",
      allocation: "6.0%",
      change: "+3.1%",
      positive: true,
    },
    {
      name: "Alphabet Inc. Class A",
      symbol: "GOOGL",
      shares: "110",
      price: "$141.80",
      value: "$15,598",
      allocation: "5.5%",
      change: "+1.5%",
      positive: true,
    },
  ];

  return (
    <div className="min-h-screen">
      <div className="max-w-7xl mx-auto px-4 py-6 sm:p-8">
        {/* Page Header */}
        <header className="mb-statement">
          <h1 className="font-serif text-3xl font-semibold text-ink mb-2">
            Portfolio Statement
          </h1>
          <p className="text-base text-foreground-secondary">
            Investment Account - {new Date().toLocaleDateString("en-US", {
              month: "long",
              day: "numeric",
              year: "numeric",
            })}
          </p>
        </header>

        {/* Stat Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-section">
          {stats.map((stat) => (
            <div
              key={stat.label}
              className="bg-white border border-paper-gray rounded-md shadow-paper p-6"
            >
              <div className="text-sm font-serif text-accent-blue mb-2">
                {stat.label}
              </div>
              <div className="font-serif text-xl font-semibold text-ink mb-1">
                {stat.value}
              </div>
              <div
                className={`text-sm ${
                  stat.positive === true
                    ? "text-green-700"
                    : stat.positive === false
                    ? "text-red-700"
                    : "text-foreground-secondary"
                }`}
              >
                {stat.change}
              </div>
            </div>
          ))}
        </div>

        {/* Holdings */}
        <div className="bg-white border border-paper-gray rounded-md shadow-paper p-4 sm:p-statement">
          <div className="flex justify-between items-center mb-6">
            <h2 className="font-serif text-lg font-semibold text-ink">
              Holdings
            </h2>
            <button className="px-4 sm:px-6 py-2 border border-paper-gray text-ink rounded-sm hover:bg-paper-gray transition-all font-sans font-semibold text-sm">
              Import CSV
            </button>
          </div>

          {/* Mobile Card Layout */}
          <div className="md:hidden space-y-4">
            {holdings.map((holding) => (
              <div
                key={holding.symbol}
                className="border border-paper-gray rounded-md p-4"
              >
                <div className="flex justify-between items-start mb-3">
                  <div>
                    <div className="font-semibold text-ink">{holding.name}</div>
                    <div className="text-sm text-foreground-secondary">
                      {holding.symbol}
                    </div>
                  </div>
                  <div
                    className={`font-semibold ${
                      holding.positive ? "text-green-700" : "text-red-700"
                    }`}
                  >
                    {holding.change}
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-3 text-sm">
                  <div>
                    <div className="text-foreground-secondary">Value</div>
                    <div className="font-serif font-semibold text-ink">
                      {holding.value}
                    </div>
                  </div>
                  <div>
                    <div className="text-foreground-secondary">Allocation</div>
                    <div className="font-semibold text-ink">
                      {holding.allocation}
                    </div>
                  </div>
                  <div>
                    <div className="text-foreground-secondary">Shares</div>
                    <div className="text-ink">{holding.shares}</div>
                  </div>
                  <div>
                    <div className="text-foreground-secondary">Price</div>
                    <div className="text-ink">{holding.price}</div>
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* Desktop Table Layout */}
          <div className="hidden md:block overflow-x-auto">
            <table className="w-full">
              <thead className="border-b-2 border-ink">
                <tr className="text-left">
                  <th className="pb-3 font-serif font-semibold text-ink">
                    Security
                  </th>
                  <th className="pb-3 font-serif font-semibold text-ink text-right">
                    Shares
                  </th>
                  <th className="pb-3 font-serif font-semibold text-ink text-right">
                    Price
                  </th>
                  <th className="pb-3 font-serif font-semibold text-ink text-right">
                    Value
                  </th>
                  <th className="pb-3 font-serif font-semibold text-ink text-right">
                    Allocation
                  </th>
                  <th className="pb-3 font-serif font-semibold text-ink text-right">
                    Change
                  </th>
                </tr>
              </thead>
              <tbody>
                {holdings.map((holding) => (
                  <tr
                    key={holding.symbol}
                    className="border-b border-paper-gray hover:bg-paper-cream/50 transition-all"
                  >
                    <td className="py-3">
                      <div className="font-semibold text-ink">
                        {holding.name}
                      </div>
                      <div className="text-sm text-foreground-secondary">
                        {holding.symbol}
                      </div>
                    </td>
                    <td className="py-3 text-right text-foreground-secondary">
                      {holding.shares}
                    </td>
                    <td className="py-3 text-right text-foreground-secondary">
                      {holding.price}
                    </td>
                    <td className="py-3 text-right font-serif font-semibold text-ink">
                      {holding.value}
                    </td>
                    <td className="py-3 text-right text-foreground-secondary">
                      {holding.allocation}
                    </td>
                    <td
                      className={`py-3 text-right font-semibold ${
                        holding.positive ? "text-green-700" : "text-red-700"
                      }`}
                    >
                      {holding.change}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}
