# Hacuba

A real estate listing platform built as a portfolio project demonstrating modern full-stack development practices.

## Tech Stack

- **Framework:** [Next.js 16](https://nextjs.org/) (App Router, Turbopack)
- **Language:** TypeScript
- **Styling:** [Tailwind CSS v4](https://tailwindcss.com/)
- **UI Components:** [shadcn/ui](https://ui.shadcn.com/)
- **State Management:** [Redux Toolkit](https://redux-toolkit.js.org/)
- **Animations:** [Framer Motion](https://www.framer.com/motion/)
- **Icons:** [Lucide React](https://lucide.dev/)

## Features

- Category-based browsing (Homes, Lots, Commercial)
- Animated search with segment pill navigation
- Location, price, and property type filtering
- Responsive grid layout with scroll-reveal animations
- Custom design system with verified WCAG 2.1 contrast ratios
- Dark-first color palette (forest green primary)

## Getting Started

```bash
# Install dependencies
npm install

# Run development server
npm run dev --workspace client

# Build for production
npm run build --workspace client
```

Open [http://localhost:3000](http://localhost:3000) to view the app.

## Project Structure

```
Hacuba/
├── client/                  # Next.js frontend
│   ├── src/
│   │   ├── app/             # App Router pages
│   │   ├── components/      # Reusable UI components
│   │   ├── data/            # Listing data
│   │   └── lib/             # Store, hooks, utils
│   └── public/              # Static assets
├── docs/                    # Documentation
├── scripts/                 # Build/deploy scripts
├── services/                # Backend (planned)
├── terraform/               # Infrastructure (planned)
├── tests/                   # Tests (planned)
├── DESIGN.md                # Design system tokens
├── AGENTS.md                # AI agent workflow rules
└── LICENSE                  # MIT License
```

## Design System

See [DESIGN.md](./DESIGN.md) for the full design system including color tokens, typography scale, spacing grid, and contrast-verified pairings.

## License

[MIT](LICENSE)
