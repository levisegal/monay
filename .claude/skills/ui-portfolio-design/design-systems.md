# UI Portfolio Design Systems

This document defines 10 distinct design directions for the UI Portfolio Design skill. Each system provides a complete design language including typography, color, spacing, components, and motion guidelines.

---

## 1. Precision & Density

**When to Use**: Data-heavy dashboards, admin interfaces, analytics platforms, developer tools with high information density requirements.

**Character**: Efficient, functional, precise. Maximum information in minimum space while maintaining clarity. Every pixel serves a purpose.

### Typography
- **Primary Font**: IBM Plex Mono (monospace for data/numbers), Inter (UI text)
- **Scale**: Tight (12/14/16/18/20/24px)
- **Weights**: 400 (Regular), 500 (Medium), 600 (Semibold)
- **Line Height**: 1.4 (tighter for density)
- **Letter Spacing**: -0.01em (slightly condensed)

### Color Palette
- **Primary**: `#2563eb` (Blue 600) - actionable items
- **Secondary**: `#475569` (Slate 600) - secondary actions
- **Accent**: `#0ea5e9` (Sky 500) - highlights, active states
- **Success**: `#10b981` (Emerald 500)
- **Warning**: `#f59e0b` (Amber 500)
- **Error**: `#ef4444` (Red 500)
- **Neutrals**: `#0f172a` to `#f1f5f9` (Slate scale)
- **Background**: `#ffffff` (white), `#f8fafc` (slate-50) for panels

### Spacing System
- **Base Unit**: 4px
- **Scale**: 4, 8, 12, 16, 20, 24, 32, 40, 48px
- **Component Padding**: 8-12px (compact)
- **Section Margins**: 16-24px (efficient)
- **Grid Gap**: 12-16px

### Component Styles
- **Borders**: 1px solid, subtle (`#e2e8f0`)
- **Border Radius**: 4px (minimal, sharp corners)
- **Shadows**: Minimal, functional only (`0 1px 2px rgba(0,0,0,0.05)`)
- **Inputs**: 36px height, compact
- **Buttons**: 36px height, minimal padding (12px horizontal)
- **Tables**: Dense rows (32-36px), zebra striping, sticky headers
- **Cards**: Minimal padding (12-16px), subtle borders

### Motion & Interaction
- **Duration**: Fast (100-150ms)
- **Easing**: `ease-out` for functional transitions
- **Hover States**: Subtle background shifts, no dramatic effects
- **Focus States**: 2px outline, high contrast
- **Loading**: Minimal spinners, progress bars with precise percentages

### Anti-Patterns
- ❌ Excessive whitespace
- ❌ Large rounded corners
- ❌ Decorative shadows or gradients
- ❌ Slow, dramatic animations
- ❌ Mixed monospace and proportional fonts for data

---

## 2. Warmth & Approachability

**When to Use**: Consumer apps, social platforms, community tools, educational products, lifestyle applications.

**Character**: Friendly, inviting, human. Rounded edges, warm colors, generous spacing. Makes users feel comfortable and welcome.

### Typography
- **Primary Font**: Plus Jakarta Sans, Satoshi, or DM Sans
- **Scale**: Generous (14/16/18/20/24/32/40px)
- **Weights**: 400 (Regular), 500 (Medium), 700 (Bold)
- **Line Height**: 1.6 (relaxed, readable)
- **Letter Spacing**: 0 (natural)

### Color Palette
- **Primary**: `#f97316` (Orange 500) - warm, friendly
- **Secondary**: `#ec4899` (Pink 500) - playful accents
- **Accent**: `#14b8a6` (Teal 500) - fresh highlights
- **Success**: `#22c55e` (Green 500)
- **Warning**: `#fbbf24` (Amber 400)
- **Error**: `#f87171` (Red 400) - softer than typical error red
- **Neutrals**: `#1e293b` to `#f8fafc` (Slate scale, warm-leaning)
- **Background**: `#fefce8` (yellow-50), `#fff7ed` (orange-50) for warmth

### Spacing System
- **Base Unit**: 4px
- **Scale**: 8, 12, 16, 20, 24, 32, 40, 48, 64px
- **Component Padding**: 16-20px (generous)
- **Section Margins**: 32-48px (breathing room)
- **Grid Gap**: 20-24px

### Component Styles
- **Borders**: 2px solid, friendly colors or none
- **Border Radius**: 12-16px (noticeably rounded)
- **Shadows**: Soft, ambient (`0 4px 12px rgba(0,0,0,0.08)`)
- **Inputs**: 44px height, rounded, friendly
- **Buttons**: 44px height, fully rounded (`9999px`), generous padding (20-24px)
- **Cards**: Rounded (12-16px), soft shadows, comfortable padding (20-24px)
- **Avatars**: Circular, colorful borders

### Motion & Interaction
- **Duration**: Medium (200-300ms)
- **Easing**: `ease-in-out` for smooth, friendly motion
- **Hover States**: Gentle scale (1.02), brightness shift
- **Focus States**: Thick (3px), rounded, colorful outline
- **Loading**: Bouncing dots, friendly spinners with personality
- **Micro-interactions**: Celebrate success with confetti, gentle bounce on click

### Anti-Patterns
- ❌ Sharp corners everywhere
- ❌ Cold, corporate blues/grays only
- ❌ Minimal, sparse layouts
- ❌ Stiff, mechanical animations
- ❌ Harsh, dramatic shadows

---

## 3. Sophistication & Trust

**When to Use**: Financial services, enterprise SaaS, legal tech, healthcare, professional services, high-stakes decision-making tools.

**Character**: Refined, trustworthy, professional. Establishes credibility through restrained elegance and consistent attention to detail.

### Typography
- **Primary Font**: Söhne, Tiempos Text, or Canela Text (serif headings) + Inter/Untitled Sans (UI)
- **Scale**: Classic (14/16/18/20/24/32/40/48px)
- **Weights**: 400 (Regular), 500 (Medium), 600 (Semibold)
- **Line Height**: 1.5 (balanced)
- **Letter Spacing**: -0.02em for headings (refined)

### Color Palette
- **Primary**: `#1e40af` (Blue 800) - trustworthy, corporate
- **Secondary**: `#334155` (Slate 700) - professional gray
- **Accent**: `#0891b2` (Cyan 600) - subtle highlights
- **Success**: `#059669` (Emerald 600) - conservative
- **Warning**: `#d97706` (Amber 600)
- **Error**: `#dc2626` (Red 600)
- **Neutrals**: `#020617` to `#f8fafc` (Slate scale, deep blacks)
- **Background**: `#ffffff` (white), `#fafafa` (neutral-50) for subtle panels

### Spacing System
- **Base Unit**: 4px
- **Scale**: 4, 8, 12, 16, 20, 24, 32, 40, 48, 64px
- **Component Padding**: 16-20px (balanced)
- **Section Margins**: 32-48px (generous but controlled)
- **Grid Gap**: 20-24px

### Component Styles
- **Borders**: 1px solid, subtle and refined (`#e5e7eb`)
- **Border Radius**: 6-8px (slightly rounded, not playful)
- **Shadows**: Layered, refined (`0 1px 3px rgba(0,0,0,0.06), 0 4px 12px rgba(0,0,0,0.04)`)
- **Inputs**: 40px height, clean borders
- **Buttons**: 40px height, medium padding (16-20px), subtle hover states
- **Cards**: Subtle elevation, consistent padding (20-24px), clean borders
- **Data Tables**: Professional, clear hierarchy, subtle row dividers

### Motion & Interaction
- **Duration**: Measured (150-250ms)
- **Easing**: `cubic-bezier(0.4, 0, 0.2, 1)` (smooth, refined)
- **Hover States**: Subtle background tint, no dramatic shifts
- **Focus States**: 2px outline, professional blue
- **Loading**: Elegant progress indicators, no flashy spinners
- **Transitions**: Smooth, never jarring or attention-seeking

### Anti-Patterns
- ❌ Bright, playful colors
- ❌ Overly rounded or pill-shaped elements
- ❌ Bouncy, playful animations
- ❌ Decorative illustrations in critical interfaces
- ❌ Inconsistent spacing or alignment

---

## 4. Boldness & Clarity

**When to Use**: Marketing sites, e-commerce, landing pages, brand-forward products, content that needs to grab attention.

**Character**: Confident, direct, impactful. Large typography, strong color contrasts, clear hierarchy. Every element has intentional visual weight.

### Typography
- **Primary Font**: Clash Display, Cabinet Grotesk, or Bricolage Grotesque (display) + Inter/Untitled Sans (body)
- **Scale**: Bold (16/18/20/24/32/40/56/72px)
- **Weights**: 500 (Medium), 600 (Semibold), 700 (Bold), 800 (Extrabold)
- **Line Height**: 1.2 for headings, 1.6 for body (dramatic contrast)
- **Letter Spacing**: -0.03em for large headings (tight, impactful)

### Color Palette
- **Primary**: `#7c3aed` (Violet 600) - bold, memorable
- **Secondary**: `#0f172a` (Slate 900) - strong contrast
- **Accent**: `#f59e0b` (Amber 500) - high-energy highlights
- **Success**: `#10b981` (Emerald 500)
- **Warning**: `#f59e0b` (Amber 500)
- **Error**: `#ef4444` (Red 500)
- **Neutrals**: `#0f172a` to `#ffffff` (high contrast scale)
- **Background**: Bold gradients (`violet-600` to `fuchsia-500`), solid `#ffffff`

### Spacing System
- **Base Unit**: 8px (larger for impact)
- **Scale**: 8, 16, 24, 32, 40, 48, 64, 80, 96px
- **Component Padding**: 20-32px (generous, confident)
- **Section Margins**: 48-80px (dramatic spacing)
- **Grid Gap**: 24-32px

### Component Styles
- **Borders**: 2-3px solid, bold colors
- **Border Radius**: 8-12px (balanced, not timid)
- **Shadows**: Strong, intentional (`0 8px 24px rgba(0,0,0,0.12), 0 4px 8px rgba(0,0,0,0.08)`)
- **Inputs**: 48px height, bold focus states
- **Buttons**: 48-56px height, large padding (24-32px), bold hover effects
- **Cards**: Strong elevation, bold borders or shadows
- **Hero Sections**: Full-bleed backgrounds, oversized typography

### Motion & Interaction
- **Duration**: Punchy (150-200ms)
- **Easing**: `ease-out` for snappy feel
- **Hover States**: Bold scale (1.05), dramatic color shifts
- **Focus States**: 3px outline, high-contrast colors
- **Loading**: Bold progress bars, strong visual feedback
- **Scroll Animations**: Fade-in, slide-up reveals for impact
- **CTA Animations**: Pulse, glow effects on key actions

### Anti-Patterns
- ❌ Timid, washed-out colors
- ❌ Small, apologetic typography
- ❌ Excessive text without visual hierarchy
- ❌ Minimal, sparse designs
- ❌ Subtle, forgettable interactions

---

## 5. Utility & Function

**When to Use**: Developer tools, CLI interfaces, technical documentation, code editors, system utilities, power-user products.

**Character**: No-nonsense, efficient, keyboard-first. Optimized for function over form. Clean, technical aesthetic that gets out of the way.

### Typography
- **Primary Font**: JetBrains Mono, Berkeley Mono, or Commit Mono (monospace throughout)
- **Scale**: Terminal-inspired (12/13/14/16/18/20px)
- **Weights**: 400 (Regular), 500 (Medium), 700 (Bold)
- **Line Height**: 1.5 (code-friendly)
- **Letter Spacing**: 0 (monospace natural)

### Color Palette
- **Primary**: `#3b82f6` (Blue 500) - functional blue
- **Secondary**: `#6b7280` (Gray 500) - neutral
- **Accent**: `#8b5cf6` (Violet 500) - syntax highlighting inspired
- **Success**: `#10b981` (Emerald 500)
- **Warning**: `#f59e0b` (Amber 500)
- **Error**: `#ef4444` (Red 500)
- **Neutrals**: `#111827` to `#f9fafb` (Gray scale)
- **Background**: `#1e1e1e` (dark mode primary), `#ffffff` (light mode)
- **Code Blocks**: Based on VS Code Dark+ theme

### Spacing System
- **Base Unit**: 4px
- **Scale**: 4, 8, 12, 16, 20, 24, 32px (no large values)
- **Component Padding**: 8-12px (compact, terminal-like)
- **Section Margins**: 16-24px (efficient)
- **Grid Gap**: 12-16px

### Component Styles
- **Borders**: 1px solid, functional (`#374151`)
- **Border Radius**: 4px (minimal) or 0px (terminal mode)
- **Shadows**: Minimal or none
- **Inputs**: 32-36px height, monospace font, terminal-inspired
- **Buttons**: 32-36px height, compact, clear states
- **Code Blocks**: Dark background, syntax highlighting, line numbers
- **Tables**: Monospace, alternating rows, dense

### Motion & Interaction
- **Duration**: Instant (50-100ms) or none
- **Easing**: `linear` (no easing curves)
- **Hover States**: Background color only, no transforms
- **Focus States**: 2px solid outline, keyboard-friendly
- **Loading**: Simple spinners, progress text (e.g., "45%")
- **Minimal Motion**: Respect prefers-reduced-motion by default

### Anti-Patterns
- ❌ Decorative elements or illustrations
- ❌ Rounded, soft shapes
- ❌ Slow, smooth animations
- ❌ Mixed fonts (stay monospace)
- ❌ Bright, playful colors

---

## 6. Data & Analysis

**When to Use**: Analytics platforms, BI tools, data visualization dashboards, scientific applications, research tools.

**Character**: Clarity through data. Every element supports understanding complex information. Charts and tables are first-class citizens.

### Typography
- **Primary Font**: IBM Plex Sans (UI) + IBM Plex Mono (data/numbers)
- **Scale**: Data-friendly (12/14/16/18/20/24/32px)
- **Weights**: 400 (Regular), 500 (Medium), 600 (Semibold)
- **Line Height**: 1.4 (tight for density)
- **Letter Spacing**: 0 for body, -0.01em for headings

### Color Palette
- **Primary**: `#2563eb` (Blue 600) - data blue
- **Secondary**: `#64748b` (Slate 500) - neutral
- **Accent**: `#06b6d4` (Cyan 500) - highlight data points
- **Data Palette**:
  - Categorical: `#3b82f6`, `#8b5cf6`, `#ec4899`, `#f59e0b`, `#10b981`, `#06b6d4`
  - Sequential: Blue scale (`#eff6ff` to `#1e3a8a`)
  - Diverging: Red to Blue (`#dc2626` → `#f3f4f6` → `#2563eb`)
- **Success**: `#059669` (Emerald 600)
- **Warning**: `#d97706` (Amber 600)
- **Error**: `#dc2626` (Red 600)
- **Neutrals**: `#0f172a` to `#f8fafc` (Slate scale)
- **Background**: `#ffffff`, `#f8fafc` for panels

### Spacing System
- **Base Unit**: 4px
- **Scale**: 4, 8, 12, 16, 20, 24, 32, 40px
- **Component Padding**: 12-16px (balanced)
- **Section Margins**: 20-32px (organized)
- **Grid Gap**: 16-20px

### Component Styles
- **Borders**: 1px solid, subtle (`#e2e8f0`)
- **Border Radius**: 6px (slightly rounded)
- **Shadows**: Subtle (`0 1px 3px rgba(0,0,0,0.06)`)
- **Inputs**: 36px height, clear states
- **Buttons**: 36px height, functional
- **Charts**: Clean axes, subtle gridlines, colorblind-safe palettes
- **Tables**: Sortable headers, right-aligned numbers, heatmap cells
- **Cards**: Stat cards with large numbers, sparklines, trend indicators

### Motion & Interaction
- **Duration**: Fast (100-150ms)
- **Easing**: `ease-out`
- **Hover States**: Highlight data points, show tooltips
- **Focus States**: 2px outline
- **Loading**: Skeleton screens for charts, progressive data loading
- **Chart Animations**: Fade-in on load, smooth transitions on filter changes
- **Interactive Tooltips**: Instant show on hover, detailed data context

### Anti-Patterns
- ❌ Decorative charts (pie charts with 3D effects)
- ❌ Rainbow color schemes (not accessible)
- ❌ Cluttered dashboards without hierarchy
- ❌ Slow chart rendering or interactions
- ❌ Mixing monospace and proportional for numbers

---

## 7. Editorial Elegance

**When to Use**: Content platforms, publishing sites, blogs, magazines, portfolios, long-form reading experiences.

**Character**: Typographic refinement, generous whitespace, classic proportions. Optimized for readability and visual pleasure. Content is the star.

### Typography
- **Primary Font**: Tiempos Text, Canela Text, or Playfair Display (serif) + Untitled Sans or Inter (UI)
- **Scale**: Editorial (16/18/20/24/32/40/56/72px)
- **Weights**: 400 (Regular), 500 (Medium), 600 (Semibold) for serif; 400/600 for sans
- **Line Height**: 1.7 for body text (generous), 1.2 for headings
- **Letter Spacing**: 0 for body, -0.02em for large headings
- **Measure**: 65-75 characters per line (optimal readability)

### Color Palette
- **Primary**: `#0f172a` (Slate 900) - near-black for text
- **Secondary**: `#475569` (Slate 600) - muted text
- **Accent**: `#dc2626` (Red 600) - editorial accent (links, highlights)
- **Success**: `#059669` (Emerald 600)
- **Warning**: `#d97706` (Amber 600)
- **Error**: `#dc2626` (Red 600)
- **Neutrals**: Rich blacks and warm whites (`#1e293b` to `#fefefe`)
- **Background**: `#fefefe` (warm white), `#f9fafb` (subtle cream) for panels

### Spacing System
- **Base Unit**: 4px
- **Scale**: 8, 16, 24, 32, 40, 48, 64, 80, 96, 128px (generous)
- **Component Padding**: 24-32px (spacious)
- **Section Margins**: 64-96px (editorial breathing room)
- **Grid Gap**: 32-48px
- **Content Width**: 680-720px max (classic measure)

### Component Styles
- **Borders**: 1px solid, minimal and refined (`#e5e7eb`)
- **Border Radius**: 0-4px (minimal, classical)
- **Shadows**: Extremely subtle or none
- **Inputs**: 44px height, clean and simple
- **Buttons**: Text buttons or minimal filled, subtle hover
- **Pull Quotes**: Large serif, generous margins, vertical rules
- **Drop Caps**: Large initial letters, classical typography
- **Figures**: Full-bleed images, elegant captions in smaller sans-serif

### Motion & Interaction
- **Duration**: Gentle (200-300ms)
- **Easing**: `ease-in-out` (smooth, never jarring)
- **Hover States**: Subtle underline on links, gentle color shifts
- **Focus States**: Minimal, unobtrusive outline
- **Loading**: Elegant fade-ins, no harsh loading states
- **Scroll**: Smooth parallax on hero images, subtle fade-ins for content
- **Reading Progress**: Thin progress bar at top

### Anti-Patterns
- ❌ Sans-serif body text (use for UI only)
- ❌ Tight line height (<1.6 for body)
- ❌ Narrow measure (<60 characters) or wide (>80 characters)
- ❌ Busy, cluttered layouts
- ❌ Harsh, mechanical animations

---

## 8. Playful & Dynamic

**When to Use**: Gaming platforms, creative tools, entertainment apps, youth-focused products, experimental interfaces.

**Character**: Energetic, fun, unexpected. Not afraid to break conventions. Celebrates personality and brand identity through bold design choices.

### Typography
- **Primary Font**: Sora, Space Grotesk, or Poppins (geometric, modern)
- **Scale**: Dynamic (14/16/18/24/32/40/56/80px)
- **Weights**: 400 (Regular), 600 (Semibold), 700 (Bold), 800 (Extrabold)
- **Line Height**: Variable (1.2 for headings, 1.6 for body)
- **Letter Spacing**: 0.02em (slightly open, playful)

### Color Palette
- **Primary**: `#a855f7` (Purple 500) - vibrant, playful
- **Secondary**: `#ec4899` (Pink 500) - energetic
- **Accent**: `#06b6d4` (Cyan 500) - fresh contrast
- **Success**: `#22c55e` (Green 500)
- **Warning**: `#fbbf24` (Amber 400)
- **Error**: `#f472b6` (Pink 400) - playful error state
- **Neutrals**: `#18181b` to `#fafafa` (Zinc scale)
- **Background**: Gradient backgrounds, animated meshes, vibrant colors

### Spacing System
- **Base Unit**: 4px
- **Scale**: 4, 8, 12, 16, 20, 24, 32, 40, 48, 64px
- **Component Padding**: 16-24px
- **Section Margins**: 32-48px
- **Grid Gap**: 20-32px (asymmetric layouts encouraged)

### Component Styles
- **Borders**: 2-3px solid, bold colors, or none
- **Border Radius**: Varies by component (8-20px, mixed for personality)
- **Shadows**: Bold, colorful (`0 8px 32px rgba(168, 85, 247, 0.25)`)
- **Inputs**: 44px height, fun focus states with color
- **Buttons**: 44-48px height, oversized, fun hover effects (scale, tilt)
- **Cards**: Tilt on hover, colorful borders, playful shadows
- **Illustrations**: Primary design element, custom characters

### Motion & Interaction
- **Duration**: Varied (100ms for micro, 300ms for dramatic)
- **Easing**: `cubic-bezier(0.68, -0.55, 0.265, 1.55)` (bounce)
- **Hover States**: Scale, rotate, tilt transforms
- **Focus States**: 3px outline, colorful and animated
- **Loading**: Playful spinners, animated characters, fun messages
- **Confetti**: Celebrate actions with particle effects
- **Sound**: Consider UI sound effects for key interactions
- **Cursor Effects**: Custom cursors, trail effects

### Anti-Patterns
- ❌ Corporate, conservative color schemes
- ❌ Rigid, uniform layouts
- ❌ Stiff, mechanical animations
- ❌ Serious, formal tone
- ❌ Predictable patterns (embrace asymmetry)

---

## 9. Brutalist Minimalism

**When to Use**: Art/design portfolios, creative agencies, experimental products, brand-forward experiences, counter-culture apps.

**Character**: Raw, unpolished, honest. Strips away decoration to expose underlying structure. Typography and layout as primary expression.

### Typography
- **Primary Font**: Neue Haas Grotesk, Univers, or Helvetica (neo-grotesque) / Space Mono (monospace variant)
- **Scale**: Stark (14/16/20/24/32/48/64/96px)
- **Weights**: 400 (Regular), 700 (Bold) only (no in-between)
- **Line Height**: 1.2-1.3 (tight, compressed)
- **Letter Spacing**: 0 (natural, no adjustment)
- **Text Transform**: Uppercase for emphasis

### Color Palette
- **Primary**: `#000000` (Pure black)
- **Secondary**: `#ffffff` (Pure white)
- **Accent**: `#ff0000` (Pure red) or `#0000ff` (Pure blue) - single accent only
- **Neutrals**: Sharp contrast, no gradients (`#000000`, `#333333`, `#cccccc`, `#ffffff`)
- **Background**: `#ffffff` or `#000000` (stark contrast)

### Spacing System
- **Base Unit**: 8px (deliberate, visible grid)
- **Scale**: 8, 16, 24, 32, 40, 48, 64, 80px
- **Component Padding**: Minimal (8-16px)
- **Section Margins**: Dramatic or none (intentional extremes)
- **Grid Gap**: Visible, structural (24-32px)

### Component Styles
- **Borders**: 1-2px solid black, harsh lines
- **Border Radius**: 0px (no rounding, sharp corners everywhere)
- **Shadows**: None (flat design, no depth)
- **Inputs**: 40px height, harsh borders, no fancy styling
- **Buttons**: Rectangle, flat, harsh hover (invert colors)
- **Cards**: Flat, bordered boxes, no elevation
- **Typography**: Bold, oversized, overlapping text as design element
- **Raw HTML**: Default styles acceptable, <marquee> ironically embraced

### Motion & Interaction
- **Duration**: None or instant (0ms or 50ms)
- **Easing**: `linear` only (no curves)
- **Hover States**: Harsh color inversion, no gradual transitions
- **Focus States**: Harsh outline, high contrast
- **Loading**: Simple text ("LOADING..."), no spinners
- **Scroll**: None or jarring snap scrolling
- **Cursor**: Large, custom, intrusive

### Anti-Patterns
- ❌ Gradients, soft shadows, blur effects
- ❌ Rounded corners
- ❌ Smooth, eased animations
- ❌ Decorative elements
- ❌ "Nice" design patterns

---

## 10. Neumorphic Softness

**When to Use**: Health/wellness apps, mindfulness tools, lifestyle applications, meditation apps, soft-skill learning platforms.

**Character**: Soft, tactile, calm. Creates sense of physical depth through subtle lighting. Invites touch and interaction while maintaining serenity.

### Typography
- **Primary Font**: Sofia Pro, Circular, or Montserrat (soft, rounded sans)
- **Scale**: Gentle (14/16/18/20/24/32/40px)
- **Weights**: 400 (Regular), 500 (Medium), 600 (Semibold)
- **Line Height**: 1.6 (relaxed)
- **Letter Spacing**: 0.01em (slightly open, breathing)

### Color Palette
- **Primary**: `#8b5cf6` (Violet 500) - soft purple
- **Secondary**: `#a78bfa` (Violet 400) - lighter variant
- **Accent**: `#c084fc` (Purple 400) - gentle highlight
- **Success**: `#34d399` (Emerald 400) - soft green
- **Warning**: `#fbbf24` (Amber 400)
- **Error**: `#fca5a5` (Red 300) - soft, non-alarming
- **Neutrals**: `#e5e7eb` to `#f9fafb` (Gray scale, light)
- **Background**: `#e5e7eb` (gray-200) or `#ede9fe` (violet-100) - soft, textured

### Spacing System
- **Base Unit**: 4px
- **Scale**: 8, 12, 16, 20, 24, 32, 40, 48px
- **Component Padding**: 16-24px (comfortable)
- **Section Margins**: 32-48px (spacious)
- **Grid Gap**: 20-32px

### Component Styles
- **Borders**: None (neumorphic shadows create edges)
- **Border Radius**: 16-24px (very rounded, soft)
- **Shadows**:
  - Raised: `8px 8px 16px rgba(0,0,0,0.1), -8px -8px 16px rgba(255,255,255,0.8)`
  - Inset: `inset 4px 4px 8px rgba(0,0,0,0.1), inset -4px -4px 8px rgba(255,255,255,0.8)`
- **Inputs**: 44px height, soft inset shadows
- **Buttons**: 44px height, soft raised shadows, press effect (inset on click)
- **Cards**: Soft raised effect, generous padding (24-32px)
- **Toggles**: Neumorphic switches, tactile feel

### Motion & Interaction
- **Duration**: Smooth (200-300ms)
- **Easing**: `ease-in-out` (gentle, flowing)
- **Hover States**: Subtle shadow increase, soft glow
- **Focus States**: Soft, glowing outline
- **Loading**: Pulsing, breathing animations
- **Press Effect**: Inset shadow on active state (feels tactile)
- **Transitions**: Smooth depth changes, floating elements

### Anti-Patterns
- ❌ Sharp corners or harsh lines
- ❌ High contrast colors
- ❌ Flat design (needs depth)
- ❌ Fast, snappy animations
- ❌ Dark mode (neumorphism needs light backgrounds)

---

## Portfolio Selection Flow

When the skill is invoked, present these 10 options in an interactive format:

### Presentation Format

```
Which design direction fits your project?

1. 🎯 Precision & Density
   Best for: Data dashboards, admin interfaces, analytics
   Style: Efficient, functional, high information density

2. 🌟 Warmth & Approachability
   Best for: Consumer apps, social platforms, communities
   Style: Friendly, inviting, generous spacing

3. 💼 Sophistication & Trust
   Best for: Financial services, enterprise SaaS, professional tools
   Style: Refined, trustworthy, attention to detail

4. ⚡ Boldness & Clarity
   Best for: Marketing sites, e-commerce, landing pages
   Style: Confident, impactful, strong hierarchy

5. 🛠️ Utility & Function
   Best for: Developer tools, technical products, power-user apps
   Style: No-nonsense, keyboard-first, efficient

6. 📊 Data & Analysis
   Best for: Analytics platforms, BI tools, research applications
   Style: Clarity through data, visualization-first

7. 📖 Editorial Elegance
   Best for: Content platforms, publishing, long-form reading
   Style: Typographic refinement, generous whitespace

8. 🎮 Playful & Dynamic
   Best for: Gaming, creative tools, entertainment apps
   Style: Energetic, fun, unexpected interactions

9. 🔲 Brutalist Minimalism
   Best for: Art portfolios, agencies, experimental products
   Style: Raw, unpolished, honest structure

10. 🌸 Neumorphic Softness
    Best for: Health/wellness, mindfulness, lifestyle apps
    Style: Soft, tactile, calming depth

Enter number (1-10) or describe your needs:
```

### After Selection

Once the user selects a direction:
1. Load the complete design system for that direction
2. Generate Tailwind config or CSS variables
3. Provide component implementation guidelines
4. Reference anti-patterns to avoid
5. Begin implementation with chosen system

---

## Design Token Generation

After selection, generate framework-specific tokens:

### Tailwind Config Example (Precision & Density)

```javascript
module.exports = {
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['IBM Plex Mono', 'monospace'],
      },
      colors: {
        primary: '#2563eb',
        secondary: '#475569',
        accent: '#0ea5e9',
      },
      spacing: {
        '4': '4px',
        '8': '8px',
        '12': '12px',
        '16': '16px',
        '20': '20px',
        '24': '24px',
      },
      borderRadius: {
        DEFAULT: '4px',
      },
    },
  },
}
```

### CSS Variables Example

```css
:root {
  /* Colors */
  --color-primary: #2563eb;
  --color-secondary: #475569;
  --color-accent: #0ea5e9;

  /* Typography */
  --font-sans: 'Inter', system-ui, sans-serif;
  --font-mono: 'IBM Plex Mono', monospace;

  /* Spacing */
  --spacing-unit: 4px;
  --spacing-compact: 8px;
  --spacing-default: 16px;
  --spacing-loose: 24px;

  /* Border */
  --border-radius: 4px;
  --border-width: 1px;
}
```

---

## Usage Guidelines

### For Claude

When implementing UI components:

1. **Ask for Selection**: Always present the portfolio before starting implementation
2. **Load System**: Once selected, load complete design system
3. **Apply Consistently**: Use typography, colors, spacing, and components from chosen system
4. **Reference Anti-Patterns**: Check anti-patterns list and avoid them
5. **Generate Tokens**: Output configuration files for the project's framework
6. **Maintain Coherence**: All components should feel like they belong to the same system

### For Users

- Select based on your project's **purpose** and **audience**
- Consider the **tone** you want to communicate
- Think about **use cases**: data-heavy vs content-focused vs action-oriented
- You can **mix systems** for different sections (e.g., Editorial for blog, Precision for dashboard)
- When in doubt, start with **Sophistication & Trust** or **Warmth & Approachability** as balanced defaults
