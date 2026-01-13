# UI Portfolio Design - Claude Code Skill

A comprehensive UI design skill for Claude Code that presents a curated portfolio of 10 professional design systems before starting any frontend work.

## What This Skill Does

Instead of generating generic "AI slop" UI, this skill:

1. **Presents 10 Design Options** - Shows a portfolio of distinct visual directions optimized for different use cases
2. **Loads Complete Design Systems** - Each direction includes typography, colors, spacing, components, and motion guidelines
3. **Generates Framework Tokens** - Outputs Tailwind configs or CSS variables for your project
4. **Maintains Consistency** - Ensures all components follow the chosen system's patterns and anti-patterns

## Installation

### For This Repository (monay)

The skill is already in `.claude/skills/ui-portfolio-design/`. Claude Code will automatically detect and load it.

### For Other Projects

1. Copy this directory to `~/.claude/skills/ui-portfolio-design/`
2. Restart Claude Code
3. The skill will auto-trigger on UI/frontend work

## Design Directions

### 1. 🎯 Precision & Density
**For**: Data dashboards, admin interfaces, analytics platforms
**Style**: Efficient, functional, high information density
**Key Traits**: Monospace for data, tight spacing, minimal shadows

### 2. 🌟 Warmth & Approachability
**For**: Consumer apps, social platforms, communities
**Style**: Friendly, inviting, generous spacing
**Key Traits**: Rounded elements, warm colors, soft shadows

### 3. 💼 Sophistication & Trust
**For**: Financial services, enterprise SaaS, professional tools
**Style**: Refined, trustworthy, restrained elegance
**Key Traits**: Serif headings, deep blues, layered shadows

### 4. ⚡ Boldness & Clarity
**For**: Marketing sites, e-commerce, landing pages
**Style**: Confident, impactful, strong hierarchy
**Key Traits**: Large typography, bold colors, dramatic spacing

### 5. 🛠️ Utility & Function
**For**: Developer tools, technical products, power-user apps
**Style**: No-nonsense, keyboard-first, terminal-inspired
**Key Traits**: Monospace throughout, minimal motion, dark mode

### 6. 📊 Data & Analysis
**For**: Analytics platforms, BI tools, research applications
**Style**: Clarity through data, visualization-first
**Key Traits**: Colorblind-safe palettes, sortable tables, interactive charts

### 7. 📖 Editorial Elegance
**For**: Content platforms, publishing, blogs, magazines
**Style**: Typographic refinement, generous whitespace
**Key Traits**: Serif body text, 65-75 char measure, subtle accents

### 8. 🎮 Playful & Dynamic
**For**: Gaming, creative tools, entertainment apps
**Style**: Energetic, fun, unexpected interactions
**Key Traits**: Bold animations, vibrant colors, asymmetric layouts

### 9. 🔲 Brutalist Minimalism
**For**: Art portfolios, creative agencies, experimental products
**Style**: Raw, unpolished, honest structure
**Key Traits**: Sharp corners, stark contrast, no shadows

### 10. 🌸 Neumorphic Softness
**For**: Health/wellness, mindfulness, lifestyle apps
**Style**: Soft, tactile, calming depth
**Key Traits**: Subtle shadows, rounded shapes, light backgrounds

## Usage

### Automatic Activation

The skill automatically triggers when you start any UI/frontend work:

```
You: "Build a dashboard component"

Claude: "🎨 UI Portfolio: Choose Your Design Direction

Select the style that best fits your project:

1. 🎯 Precision & Density
   → Data dashboards • Admin interfaces • Analytics platforms
   ...

Choose a number (1-10), or describe your project and I'll recommend options."
```

### Manual Activation

You can also explicitly invoke it:

```
You: "/ui-portfolio-design"
```

### Example Workflow

```
You: "Create a user profile page"

Claude: [Presents 10 options]

You: "Option 2 - Warmth & Approachability"

Claude: "✓ Loaded: Warmth & Approachability

Key characteristics:
- Typography: Plus Jakarta Sans, generous scale (14-40px)
- Primary colors: Orange 500, Pink 500, Teal 500
- Spacing: 8px base, generous (16-20px padding)
- Component style: 12-16px rounded, soft shadows

Ready to implement. What would you like to build?"

You: "Start with a profile card"

Claude: [Implements with rounded corners, warm colors, generous spacing]
```

## File Structure

```
ui-portfolio-design/
├── SKILL.md               # Main skill definition with workflow
├── design-systems.md      # Complete specs for all 10 directions
├── README.md              # This file
└── examples/              # Visual previews (optional)
    ├── precision.png
    ├── warmth.png
    └── ...
```

## Customization

### Using Custom Directions

If none of the 10 directions fit perfectly:

1. Choose the closest match
2. Specify deviations:
   ```
   "I want Precision & Density, but with warmer colors and slightly more spacing"
   ```
3. Claude will generate modified tokens while maintaining coherence

### Multi-Section Applications

You can use different directions for different sections:

**Example: SaaS Product**
- Marketing site → Boldness & Clarity
- Product docs → Editorial Elegance
- App dashboard → Precision & Density

**Implementation**:
1. Choose primary direction for main app
2. Select secondary for distinct sections
3. Maintain clear boundaries (routes/layouts)
4. Share brand colors across sections

## Framework Support

### React + Tailwind + shadcn/ui
✅ Full support with Tailwind config generation

### Next.js
✅ Global styles and typography setup

### Vue / Svelte
✅ CSS variables and design tokens

### Other Frameworks
✅ Adaptable to any CSS-based framework

## Anti-Patterns Avoided

This skill actively avoids common "AI slop" patterns:

- ❌ Generic Inter/Roboto everywhere
- ❌ Timid, washed-out color schemes
- ❌ Inconsistent spacing (mixing 8px and 12px randomly)
- ❌ Forgettable shadows and borders
- ❌ One-size-fits-all approach

Instead, each direction has:
- ✅ Distinctive typography choices
- ✅ Intentional color palettes
- ✅ Systematic spacing scales
- ✅ Characteristic component styles
- ✅ Purpose-built motion patterns

## Best Practices

1. **Always Present Portfolio First** - Don't skip the selection step
2. **Load Complete System** - Read the full design-systems.md section
3. **Generate Tokens Early** - For new projects, output config files immediately
4. **Stay Consistent** - Apply the chosen system systematically
5. **Reference Anti-Patterns** - Check what NOT to do for each direction
6. **Document Deviations** - If customizing, write down the changes
7. **Test Accessibility** - Verify color contrast and keyboard navigation
8. **Iterate with System** - New components should match existing ones

## Credits

Inspired by:
- Anthropic's "Improving frontend design through Skills" blog post
- Community skills: claude-design-skill, ui-designer
- Best practices from official Claude Code skills documentation

## License

Part of the monay project. See repository root for license.

## Feedback

Found issues or have suggestions? Open an issue in the monay repository or update this skill directly.
