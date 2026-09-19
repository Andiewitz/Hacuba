import type { Metadata } from "next";
import Link from "next/link";

export const metadata: Metadata = {
  title: "Privacy Policy — Hacuba",
  description: "How Hacuba collects, uses, and protects your data.",
};

const sections = [
  {
    h: "Data we collect",
    p: "Account data: your email address and password hash when you register or log in. Product data: homes you save and listings you view, so picks follow you across devices. Technical data: basic device, log, and security signals (for example failed-login counters) used to keep accounts safe.",
  },
  {
    h: "Cookies and sessions",
    p: "Sign-in uses a short-lived access token kept in memory (never in localStorage), a 7-day HttpOnly refresh cookie, and a CSRF cookie plus header pair on mutating requests. These cookies are required for login, logout, and session refresh — blocking them means you stay logged out.",
  },
  {
    h: "How we use data",
    p: "To create and protect your account, remember saved homes, enforce rate limits on login (5 failures per email per 15 minutes), and detect abuse. We do not sell personal data and do not use it for third-party advertising.",
  },
  {
    h: "Sharing and storage",
    p: "Identity data lives in a private auth database reachable only by the auth service — no other service holds its credentials. Passwords are stored as bcrypt hashes and never returned by the API. We share data only when required by law or to protect users and the service.",
  },
  {
    h: "Retention and your choices",
    p: "We keep account data while your account exists. Logging out clears your session cookies; refresh sessions are single-use and rotate on refresh. Contact us through the repository below to request access or deletion of your account data.",
  },
];

export default function PrivacyPage() {
  return (
    <div className="min-h-screen bg-background">
      <main className="mx-auto max-w-3xl px-6 pt-12 pb-24 md:px-10">
        <p className="text-sm font-semibold text-muted-foreground">
          Legal
        </p>
        <h1 className="mt-2 font-heading text-4xl leading-tight font-bold">
          Privacy Policy
        </h1>
        <p className="mt-4 max-w-[65ch] leading-7 text-muted-foreground">
          Last updated September 19, 2026. Short version: your email runs
          your account, your saves run your picks, and your password never
          leaves the auth service in readable form.
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
          Questions about this policy? See the{" "}
          <Link href="/terms" className="underline hover:text-foreground">
            Terms of Service
          </Link>{" "}
          or reach us through the project repository.
        </p>
      </main>
    </div>
  );
}
