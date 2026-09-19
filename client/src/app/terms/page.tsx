import type { Metadata } from "next";
import Link from "next/link";

export const metadata: Metadata = {
  title: "Terms of Service — Hacuba",
  description: "The rules for using Hacuba listings and accounts.",
};

const sections = [
  {
    h: "What Hacuba is",
    p: "Hacuba lists homes, lots, and commercial properties in and around Cebu City. Listings are reviewed against photo, location, and ownership checks before they appear, but every deal is ultimately between you and the seller.",
  },
  {
    h: "Accounts",
    p: "You need an account to save homes and pick up across devices. Use a valid email and a password of 8–72 characters with at least one letter and one number. You are responsible for activity under your account. After 5 failed logins for an email within 15 minutes, sign-in is temporarily rate-limited.",
  },
  {
    h: "Acceptable use",
    p: "No scraping at abusive rates, no probing other users' accounts or sessions, no posting false listings, and no using the service to defraud or harass anyone. Cross-account access attempts are treated as abuse and may lead to suspension.",
  },
  {
    h: "Listings disclaimer",
    p: "Treat every listing as a starting point, not a guarantee. View the property in person, verify ownership and documents independently, and never send money before viewing. If something looks wrong, report it so it can be reviewed.",
  },
  {
    h: "Availability and liability",
    p: "The service is provided as-is. We aim for reliable search, saves, and sign-in, but we do not guarantee uninterrupted availability or the accuracy of third-party listing content. To the extent permitted by law, Hacuba is not liable for deals or losses arising from listings.",
  },
  {
    h: "Changes",
    p: "If these terms change materially, the updated date at the top of this page changes with them. Continued use after changes take effect means you accept the updated terms.",
  },
];

export default function TermsPage() {
  return (
    <div className="min-h-screen bg-background">
      <main className="mx-auto max-w-3xl px-6 pt-12 pb-24 md:px-10">
        <p className="text-sm font-semibold text-muted-foreground">
          Legal
        </p>
        <h1 className="mt-2 font-heading text-4xl leading-tight font-bold">
          Terms of Service
        </h1>
        <p className="mt-4 max-w-[65ch] leading-7 text-muted-foreground">
          Last updated September 19, 2026. The plain-language rules for
          accounts, acceptable use, and listings.
        </p>

        <div className="mt-8 flex flex-col gap-4">
          {sections.map((item) => (
            <section
              key={item.h}
              className="rounded-lg border border-border bg-card p-6 text-card-foreground shadow-sm"
            >
              <h2 className="font-heading text-lg font-medium">{item.h}</h2>
              <p className="mt-2 max-w-[70ch] text-sm leading-6 text-muted-foreground">
                {item.p}
              </p>
            </section>
          ))}
        </div>

        <p className="mt-8 max-w-[70ch] text-sm leading-6 text-muted-foreground">
          Related: our{" "}
          <Link href="/privacy" className="underline hover:text-foreground">
            Privacy Policy
          </Link>{" "}
          explains what data we collect and why.
        </p>
      </main>
    </div>
  );
}
