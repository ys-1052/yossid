# yossid Login Screen Design Proposals

To establish a strong and premium visual identity for **yossid**, we have designed five distinct aesthetic directions for the login interface.

We have created **fully interactive HTML/CSS mockups** for each concept in the workspace so you can test them live on your device.

---

## Concept Carousel

````carousel
### Concept 1: Deep Obsidian & Aurora Glow (Modern Dark Mode)

![Aurora Dark Design Mockup](/Users/ytakahashi/.gemini/antigravity-ide/brain/be6e21ca-e98b-4196-af84-848beee66a39/login_aurora_dark_1782643618954.png)

A tech-forward, high-security, premium dark mode interface with floating glassmorphism and an interactive background aurora blur.

- **Interactive Mockup HTML File**: [aurora_dark.html](file:///Users/ytakahashi/app/yossid/infra/mockups/aurora_dark.html)
- **Palette**: Obsidian Navy (`#07090E`), Glass Fill (`rgba(13, 17, 28, 0.7)`), Indigo-to-Purple Gradient (`#6366F1` ➔ `#A855F7`)
- **Typography**: Display: `Outfit` (Bold Sans), Body: `Inter` (UI Sans)
- **Aesthetic Risk**: Ambient floating glow animation behind a glassmorphic card.

<!-- slide -->
### Concept 2: Neo-Brutalist Bauhaus (Structured High-Contrast)

![Neo-Brutalist Bauhaus Design Mockup](/Users/ytakahashi/.gemini/antigravity-ide/brain/be6e21ca-e98b-4196-af84-848beee66a39/login_brutalist_bauhaus_1782643630391.png)

A high-contrast, structured layout utilizing sharp 90-degree corners, thick borders, solid flat offsets, and metadata tagging to convey absolute transparency, speed, and developer-first structure.

- **Interactive Mockup HTML File**: [brutalist_bauhaus.html](file:///Users/ytakahashi/app/yossid/infra/mockups/brutalist_bauhaus.html)
- **Palette**: Alabaster White (`#FAF9F6`), Ink Black (`#111111`), Cobalt Blue (`#2563EB`)
- **Typography**: Display: `Space Grotesk` (Geometric Sans), Body: `JetBrains Mono` (Monospace)
- **Aesthetic Risk**: Intentionally zero border-radius, heavy ink borders, and hard-edged flat shadows.

<!-- slide -->
### Concept 3: Swiss Editorial (Quiet Luxury Minimalist)

![Swiss Editorial Design Mockup](/Users/ytakahashi/.gemini/antigravity-ide/brain/be6e21ca-e98b-4196-af84-848beee66a39/login_swiss_mono_1782643643651.png)

An understated, quiet luxury design inspired by high-end print magazines and architectural journals. Focuses on generous whitespace, asymmetrical serif titles, and hairline rules.

- **Interactive Mockup HTML File**: [swiss_mono.html](file:///Users/ytakahashi/app/yossid/infra/mockups/swiss_mono.html)
- **Palette**: Silk Grey (`#F3F3F3`), Solid Pitch Black (`#000000`), Charcoal Text (`#1A1A1A`)
- **Typography**: Display: `Playfair Display` (Classic Serif), Body: `Inter` (Neutral Sans)
- **Aesthetic Risk**: High-contrast serif headings on a clean white-and-gray grid layout.

<!-- slide -->
### Concept 4: Cyber-Retro Terminal (Phosphor Console Vibe)

![Retro Terminal Design Mockup](/Users/ytakahashi/.gemini/antigravity-ide/brain/be6e21ca-e98b-4196-af84-848beee66a39/login_retro_terminal_1782643933111.png)

A geek-chic, retro-futurism terminal console layout with green phosphor glowing text on a pitch black background, CRT scanlines, and a retro computer screen aesthetic.

- **Interactive Mockup HTML File**: [retro_terminal.html](file:///Users/ytakahashi/app/yossid/infra/mockups/retro_terminal.html)
- **Palette**: Terminal Black (`#07090B`), Phosphor Green (`#39FF14`), Cyan Neon (`#00FFFF`)
- **Typography**: Display: `Share Tech Mono` (Utility Sans), Body: Monospace
- **Aesthetic Risk**: Full neon terminal phosphor layout that is highly nostalgic and developer-oriented.

<!-- slide -->
### Concept 5: Frosted Mint & Pearlescent (Airy Glassmorphism)

![Frosted Mint Design Mockup](/Users/ytakahashi/.gemini/antigravity-ide/brain/be6e21ca-e98b-4196-af84-848beee66a39/login_frosted_mint_1782643949237.png)

A light, fresh, optimistic light-theme glassmorphism card floating on a soft mint and warm peach pastel gradient background.

- **Interactive Mockup HTML File**: [frosted_mint.html](file:///Users/ytakahashi/app/yossid/infra/mockups/frosted_mint.html)
- **Palette**: Soft Pastels (`#E0F2FE` to `#F0FDF4`), Frosted Card (`rgba(255, 255, 255, 0.45)`), Mint Teal (`#0F766E`)
- **Typography**: Display & Body: `Plus Jakarta Sans` (Fluid Sans)
- **Aesthetic Risk**: High reliance on multi-color background gradients and frosted transparency under contrast requirements.
````

---

## Best Practices Implemented

All mockups adhere strictly to modern web standards and accessibility patterns:
- **Autofill Support**: Input configurations use `autocomplete="username"` and `autocomplete="current-password"`, standard semantic names, stable IDs, and wrap inside `<form>` elements.
- **Mobile Friendliness**: Virtual keyboards set to correct inputs (`type="email"` for email address inputs, `enterkeyhint` configured for seamless form navigation).
- **Interactivity**: Clean vanilla JS included inside each mockup file to handle password show/hide toggles seamlessly.
