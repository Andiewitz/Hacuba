import Link from "next/link";
import HacubaLogo from "@/components/hacuba-logo";

const GITHUB_URL = "https://github.com/Andiewitz/Hacuba";

const columns = [
  {
    heading: "Explore",
    links: [
      { label: "All listings", href: "/" },
      { label: "Homes", href: "/homes" },
      { label: "Lots", href: "/lots" },
      { label: "Commercial", href: "/commercial" },
    ],
  },
  {
    heading: "Support",
    links: [
      { label: "Help center", href: "/help" },
      { label: "GitHub", href: GITHUB_URL, external: true },
    ],
  },
  {
    heading: "Legal",
    links: [
      { label: "Privacy Policy", href: "/privacy" },
      { label: "Terms of Service", href: "/terms" },
    ],
  },
];

export default function Footer() {
  return (
    <footer className="mt-auto bg-[var(--color-forest)] text-[var(--text-body-on-dark)]">
      <div className="mx-auto max-w-[1760px] px-6 py-12 md:px-10 lg:px-20">
        <div className="grid gap-8 md:grid-cols-[1.4fr_1fr_1fr_1fr]">
          <div>
            <Link href="/" className="inline-flex items-center gap-2">
              <HacubaLogo
                iconOnly
                className="text-[var(--text-primary-on-dark)]"
              />
              <span className="font-heading text-xl font-semibold tracking-tight text-[var(--text-primary-on-dark)]">
                Hacuba
              </span>
            </Link>
            <p className="mt-4 max-w-[38ch] text-sm leading-6 text-[var(--text-secondary-on-dark)]">
              Homes, lots, and commercial spaces in and around Cebu City —
              verified before they appear in search.
            </p>
          </div>

          {columns.map((col) => (
            <nav key={col.heading} aria-label={`Footer — ${col.heading}`}>
              <h2 className="font-heading text-lg font-medium text-[var(--text-primary-on-dark)]">
                {col.heading}
              </h2>
              <ul className="mt-4 flex flex-col gap-3">
                {col.links.map((link) =>
                  "external" in link && link.external ? (
                    <li key={link.label}>
                      <a
                        href={link.href}
                        target="_blank"
                        rel="noreferrer"
                        className="text-sm text-[var(--text-secondary-on-dark)] transition-colors hover:text-[var(--text-primary-on-dark)] hover:underline"
                      >
                        {link.label}
                      </a>
                    </li>
                  ) : (
                    <li key={link.label}>
                      <Link
                        href={link.href}
                        className="text-sm text-[var(--text-secondary-on-dark)] transition-colors hover:text-[var(--text-primary-on-dark)] hover:underline"
                      >
                        {link.label}
                      </Link>
                    </li>
                  ),
                )}
              </ul>
            </nav>
          ))}
        </div>

        <div className="mt-12 flex flex-col gap-2 border-t border-[var(--color-forest-border-subtle)] pt-6 text-xs text-[var(--text-muted-on-dark)] sm:flex-row sm:items-center sm:justify-between">
          <p>© 2026 Hacuba. All rights reserved.</p>
          <p>Cebu City listings — view in person before sending money.</p>
        </div>
      </div>
    </footer>
  );
}
