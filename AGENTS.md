# AGENTS.md

## Workflow

Before making any code changes, follow this exact sequence:

1. **Reason.** Think through the problem silently — consider what already exists in the codebase, what the task actually requires, and any edge cases or tradeoffs. Do not skip this step.

2. **Explain.** After reasoning, write a brief plan in plain language. State:
   - What files will be created or modified
   - What the change does and why
   - Any assumptions or constraints you're working with

3. **Ask for approval.** End with a clear prompt:
   > Type `1` to proceed.

   Do not proceed until the user replies `1`. If the user says anything else or gives different instructions, follow those instead.

4. **Execute.** Once approved, make the changes, run lint/typecheck/build, and confirm success.

## Design System

Every component, layout, and style token must conform to `DESIGN.md` at the project root. Before writing any component:

- Read `DESIGN.md` for the active color tokens, type scale, spacing, radius, shadow, and contrast rules.
- Use the semantic CSS variables (`--color-forest`, `--color-terracotta`, `--text-body-on-dark`, etc.), not raw hex values.
- Follow the type scale — headings use Figtree, body/UI uses Inter, with the correct weight per level.
- Respect contrast pairings. Never use a pairing from the "avoid" list.
- Use the 8px spacing grid. Pick the nearest `--space-*` token.
- Use hierarchical radius (`--radius-sm/md/lg/full`), not uniform rounding.
- Use forest-tinted shadows (`--shadow-sm/md/lg`), not generic black.

When in doubt, re-read `DESIGN.md` before writing code. The design system is the source of truth — not assumptions, not convenience.
