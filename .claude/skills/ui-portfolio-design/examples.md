# Visual Examples - UI Portfolio Design Directions

This document provides concrete code examples showing how each design direction looks in practice.

---

## 1. Precision & Density - Dashboard Card

**Characteristics**: Tight spacing, monospace for data, minimal shadows, functional blue

```tsx
<div className="rounded border border-slate-200 bg-white p-3 shadow-sm">
  <div className="flex items-center justify-between mb-2">
    <h3 className="text-sm font-semibold text-slate-900">Revenue</h3>
    <span className="text-xs text-slate-500">MTD</span>
  </div>

  <div className="font-mono text-2xl font-semibold text-blue-600 mb-1">
    $482,394.12
  </div>

  <div className="flex items-center text-xs text-emerald-600">
    <ArrowUpIcon className="w-3 h-3 mr-1" />
    <span className="font-mono">+12.4%</span>
  </div>
</div>
```

**CSS Variables**:
```css
:root {
  --color-primary: #2563eb;  /* blue-600 */
  --spacing-compact: 8px;
  --spacing-default: 12px;
  --border-radius: 4px;
  --shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}
```

---

## 2. Warmth & Approachability - Profile Card

**Characteristics**: Rounded corners, warm colors, generous spacing, friendly

```tsx
<div className="rounded-2xl bg-white p-6 shadow-lg">
  <div className="flex items-center space-x-4 mb-4">
    <img
      src="/avatar.jpg"
      className="w-16 h-16 rounded-full ring-2 ring-orange-200"
    />
    <div>
      <h3 className="text-xl font-semibold text-slate-900">Sarah Johnson</h3>
      <p className="text-base text-slate-600">Product Designer</p>
    </div>
  </div>

  <p className="text-base text-slate-700 leading-relaxed mb-6">
    Creating delightful experiences for users worldwide. Coffee enthusiast ☕
  </p>

  <button className="w-full rounded-full bg-orange-500 hover:bg-orange-600
                     text-white font-medium py-3 px-6 transition-all duration-300
                     hover:scale-102 shadow-md">
    Send Message
  </button>
</div>
```

**CSS Variables**:
```css
:root {
  --color-primary: #f97316;  /* orange-500 */
  --color-secondary: #ec4899;  /* pink-500 */
  --spacing-default: 16px;
  --spacing-generous: 24px;
  --border-radius: 12px;
  --shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}
```

---

## 3. Sophistication & Trust - Form Input

**Characteristics**: Refined borders, professional colors, balanced spacing

```tsx
<div className="space-y-6">
  <div>
    <label className="block text-sm font-medium text-slate-700 mb-2">
      Account Number
    </label>
    <input
      type="text"
      className="w-full h-10 px-4 rounded-md border border-slate-300
                 focus:border-blue-800 focus:ring-2 focus:ring-blue-800/20
                 transition-colors duration-150"
      placeholder="Enter account number"
    />
    <p className="mt-2 text-sm text-slate-600">
      Your 10-digit account identifier
    </p>
  </div>

  <button className="w-full h-10 px-5 rounded-md bg-blue-800 hover:bg-blue-900
                     text-white font-medium transition-colors duration-150
                     shadow-sm">
    Verify Account
  </button>
</div>
```

**CSS Variables**:
```css
:root {
  --color-primary: #1e40af;  /* blue-800 */
  --color-secondary: #334155;  /* slate-700 */
  --spacing-default: 16px;
  --border-radius: 6px;
  --shadow: 0 1px 3px rgba(0, 0, 0, 0.06),
            0 4px 12px rgba(0, 0, 0, 0.04);
}
```

---

## 4. Boldness & Clarity - Hero Section

**Characteristics**: Large typography, bold colors, dramatic spacing, strong shadows

```tsx
<section className="py-20 px-8 bg-gradient-to-br from-violet-600 to-fuchsia-500">
  <div className="max-w-4xl mx-auto text-center">
    <h1 className="text-7xl font-extrabold text-white leading-tight tracking-tight mb-6">
      Transform Your
      <br />
      Business Today
    </h1>

    <p className="text-2xl text-violet-50 leading-relaxed mb-12 max-w-2xl mx-auto">
      The all-in-one platform for modern teams to collaborate,
      ship faster, and achieve more.
    </p>

    <div className="flex justify-center gap-4">
      <button className="h-14 px-8 rounded-xl bg-white text-violet-600
                         font-bold text-lg hover:scale-105 transition-transform
                         duration-200 shadow-2xl">
        Start Free Trial
      </button>

      <button className="h-14 px-8 rounded-xl bg-transparent border-3 border-white
                         text-white font-bold text-lg hover:bg-white/10
                         transition-colors duration-200">
        Watch Demo
      </button>
    </div>
  </div>
</section>
```

**CSS Variables**:
```css
:root {
  --color-primary: #7c3aed;  /* violet-600 */
  --color-accent: #f59e0b;  /* amber-500 */
  --spacing-large: 48px;
  --spacing-dramatic: 80px;
  --border-radius: 12px;
  --shadow: 0 8px 24px rgba(0, 0, 0, 0.12),
            0 4px 8px rgba(0, 0, 0, 0.08);
}
```

---

## 5. Utility & Function - Code Block

**Characteristics**: Monospace throughout, minimal motion, terminal-inspired

```tsx
<div className="rounded bg-[#1e1e1e] p-4 border border-gray-700">
  <div className="flex items-center justify-between mb-3">
    <span className="font-mono text-xs text-gray-400">main.go</span>
    <button className="px-2 py-1 rounded bg-gray-800 hover:bg-gray-700
                       font-mono text-xs text-gray-300 transition-colors duration-100">
      Copy
    </button>
  </div>

  <pre className="font-mono text-sm text-gray-100 overflow-x-auto">
    <code>{`func main() {
  fmt.Println("Hello, World!")
}`}</code>
  </pre>

  <div className="mt-3 pt-3 border-t border-gray-700">
    <div className="flex items-center gap-2 font-mono text-xs">
      <span className="text-emerald-500">✓</span>
      <span className="text-gray-400">Build successful</span>
      <span className="text-gray-500">|</span>
      <span className="text-gray-400">0.342s</span>
    </div>
  </div>
</div>
```

**CSS Variables**:
```css
:root {
  --color-primary: #3b82f6;  /* blue-500 */
  --bg-primary: #1e1e1e;
  --bg-secondary: #2d2d2d;
  --spacing-compact: 8px;
  --border-radius: 4px;
  --font-mono: 'JetBrains Mono', monospace;
}
```

---

## 6. Data & Analysis - Data Table

**Characteristics**: Clean axes, colorblind-safe, sortable headers, right-aligned numbers

```tsx
<div className="rounded-md border border-slate-200 overflow-hidden">
  <table className="w-full">
    <thead className="bg-slate-50 border-b border-slate-200">
      <tr>
        <th className="px-4 py-3 text-left text-xs font-semibold text-slate-700 uppercase tracking-wider">
          Product
        </th>
        <th className="px-4 py-3 text-right text-xs font-semibold text-slate-700 uppercase tracking-wider">
          <button className="hover:text-blue-600 transition-colors">
            Revenue ↓
          </button>
        </th>
        <th className="px-4 py-3 text-right text-xs font-semibold text-slate-700 uppercase tracking-wider">
          Change
        </th>
      </tr>
    </thead>
    <tbody className="divide-y divide-slate-100">
      <tr className="hover:bg-slate-50 transition-colors">
        <td className="px-4 py-3 text-sm text-slate-900">Enterprise Plan</td>
        <td className="px-4 py-3 text-sm font-mono text-right text-slate-900">
          $284,392
        </td>
        <td className="px-4 py-3 text-sm font-mono text-right">
          <span className="text-emerald-600">+12.4%</span>
        </td>
      </tr>
      <tr className="hover:bg-slate-50 transition-colors">
        <td className="px-4 py-3 text-sm text-slate-900">Pro Plan</td>
        <td className="px-4 py-3 text-sm font-mono text-right text-slate-900">
          $158,203
        </td>
        <td className="px-4 py-3 text-sm font-mono text-right">
          <span className="text-red-600">-3.2%</span>
        </td>
      </tr>
    </tbody>
  </table>
</div>
```

**CSS Variables**:
```css
:root {
  --color-primary: #2563eb;  /* blue-600 */
  --color-success: #059669;  /* emerald-600 */
  --color-error: #dc2626;  /* red-600 */
  --spacing-default: 12px;
  --border-radius: 6px;
}
```

---

## 7. Editorial Elegance - Article Layout

**Characteristics**: Serif body text, generous line-height, 65-75 char measure

```tsx
<article className="max-w-[720px] mx-auto px-8 py-16">
  <header className="mb-12">
    <h1 className="font-serif text-6xl font-medium text-slate-900
                   leading-tight tracking-tight mb-4">
      The Art of Thoughtful Design
    </h1>
    <p className="text-lg text-slate-600">
      How intentional choices create meaningful experiences
    </p>
  </header>

  <div className="prose prose-lg prose-slate">
    <p className="font-serif text-xl leading-relaxed text-slate-800 mb-6">
      In the world of design, the difference between good and great often
      lies not in what we add, but in what we deliberately choose to omit.
      Every element must earn its place.
    </p>

    <p className="font-serif text-lg leading-loose text-slate-700 mb-6">
      This philosophy extends beyond mere aesthetics. It's about respecting
      the reader's time, attention, and cognitive capacity. When we strip
      away the unnecessary, what remains speaks with clarity and purpose.
    </p>
  </div>

  <figure className="my-12 -mx-8">
    <img src="/image.jpg" alt="" className="w-full" />
    <figcaption className="mt-3 text-sm text-slate-600 text-center">
      Minimalism as a design principle, captured in form
    </figcaption>
  </figure>
</article>
```

**CSS Variables**:
```css
:root {
  --font-serif: 'Tiempos Text', serif;
  --font-sans: 'Inter', sans-serif;
  --color-text: #0f172a;  /* slate-900 */
  --spacing-generous: 48px;
  --spacing-dramatic: 96px;
  --max-width: 720px;
  --line-height-body: 1.7;
}
```

---

## 8. Playful & Dynamic - Animated Button

**Characteristics**: Vibrant colors, bouncy animations, asymmetric layouts

```tsx
<button className="group relative px-8 py-4 rounded-2xl bg-gradient-to-r
                   from-purple-500 to-pink-500 text-white font-bold text-lg
                   shadow-[0_8px_32px_rgba(168,85,247,0.25)]
                   hover:shadow-[0_12px_48px_rgba(168,85,247,0.4)]
                   hover:scale-105 active:scale-95
                   transition-all duration-200 ease-out">
  <span className="relative z-10 flex items-center gap-2">
    Start Adventure
    <svg className="w-5 h-5 group-hover:translate-x-1 transition-transform"
         fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
            d="M13 7l5 5m0 0l-5 5m5-5H6" />
    </svg>
  </span>

  {/* Animated sparkles on hover */}
  <div className="absolute inset-0 rounded-2xl opacity-0 group-hover:opacity-100
                  transition-opacity duration-300">
    <div className="absolute top-2 right-4 w-2 h-2 bg-cyan-300 rounded-full
                    animate-ping" />
    <div className="absolute bottom-3 left-6 w-1 h-1 bg-yellow-300 rounded-full
                    animate-bounce" />
  </div>
</button>
```

**CSS Variables**:
```css
:root {
  --color-primary: #a855f7;  /* purple-500 */
  --color-secondary: #ec4899;  /* pink-500 */
  --color-accent: #06b6d4;  /* cyan-500 */
  --border-radius-dynamic: 16px;
  --animation-duration: 200ms;
  --animation-easing: cubic-bezier(0.68, -0.55, 0.265, 1.55);
}
```

---

## 9. Brutalist Minimalism - Raw Layout

**Characteristics**: Sharp corners, stark contrast, no shadows, harsh lines

```tsx
<div className="border-2 border-black bg-white p-4">
  <h2 className="font-sans text-4xl font-bold uppercase text-black mb-4
                 leading-none tracking-tight">
    PORTFOLIO
  </h2>

  <div className="space-y-0">
    <a href="#" className="block border-t-2 border-black p-4
                           hover:bg-black hover:text-white transition-colors
                           duration-0">
      <span className="font-sans text-xl font-bold uppercase">
        PROJECT 001 — IDENTITY SYSTEM
      </span>
    </a>

    <a href="#" className="block border-t-2 border-black p-4
                           hover:bg-black hover:text-white transition-colors
                           duration-0">
      <span className="font-sans text-xl font-bold uppercase">
        PROJECT 002 — WEB PLATFORM
      </span>
    </a>

    <a href="#" className="block border-y-2 border-black p-4
                           hover:bg-black hover:text-white transition-colors
                           duration-0">
      <span className="font-sans text-xl font-bold uppercase">
        PROJECT 003 — EXHIBITION DESIGN
      </span>
    </a>
  </div>
</div>
```

**CSS Variables**:
```css
:root {
  --color-primary: #000000;
  --color-secondary: #ffffff;
  --color-accent: #ff0000;  /* use sparingly */
  --border-width: 2px;
  --border-radius: 0px;
  --spacing-grid: 8px;
}
```

---

## 10. Neumorphic Softness - Toggle Switch

**Characteristics**: Soft shadows, rounded shapes, tactile feel, light backgrounds

```tsx
<div className="flex items-center justify-between p-6 rounded-3xl
                bg-gray-200 shadow-[8px_8px_16px_rgba(0,0,0,0.1),-8px_-8px_16px_rgba(255,255,255,0.8)]">
  <div>
    <h3 className="text-lg font-semibold text-slate-700 mb-1">
      Dark Mode
    </h3>
    <p className="text-sm text-slate-500">
      Switch to darker theme for comfortable reading
    </p>
  </div>

  <button
    className="relative w-16 h-8 rounded-full bg-gray-200
               shadow-[inset_4px_4px_8px_rgba(0,0,0,0.1),inset_-4px_-4px_8px_rgba(255,255,255,0.8)]
               transition-all duration-300 ease-in-out"
  >
    <div className="absolute top-1 left-1 w-6 h-6 rounded-full bg-gray-200
                    shadow-[4px_4px_8px_rgba(0,0,0,0.1),-4px_-4px_8px_rgba(255,255,255,0.8)]
                    transition-transform duration-300 ease-in-out" />
  </button>
</div>
```

**CSS Variables**:
```css
:root {
  --color-primary: #8b5cf6;  /* violet-500 */
  --bg-surface: #e5e7eb;  /* gray-200 */
  --border-radius: 24px;
  --shadow-raised: 8px 8px 16px rgba(0,0,0,0.1),
                   -8px -8px 16px rgba(255,255,255,0.8);
  --shadow-inset: inset 4px 4px 8px rgba(0,0,0,0.1),
                  inset -4px -4px 8px rgba(255,255,255,0.8);
}
```

---

## Comparison Table

| Direction | Font Primary | Color Primary | Spacing Base | Border Radius | Shadow Style |
|-----------|--------------|---------------|--------------|---------------|--------------|
| Precision & Density | IBM Plex Mono + Inter | Blue 600 | 4px | 4px | Minimal |
| Warmth & Approachability | Plus Jakarta Sans | Orange 500 | 8px | 12-16px | Soft |
| Sophistication & Trust | Tiempos + Inter | Blue 800 | 4px | 6-8px | Layered |
| Boldness & Clarity | Cabinet Grotesk | Violet 600 | 8px | 8-12px | Strong |
| Utility & Function | JetBrains Mono | Blue 500 | 4px | 4px/0px | None/Minimal |
| Data & Analysis | IBM Plex Sans/Mono | Blue 600 | 4px | 6px | Subtle |
| Editorial Elegance | Tiempos Text | Slate 900 | 8px | 0-4px | Minimal/None |
| Playful & Dynamic | Sora/Space Grotesk | Purple 500 | 4px | 8-20px | Bold+Colorful |
| Brutalist Minimalism | Helvetica/Grotesk | Black | 8px | 0px | None |
| Neumorphic Softness | Sofia Pro/Montserrat | Violet 500 | 4px | 16-24px | Soft Dual |

---

## Quick Decision Guide

**Need high information density?** → Precision & Density or Data & Analysis
**Need friendly consumer app?** → Warmth & Approachability
**Need professional credibility?** → Sophistication & Trust
**Need attention-grabbing landing page?** → Boldness & Clarity
**Need developer tool?** → Utility & Function
**Need content/publishing platform?** → Editorial Elegance
**Need gaming/creative tool?** → Playful & Dynamic
**Need artistic portfolio?** → Brutalist Minimalism
**Need wellness/lifestyle app?** → Neumorphic Softness

---

These examples demonstrate the distinct visual character of each direction. When implementing, always reference the full design system in `design-systems.md` for complete specifications.
