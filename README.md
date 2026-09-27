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
- **Services:** Go Auth and Listings APIs with PostgreSQL in production
- **Production infrastructure:** AWS RDS, S3, CloudFront, IAM, and Secrets Manager via Terraform

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

# Install client dependencies and run the client
npm install
npm run dev --workspace client

# Build for production
npm run build --workspace client
```

Open [http://localhost:3000](http://localhost:3000) to view the app.

For the complete local stack without Docker, set the ignored `client/.env.local`
from `client/.env.example`, then start each API with an explicit development
SQLite path. `DEV_SQLITE_PATH` is intentionally rejected in production.

```powershell
$env:DEV_SQLITE_PATH = ".dev/auth.db"; $env:JWT_SECRET = "a-development-secret-at-least-32-characters"; go run ./cmd/server
# Run the Listings service in a second terminal, with the same JWT_SECRET.
$env:DEV_SQLITE_PATH = ".dev/listings.db"; $env:DEV_UPLOAD_URL = "http://localhost:3000/api/dev-uploads"; $env:JWT_SECRET = "a-development-secret-at-least-32-characters"; go run ./cmd/server
```

Run those commands from `services/auth` and `services/listings` respectively.
For production provisioning, migration, IAM attachment, and deployment values,
see [terraform/README.md](./terraform/README.md).

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
├── services/                # Auth and Listings Go APIs
├── terraform/               # Production AWS infrastructure
├── tests/                   # Tests (planned)
├── DESIGN.md                # Design system tokens
├── AGENTS.md                # AI agent workflow rules
└── LICENSE                  # MIT License
```

## Design System

See [DESIGN.md](./DESIGN.md) for the full design system including color tokens, typography scale, spacing grid, and contrast-verified pairings.

## License

[MIT](LICENSE)
