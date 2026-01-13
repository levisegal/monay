# Paper Portfolio Design System

Print-inspired design system for financial applications. Sophisticated, tactile, and trustworthy.

## Quick Start

### Tailwind CSS

```bash
# Use the Paper Portfolio Tailwind config
cp tailwind.config.paper-portfolio.js tailwind.config.js
```

### CSS Variables

```html
<!-- Import the CSS variables -->
<link rel="stylesheet" href="./paper-portfolio.css">
```

## Design Tokens

### Colors

| Token | Value | Usage |
|-------|-------|-------|
| `paper-cream` | #FBF8F3 | Main background - paper color |
| `paper-gray` | #E8E4DD | Borders, dividers |
| `paper-warm-gray` | #6B6456 | Secondary text |
| `ink` | #1A1A1A | Primary text - ink black |
| `accent-blue` | #2C5F77 | Links, highlights |

**Tailwind:** `bg-paper-cream`, `text-ink`, `border-paper-gray`
**CSS:** `var(--color-paper-cream)`, `var(--color-ink)`, `var(--color-border)`

### Typography

**Font Families:**
- **Serif (Lora)**: Headings, numbers, key elements
- **Sans (System)**: Body text

**Font Sizes:**
- xs: 14px (small labels)
- sm: 16px (body text minimum)
- base: 18px (body text)
- lg: 24px (subheadings)
- xl: 32px (large values)
- 2xl: 40px (page headings)
- 3xl: 48px (hero text)

**Examples:**

```tsx
// Tailwind
<h1 className="font-serif text-2xl font-semibold text-ink">Portfolio Statement</h1>
<p className="font-sans text-base text-foreground-secondary">As of January 12, 2026</p>
<div className="font-serif text-xl font-semibold text-ink">$284,523</div>

// CSS
<h1 style="font-family: var(--font-serif); font-size: var(--font-size-2xl); color: var(--color-ink);">
  Portfolio Statement
</h1>
```

### Spacing

**Base Unit:** 8px

| Token | Value | Usage |
|-------|-------|-------|
| `statement` | 32px | Generous spacing between major sections |
| `section` | 24px | Between related groups |
| Standard scale | 8/16/24/32/40/48/64px | General spacing |

**Examples:**

```tsx
// Tailwind
<div className="p-statement space-y-section">
<div className="p-8 mb-6">

// CSS
<div style="padding: var(--spacing-statement); margin-bottom: var(--spacing-section);">
```

### Shadows

Tactile, paper-like shadows that suggest depth without drama.

```tsx
// Tailwind
<div className="shadow-paper">        // Default paper shadow
<div className="shadow-paper-lg">     // Larger elevation
<div className="hover:shadow-paper-hover">  // Hover state

// CSS
<div style="box-shadow: var(--shadow-paper);">
```

### Border Radius

Subtle, professional corners.

```tsx
// Tailwind
<div className="rounded-md">   // 8px - cards, containers
<div className="rounded-sm">   // 4px - inputs, buttons
<div className="rounded-lg">   // 12px - large cards

// CSS
<div style="border-radius: var(--radius-md);">
```

## Component Patterns

### Card / Container

```tsx
// Tailwind
<div className="bg-white border border-paper-gray rounded-md shadow-paper p-statement">
  <h2 className="font-serif text-lg font-semibold text-ink mb-4">
    Holdings Statement
  </h2>
  {/* content */}
</div>

// CSS
<div class="card">
  <h2>Holdings Statement</h2>
  {/* content */}
</div>
```

### Stat Card

```tsx
<div className="bg-white border border-paper-gray rounded-md shadow-paper p-6">
  <div className="text-sm font-serif text-accent-blue mb-2">
    Total Value
  </div>
  <div className="font-serif text-xl font-semibold text-ink mb-1">
    $284,523
  </div>
  <div className="text-sm text-foreground-secondary">
    +18.4% YTD
  </div>
</div>
```

### Table (Statement Style)

```tsx
<div className="bg-white border border-paper-gray rounded-md shadow-paper p-statement">
  <h2 className="font-serif text-lg font-semibold text-ink mb-6">
    Holdings
  </h2>
  <table className="w-full">
    <thead className="border-b-2 border-ink">
      <tr className="text-left">
        <th className="pb-3 font-serif font-semibold text-ink">Security</th>
        <th className="pb-3 font-serif font-semibold text-ink text-right">Shares</th>
        <th className="pb-3 font-serif font-semibold text-ink text-right">Value</th>
        <th className="pb-3 font-serif font-semibold text-ink text-right">Allocation</th>
      </tr>
    </thead>
    <tbody>
      <tr className="border-b border-paper-gray">
        <td className="py-3">
          <div className="font-semibold text-ink">Apple Inc.</div>
          <div className="text-sm text-foreground-secondary">AAPL</div>
        </td>
        <td className="py-3 text-right">240</td>
        <td className="py-3 text-right font-serif font-semibold text-ink">
          $42,840
        </td>
        <td className="py-3 text-right">15.1%</td>
      </tr>
      {/* more rows */}
    </tbody>
  </table>
</div>
```

### Button

```tsx
// Primary action
<button className="px-6 py-2 bg-ink text-white rounded-sm
                   hover:opacity-90 transition-all font-sans font-semibold">
  Export Statement
</button>

// Secondary action
<button className="px-6 py-2 border border-paper-gray text-ink rounded-sm
                   hover:bg-paper-gray transition-all font-sans font-semibold">
  Import CSV
</button>
```

### Page Layout

```tsx
<div className="paper-bg min-h-screen">
  <div className="max-w-6xl mx-auto p-8">
    {/* Page header */}
    <header className="mb-statement">
      <h1 className="font-serif text-3xl font-semibold text-ink mb-2">
        Portfolio Statement
      </h1>
      <p className="text-base text-foreground-secondary">
        Investment Account - January 2026
      </p>
    </header>

    {/* Content sections */}
    <div className="space-y-section">
      {/* Stat cards, tables, etc. */}
    </div>
  </div>
</div>
```

## Anti-Patterns (What NOT to Do)

❌ **Don't use bright, saturated colors**
```tsx
// Bad
<div className="bg-blue-600 text-white">

// Good
<div className="bg-white border border-paper-gray text-ink">
```

❌ **Don't use large border radius**
```tsx
// Bad
<div className="rounded-3xl">

// Good
<div className="rounded-md">
```

❌ **Don't use flashy animations**
```tsx
// Bad
<div className="animate-bounce transition-all duration-700">

// Good
<div className="transition-all duration-200 hover:shadow-paper-hover">
```

❌ **Don't use sans-serif for financial numbers**
```tsx
// Bad
<div className="font-sans text-2xl">$284,523</div>

// Good
<div className="font-serif text-2xl font-semibold">$284,523</div>
```

❌ **Don't skip the paper texture on backgrounds**
```tsx
// Bad
<body className="bg-paper-cream">

// Good
<body className="paper-bg">
```

## Motion & Interaction

**Subtle, refined, professional.**

- Duration: 200ms (fast), 300ms (normal)
- Easing: ease-out
- Hover: Slight lift (shadow increase), subtle opacity change
- No bouncing, spinning, or dramatic effects
- Transitions should enhance, not distract

## Accessibility

- Minimum text size: 16px (sm)
- Color contrast: All combinations meet WCAG AA
- Focus states: Visible outline on all interactive elements
- Touch targets: Minimum 44x44px on mobile

## Google Fonts Import

```html
<!-- Add to <head> -->
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Lora:wght@400;600&display=swap" rel="stylesheet">
```

Or via CSS:

```css
@import url('https://fonts.googleapis.com/css2?family=Lora:wght@400;600&display=swap');
```

## Next.js Setup

```tsx
// app/layout.tsx
import { Lora } from 'next/font/google'
import './paper-portfolio.css'

const lora = Lora({
  subsets: ['latin'],
  weight: ['400', '600'],
  variable: '--font-serif',
})

export default function RootLayout({ children }) {
  return (
    <html lang="en" className={lora.variable}>
      <body className="paper-portfolio">{children}</body>
    </html>
  )
}
```

## React Setup

```tsx
// main.tsx or App.tsx
import './paper-portfolio.css'

function App() {
  return (
    <div className="paper-portfolio">
      {/* Your app */}
    </div>
  )
}
```

---

**Ready to build!** This design system provides a sophisticated, trustworthy aesthetic perfect for financial applications like Wealthfolio.
