---
name: ui-portfolio-design
description: Dynamic UI design system generator that creates 10 unique, diverse design directions every time it's invoked. Generates both detailed specifications AND an interactive HTML preview so users can visually explore all options. Each direction includes complete specifications (typography, colors, spacing, components, motion). Automatically triggers before any new UI/frontend work to ensure intentional, cohesive aesthetics. Never shows the same 10 designs twice - always generates fresh options based on varied inspirations, aesthetics, and use cases.
---

# UI Portfolio Design - Dynamic Design Direction Generator

## Overview

This skill generates **10 completely new design directions** every time you start UI work. Each direction is a fully-specified design system with typography, colors, spacing, components, and motion guidelines.

**Important**: This skill generates NEW designs each time. Never reuse previous designs. Always create fresh, diverse options.

**Visual Preview**: After generating designs, create an interactive HTML preview at `/docs/ui-portfolio/index.html` so users can visually explore all 10 options.

## When This Skill Activates

Automatically triggers when:
- Starting ANY new UI/frontend work
- User requests design direction selection
- Building new components, pages, or interfaces
- User explicitly invokes `/ui-portfolio-design`

Skip only when:
- Modifying existing components (maintain current style)
- User explicitly provides a complete design specification
- Working in a codebase with established, documented design system

## Design Direction Generation Process

### Step 1: Generate 10 Diverse Design Directions

**CRITICAL**: Generate 10 COMPLETELY NEW design directions. Do NOT reuse these examples:
- ❌ Neon Cyberpunk, Organic Nature, Glass Morphism, Industrial Concrete, Luxury Gradient
- ❌ Retro Terminal, Pastel Dreamscape, High Contrast, Minimalist Japanese, Art Deco Glamour
- ❌ Bauhaus Precision, Alpine Capital, Deep Ocean Analytics, Paper Portfolio, Crystal Clarity
- ❌ Atomic Portfolio, Constructivist Data, Golden Hour Markets, Terminal Finance, Dyslexia-Friendly

**Generation Strategy**: Draw inspiration from varied sources to ensure diversity:

**Inspiration Sources** (pick from different categories):
1. **Art Movements**: Bauhaus, Constructivism, Memphis Design, Suprematism, De Stijl, Art Nouveau, Futurism, Cubism, Expressionism
2. **Architectural Styles**: Mid-century modern, Scandinavian, Industrial loft, Brutalism, Parametric, Deconstructivism
3. **Cultural Aesthetics**: Vaporwave, Cottagecore, Dark academia, Y2K, 90s grunge, 80s Miami, Steampunk, Solarpunk
4. **Natural Themes**: Desert landscapes, Ocean depths, Mountain peaks, Forest canopy, Arctic minimalism, Tropical vibrance
5. **Time Periods**: Belle Époque, Atomic Age, Victorian, Jazz Age, Postwar, Digital Age, Future noir
6. **Emotional Tones**: Energetic, Calm, Playful, Serious, Mysterious, Transparent, Bold, Subtle
7. **Material Inspirations**: Paper, Metal, Fabric, Wood, Stone, Glass, Neon, Holographic
8. **Industries**: Gaming, Healthcare, Education, Music, Sports, Travel, Food, Tech

**Diversity Requirements**:
- At least 8 of 10 must be from different inspiration sources
- Vary color palettes: monochrome, vibrant, muted, gradients, duotone
- Vary typography: serif, sans-serif, monospace, display, handwritten
- Vary spacing: compact, generous, asymmetric, grid-based
- Include at least 1 accessibility-focused option (high contrast, dyslexia-friendly)

### Step 2: Create Full Design System for Each Direction

For each of the 10 directions, generate complete specifications:

#### Typography System
```yaml
Primary Font: [Font name and fallbacks]
Secondary Font: [If needed]
Scale: [Size range, e.g., 12/14/16/20/24/32/40/48px]
Weights: [e.g., 300, 400, 600, 700]
Line Height: [e.g., 1.5, tight/normal/relaxed]
Letter Spacing: [e.g., 0.02em, tight/normal/wide]
Special: [Any unique characteristics, e.g., all-caps headers]
```

#### Color Palette
```yaml
Primary: [Hex codes and usage]
Secondary: [Hex codes and usage]
Accent: [Hex codes for highlights]
Success/Warning/Error: [State colors]
Neutrals: [Gray scale or equivalent]
Backgrounds: [Main bg colors]
Text: [Text colors for different backgrounds]
Special: [Any unique colors, gradients, or effects]
```

#### Spacing System
```yaml
Base Unit: [e.g., 4px, 8px]
Scale: [e.g., 4, 8, 12, 16, 24, 32, 48, 64]
Component Padding: [e.g., 16-24px, compact/generous]
Margins: [Approach to margins/gaps]
Layout: [Grid system, if any]
```

#### Component Styles
```yaml
Borders: [Width, style, color approach]
Border Radius: [Values, e.g., 0px sharp, 8px rounded, 16px+ pill]
Shadows: [Approach and values]
Textures: [Any background textures or patterns]
Inputs: [Form element styling]
Buttons: [Button styling approach]
Cards: [Card/container styling]
```

#### Motion & Interaction
```yaml
Duration: [Timing values, e.g., 150ms, 300ms, 500ms]
Easing: [Functions, e.g., ease-out, cubic-bezier]
Hover States: [How elements respond to hover]
Focus States: [Keyboard focus appearance]
Loading: [Loading state animations]
Transitions: [What properties animate]
```

#### Anti-Patterns (What NOT to do)
```yaml
- ❌ [Specific things that break this design direction]
- ❌ [At least 4-5 anti-patterns]
```

### Step 3: Generate Interactive HTML Preview

**CRITICAL**: After generating all 10 design directions, create an interactive HTML preview at `/docs/ui-portfolio/index.html` so the user can visually explore them.

**HTML Structure**:
```html
<!DOCTYPE html>
<html>
<head>
  <title>UI Portfolio - Design Directions</title>
  <script src="https://cdn.tailwindcss.com"></script>
  <link href="https://fonts.googleapis.com/css2?family=[fonts]&display=swap">
  <style>
    /* Custom styles for design-specific effects */
  </style>
</head>
<body>
  <!-- Selection Screen with all 10 options -->
  <div id="selection-screen">
    <!-- Grid of 10 cards, each clickable -->
  </div>

  <!-- 10 Full Dashboard Views (hidden by default) -->
  <div id="dashboard-1" class="hidden">
    <!-- Full dashboard example in that design system -->
  </div>
  <!-- ... repeat for all 10 -->

  <script>
    function showDashboard(id) {
      // Hide selection, show specific dashboard
    }
    function showSelection() {
      // Back to selection screen
    }
  </script>
</body>
</html>
```

**What to Include in Each Dashboard View**:
- Navigation bar
- 3-4 stat cards (showing example metrics)
- 1-2 data visualizations (charts, tables, or graphs)
- Example buttons, inputs, and other UI elements
- Typography hierarchy example
- All in the specific design direction's style

**After creating HTML**:
```typescript
// Write the file
Write({ file_path: "/Users/.../docs/ui-portfolio/index.html", content: html })

// Open it in browser
Bash({ command: "open /Users/.../docs/ui-portfolio/index.html" })
```

### Step 4: Present Portfolio to User

After generating and opening the HTML preview:

```
🎨 UI Portfolio: Choose Your Design Direction

I've generated 10 unique design directions for your [project type] and opened an interactive preview in your browser.

**Quick Overview**:

1. [Emoji] **[Direction Name]** - [One-line description]
2. [Emoji] **[Direction Name]** - [One-line description]
...
[All 10]

**For your [project type] project, I recommend:**
- Option [X]: [Direction Name] - [Why it fits]
- Option [Y]: [Direction Name] - [Alternative reasoning]

👉 **Explore them visually in your browser**, then choose a number (1-10) or describe what you're looking for.

Want to see 10 completely different options? Just ask!
```

### Step 5: Load Selected Design System

Once user selects a direction:

1. **Confirm Selection**
```
✓ Loaded: [Direction Name]

Key characteristics:
- Typography: [Summary]
- Colors: [Summary]
- Spacing: [Summary]
- Component style: [Summary]

Ready to implement. What would you like to build?
```

2. **Keep Full Spec in Memory**
   - Store the complete design system specifications
   - Reference them throughout implementation
   - Ensure consistency across all components

3. **Offer Token Generation** (if new project)
```
Would you like me to generate:
- Tailwind config with these design tokens
- CSS variables
- Design system documentation
```

## Implementation Guidelines

### Applying the Design System

**Typography**
```typescript
// Apply font family, size, weight, line-height, letter-spacing
<h1 className="font-[primary] text-[size] font-[weight] leading-[height] tracking-[spacing]">
```

**Colors**
```typescript
// Use palette consistently
<div className="bg-[background] text-[text]">
  <button className="bg-[primary] hover:bg-[primary-dark]">
```

**Spacing**
```typescript
// Follow spacing scale
<div className="p-[padding] space-y-[gap]">
```

**Components**
```typescript
// Match border radius, shadows, etc.
<Card className="rounded-[radius] shadow-[shadow] border-[border]">
```

**Motion**
```typescript
// Apply duration and easing
<button className="transition-[property] duration-[duration] ease-[easing]">
```

### Consistency Checks

Throughout implementation, verify:
- [ ] Typography matches specified fonts, sizes, weights
- [ ] Colors come from the defined palette
- [ ] Spacing uses the scale (not arbitrary values)
- [ ] Component styles match the direction (radius, shadows, borders)
- [ ] Animations use specified durations and easing
- [ ] Anti-patterns are avoided

## HTML Generation Guidelines

**For Financial Dashboards**:
- Show portfolio value, returns, holdings count
- Include chart or table
- Show transaction list or holdings breakdown
- Use realistic but fake data

**For E-commerce**:
- Product cards
- Shopping cart
- Checkout flow elements

**For Analytics**:
- Data tables
- Charts/graphs
- Filters and controls

**For Content Platforms**:
- Article layouts
- Typography hierarchy
- Reading experience

**Key Requirements**:
- Each dashboard must look distinctly different
- Use actual fonts from Google Fonts (or fallbacks)
- Include real color values from the palette
- Show motion/interaction patterns (hover states, etc.)
- Make it fully functional (navigation between views works)

## Generation Tips for Claude

**Creating Memorable Directions**:
- Give each direction a distinctive name (2-3 words + emoji)
- Make them visually distinct (user should easily tell them apart)
- Ensure each has a clear use case/target audience
- Make them complete enough to implement immediately

**Avoiding Repetition**:
- Track what you've generated before (mentally, in conversation)
- If user says "generate again", create 10 completely different options
- Draw from unexpected inspiration sources
- Mix and match elements in novel ways

**Quality Standards**:
- Each direction should be production-ready
- Colors should have sufficient contrast (check WCAG)
- Typography should be readable at stated sizes
- Spacing should feel intentional and systematic
- Motion should enhance, not distract

**Tailoring to Project**:
- Consider the project type (dashboard, marketing site, app, etc.)
- Consider the user's industry/domain
- Consider the target audience
- Suggest 2-3 directions that fit best, but present all 10

## Examples of Unique Directions

### Good: Novel and Specific
- 🏔️ **Alpine Precision** - Inspired by Swiss design + mountain landscapes
- 🌊 **Fluid Morphism** - Liquid animations + organic shapes
- 📺 **Broadcast Standard** - TV graphics + motion design patterns
- 🎭 **Theatrical Drama** - Stage lighting + bold typography
- 🧪 **Laboratory Clean** - Scientific instruments + clinical precision

### Avoid: Too Generic
- ❌ "Modern Minimal" - Too vague
- ❌ "Dark Mode" - Not a design system
- ❌ "Professional" - Doesn't specify aesthetic

## Multi-Session Support

**If user returns to the same project later**:
- Ask if they want to continue with previous design direction
- Or generate 10 new options
- Never assume they want the same designs

**If user wants to try alternatives**:
- Generate 10 completely new directions
- Can reference "similar to [previous choice] but more [quality]"

## Edge Cases

**User asks "show me more options"**:
- Generate 10 brand new directions
- Regenerate HTML with new designs
- Open in browser

**User says "I don't like any of these"**:
- Ask what they're looking for specifically
- Generate 10 new options based on their feedback
- Regenerate HTML

**User provides their own design spec**:
- Skip portfolio selection
- Use their specifications directly

**Working in existing codebase with design system**:
- Skip portfolio selection
- Analyze and follow existing patterns

---

**Remember**:
1. Every time this skill activates, generate 10 BRAND NEW design directions
2. Create an interactive HTML preview at `/docs/ui-portfolio/index.html`
3. Open the HTML in the browser so user can explore visually
4. Then let them choose which direction to use
