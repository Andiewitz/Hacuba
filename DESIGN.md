# Design System

This document defines how color and typography are used across the site/app. It is the source of truth — component work should reference these tokens rather than picking new values ad hoc. All contrast ratios below are calculated against WCAG 2.1 relative luminance, not estimated.

## 1. Color System

### 1.1 Core Palette

| Role | Name | Hex | Notes |
|------|------|-----|-------|
| Primary background | Dark Forest Green | `#1C352D` | The brand's default surface. This is a dark-first system — forest green is the primary background, not an accent on a light page. |
| Brand accent | Terracotta / Coral | `#D25C43` | Buttons, highlights, icon fills, key interactive moments. Used deliberately, not as a background for large areas. |
| Light background | Warm Off-White / Cream | `#F6F3E6` | Secondary surface for light sections, light-mode contexts, print-adjacent material. |
| Secondary accent | Sage / Pistachio | `#C2D8A6` | Tags, badges, subtle fills, success-adjacent states. A fill color — never used as a text color (see 1.4). |

Why forest green leads, not cream: cream + terracotta is one of the most common default palettes in web design right now — on its own it reads as a template choice, not a brand choice. Leading with the dark forest surface, and treating cream as a secondary light context rather than the default page background, is what keeps this palette feeling like this brand instead of a brand. Don't let cream quietly become the default background just because it's easy to reach for.

### 1.2 Semantic Tokens

```css
:root {
  /* Core palette */
  --color-forest: #1C352D;
  --color-terracotta: #D25C43;
  --color-cream: #F6F3E6;
  --color-sage: #C2D8A6;
  --color-white: #FFFFFF;

  /* Derived — see 1.3 for how these were built */
  --color-forest-surface-1: #334942;   /* card/elevated surface on dark bg */
  --color-forest-surface-2: #495D57;   /* modal/popover on dark bg */
  --color-forest-border-subtle: #495D57;
  --color-forest-border-strong: #778681;
  --color-terracotta-hover: #BD533C;
  --color-terracotta-active: #A84A36;
  --color-terracotta-ink: #A84A36;     /* terracotta as body-sized text — see 1.4 */
  --color-sage-hover: #AFC295;
  --color-cream-surface: #EBEADD;      /* card on cream bg */
  --color-cream-border: #E0E0D4;

  /* Text on dark (forest) background */
  --text-primary-on-dark: #F6F3E6;     /* headlines, highest emphasis */
  --text-body-on-dark: #D2D7D5;        /* default paragraph text — softer than pure cream */
  --text-secondary-on-dark: #A4AEAB;
  --text-muted-on-dark: #778681;       /* large text / icons / UI only, not small body copy */

  /* Text on light (cream) background */
  --text-primary-on-light: #1C352D;
  --text-secondary-on-light: #52645B;
  --text-muted-on-light: #68786E;      /* large text / UI only */

  /* Backgrounds */
  --bg-primary: var(--color-forest);
  --bg-light: var(--color-cream);
}
```

### 1.3 Derived Tints & Shades

Rather than pulling in generic grays, every neutral/hover/border step below is mixed from the four brand colors — mostly from forest green — so the whole neutral scale still feels native to the brand instead of looking like a default UI kit was dropped on top of it.

| Token | Hex | Built as |
|-------|-----|----------|
| forest-surface-1 | `#334942` | forest + 10% white |
| forest-surface-2 | `#495D57` | forest + 20% white |
| forest-border-subtle | `#495D57` | forest + 20% white |
| forest-border-strong | `#778681` | forest + 40% white |
| text-body-on-dark | `#D2D7D5` | forest + 80% white |
| text-secondary-on-dark | `#A4AEAB` | forest + 60% white |
| text-muted-on-dark | `#778681` | forest + 40% white |
| terracotta-hover | `#BD533C` | terracotta + 10% black |
| terracotta-active / terracotta-ink | `#A84A36` | terracotta + 20% black |
| sage-hover | `#AFC295` | sage + 10% black |
| cream-surface | `#EBEADD` | cream + 5% forest |
| cream-border | `#E0E0D4` | cream + 10% forest |
| text-secondary-on-light | `#52645B` | cream + 75% forest |
| text-muted-on-light | `#68786E` | cream + 65% forest |

### 1.4 Verified Contrast Pairings

Ratios below are exact (WCAG relative luminance formula). "AA normal" means safe for body-sized text (≥4.5:1). "Large/UI only" means safe for large or bold text and UI components (≥3:1) but not for regular body copy.

| Foreground | Background | Ratio | Rating | Use for |
|-----------|------------|-------|--------|---------|
| Forest `#1C352D` | Cream `#F6F3E6` | 11.83:1 | AAA | Primary body text on light sections |
| Forest `#1C352D` | Sage `#C2D8A6` | 8.56:1 | AAA | Text on sage tags/badges — the only correct text color on sage |
| Forest `#1C352D` | White `#FFFFFF` | 13.15:1 | AAA | Primary body text on white |
| Text-body-on-dark `#D2D7D5` | Forest `#1C352D` | 9.03:1 | AAA | Default body text on the dark primary surface |
| Text-secondary-on-dark `#A4AEAB` | Forest `#1C352D` | 5.78:1 | AA normal | Secondary text on dark surface |
| Text-secondary-on-light `#52645B` | Cream `#F6F3E6` | 5.67:1 | AA normal | Secondary text on cream |
| Terracotta-ink `#A84A36` | Cream `#F6F3E6` | 5.11:1 | AA normal | Terracotta-hued body-sized text/links on cream |
| Terracotta-ink `#A84A36` | White `#FFFFFF` | 5.68:1 | AA normal | Terracotta-hued body-sized text on white |
| Black | Terracotta `#D25C43` | 5.35:1 | AA normal | Body-sized text directly on a terracotta fill (e.g. a filled button label) |
| Terracotta `#D25C43` | Cream `#F6F3E6` | 3.53:1 | Large/UI only | Terracotta as text: large/bold headings only, not body copy |
| Terracotta `#D25C43` | White `#FFFFFF` | 3.92:1 | Large/UI only | Same rule — large/bold only |
| Terracotta `#D25C43` | Forest `#1C352D` | 3.35:1 | Large/UI only | Terracotta text directly on forest: large/bold only. For body-sized terracotta-toned content on forest, use it as an icon/fill instead of text. |
| Text-muted-on-dark `#778681` | Forest `#1C352D` | 3.45:1 | Large/UI only | Disabled states, meta labels, icons on dark surface |

### 1.5 Pairings to Avoid

These aren't style opinions — they fail contrast outright:

- **Terracotta directly on sage** (2.55:1) — fails even at large-text size. Never combine these two as foreground/background of each other.
- **Cream directly on sage, or sage directly on cream, with no separation** (1.38:1) — the two are close enough in luminance that they read as flat and undifferentiated. If they sit next to each other, add a border, shadow, or gap; never rely on the color difference alone.
- **White or cream text on sage** (1.54:1 / 1.38:1) — fails.
- **Sage as a text color, at any darkness** — even sage mixed 50% toward forest only reaches 3.57:1 against cream, still short of the 4.5:1 body-text minimum. Sage is a fill/surface color only. It never carries text itself — text on sage is always forest green (see 1.4).
- **Pure terracotta as body-sized text anywhere** — it only clears "large/bold" thresholds (3.35–3.92:1 depending on background), never the 4.5:1 body-text minimum. Swap to terracotta-ink (`#A84A36`) for any terracotta-hued text below ~18px/bold-14px.
- **Treating forest as a "neutral that can still go darker" against true black** — forest vs. black is only 1.60:1. Forest is effectively already your darkest neutral; don't manufacture a near-black "shade" of it for extra emphasis and expect it to read as meaningfully darker.

## 2. Typography

### 2.1 Typeface Roles

| Role | Typeface | Why |
|------|----------|-----|
| Headings, display, brand-forward moments | Figtree | Geometric-humanist sans with warm, rounded terminals. Has real character at heavier weights (600–800) without tipping into decorative — carries the brand's personality. |
| Body copy, UI, forms, data | Inter | Purpose-built for screen legibility at small sizes. Near-neutral by design so it doesn't compete with content — the right choice for anything read at length or scanned quickly. |

Two families, clearly distinct roles. Don't blend them within the same text element (e.g. don't set a heading partially in Inter for "contrast") — the split itself is the contrast.

### 2.2 Loading

```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Figtree:wght@500;600;700;800&family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
```

```css
:root {
  --font-heading: 'Figtree', 'Segoe UI', system-ui, sans-serif;
  --font-body: 'Inter', 'Segoe UI', system-ui, sans-serif;
}
```

Only the weights actually used are loaded (Figtree 500/600/700/800, Inter 400/500/600/700) — fewer weight files means faster load, and it forces the discipline of not introducing a random 300 or 900 somewhere later without a reason.

### 2.3 Type Scale

Base unit: 16px / 1rem. All headings use Figtree; all body/UI rows use Inter.

| Level | Size | Line-height | Weight | Tracking | Typeface |
|-------|------|-------------|--------|----------|----------|
| Display | 3.5rem / 56px | 1.1 | 700 | -0.02em | Figtree |
| H1 | 2.75rem / 44px | 1.15 | 700 | -0.015em | Figtree |
| H2 | 2.25rem / 36px | 1.2 | 600 | -0.01em | Figtree |
| H3 | 1.75rem / 28px | 1.25 | 600 | -0.005em | Figtree |
| H4 | 1.375rem / 22px | 1.3 | 500 | 0 | Figtree |
| H5 (card/section titles) | 1.125rem / 18px | 1.35 | 500 | 0 | Figtree |
| Body — lead | 1.125rem / 18px | 1.6 | 400 | 0 | Inter |
| Body — default | 1rem / 16px | 1.6 | 400 | 0 | Inter |
| Body — small | 0.875rem / 14px | 1.5 | 400 | 0 | Inter |
| Caption / meta | 0.75rem / 12px | 1.4 | 400 | +0.01em | Inter |
| Button / UI label | 0.875rem / 14px | 1 | 600 | 0 | Inter |

### 2.4 Weight Discipline

- Figtree headings step down in weight as they step down in size: 700 for Display/H1, 600 for H2/H3, 500 for H4/H5. Don't set every heading level at 700 — reserve the heaviest weight for the top of the hierarchy so it still means something.
- Inter body text stays at 400. Use 500 only for a lead paragraph or an inline emphasis; use 600 for UI labels and button text specifically, not for body emphasis.
- Button and label text is sentence case, never all-caps. All-caps UI labels are one of the more obvious "default template" tells — skip it.

### 2.5 Measure & Rhythm

- Body text line length: keep under 80 characters, ideally 60–75. Constrain paragraph containers with a `max-width` in `ch` units rather than relying on the viewport.
- Don't rely on a single bolded or colored word inside an otherwise plain headline as a hierarchy device — if a headline needs emphasis, that's a sizing/weight decision at the type-scale level, not a one-word trick.
- Spend the loudest weight and the largest jump in the scale in one place per screen. If Display and H1 both appear together, one of them should visually recede (via color/weight) so there's a single clear entry point, not two competing ones.

## 3. Spacing

8px base grid.

```css
:root {
  --space-1: 0.25rem;  /* 4px  */
  --space-2: 0.5rem;   /* 8px  */
  --space-3: 0.75rem;  /* 12px */
  --space-4: 1rem;     /* 16px */
  --space-6: 1.5rem;   /* 24px */
  --space-8: 2rem;     /* 32px */
  --space-12: 3rem;    /* 48px */
  --space-16: 4rem;    /* 64px */
  --space-24: 6rem;    /* 96px */
  --space-32: 8rem;    /* 128px */
}
```

Use `--space-4`/`--space-6` for internal component padding, `--space-8`–`--space-16` between related elements within a section, and `--space-24`/`--space-32` between distinct page sections.

## 4. Radius, Elevation & Shadow

Radius is used hierarchically, not applied uniformly to every element — one flat radius on every card, button, and input regardless of size is a generic default, not a decision.

```css
:root {
  --radius-sm: 6px;    /* inputs, small tags, checkboxes */
  --radius-md: 10px;   /* buttons, standard cards */
  --radius-lg: 16px;   /* prominent/hero cards, modals */
  --radius-full: 9999px; /* pills, sage badges/tags */

  /* Shadow tinted from forest green, not generic black */
  --shadow-sm: 0 1px 3px rgba(28, 53, 45, 0.10);
  --shadow-md: 0 4px 12px rgba(28, 53, 45, 0.14);
  --shadow-lg: 0 12px 32px rgba(28, 53, 45, 0.18);
}
```

Using `rgba(28, 53, 45, …)` (forest, low opacity) instead of `rgba(0, 0, 0, …)` for shadows keeps elevation feeling like it belongs to this palette instead of the default soft-grey shadow every SaaS card kit ships with.

## 5. Component Quick Reference

| Component | Background | Text | Border | Radius |
|-----------|-----------|------|--------|--------|
| Primary button | terracotta | Black or text-body-on-dark (18px+/14px bold) | none | radius-md |
| Primary button (hover/active) | terracotta-hover / terracotta-active | same | none | radius-md |
| Secondary button (on dark) | transparent | text-body-on-dark | forest-border-strong | radius-md |
| Secondary button (on light) | transparent | forest | cream-border | radius-md |
| Card (on dark page) | forest-surface-1 | text-body-on-dark | forest-border-subtle (optional) | radius-lg |
| Card (on light page) | cream-surface | forest | cream-border (optional) | radius-lg |
| Tag / badge | sage | forest (always — never white/cream/terracotta) | none | radius-full |
| Input field | cream-surface or forest-surface-1 | matches page context | cream-border / forest-border-subtle | radius-sm |
| Link (on light) | — | terracotta-ink | underline on hover | — |
| Link (on dark) | — | text-body-on-dark | underline on hover | — |

## 6. Accessibility Cheat Sheet

Fastest lookup — the pairings that are always safe for body text:

- ✅ Forest text on cream, sage, or white
- ✅ text-body-on-dark / text-secondary-on-dark on forest
- ✅ text-secondary-on-light on cream
- ✅ terracotta-ink (`#A84A36`) as body-sized terracotta text/links on cream or white
- ✅ Black text on a solid terracotta fill

- ❌ Pure terracotta as small/body text (large/bold headings only)
- ❌ Sage as a text color, ever
- ❌ White or cream text on sage
- ❌ Terracotta directly against sage, in either direction
- ❌ Cream and sage touching with no border/shadow between them
