"use client";

import { AlertTriangle, Bug, Building2, CircleHelp, ShieldAlert } from "lucide-react";
import { useState } from "react";
import { useAppSelector } from "@/lib/hooks";

const faqs = [
  { q: "How do I contact a seller?", a: "Open a property and use the contact details provided by its seller. Hacuba does not handle viewings, contracts, or payments." },
  { q: "How do I stay safe?", a: "View a property in person, verify ownership and documents independently, and never send money before a viewing." },
  { q: "Can I save properties?", a: "Saved properties are planned but are not available yet. Keep a listing link while you compare options." },
];

const categories = [
  { value: "listing", label: "A listing looks wrong", icon: Building2 },
  { value: "safety", label: "Safety concern", icon: ShieldAlert },
  { value: "bug", label: "Something is broken", icon: Bug },
  { value: "account", label: "Account help", icon: CircleHelp },
];

type Fields = { category?: string; description?: string; contact_email?: string; listing_reference?: string };

export default function HelpPage() {
  const { accessToken, initialized } = useAppSelector((state) => state.auth);
  const [category, setCategory] = useState("listing");
  const [listingReference, setListingReference] = useState("");
  const [email, setEmail] = useState("");
  const [description, setDescription] = useState("");
  const [fields, setFields] = useState<Fields>({});
  const [status, setStatus] = useState<"idle" | "sending" | "sent" | "error">("idle");
  const [message, setMessage] = useState("");

  const submit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setStatus("sending"); setFields({}); setMessage("");
    const response = await fetch("/api/support/reports", { method: "POST", headers: { "Content-Type": "application/json", ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}) }, body: JSON.stringify({ category, listing_reference: listingReference, contact_email: email, description }) });
    const body = await response.json().catch(() => null) as { error?: string; fields?: Fields } | null;
    if (!response.ok) { setFields(body?.fields ?? {}); setMessage(body?.error ?? "We could not send your report. Please try again."); setStatus("error"); return; }
    setDescription(""); setListingReference(""); setEmail("");
    setMessage("Your report is with the Hacuba team. Thank you for helping keep the marketplace useful and safe."); setStatus("sent");
  };

  return <main className="min-h-screen bg-background pb-[var(--space-24)]">
    <section className="bg-[var(--color-forest)] px-6 py-[var(--space-16)] md:px-10 lg:px-20"><div className="mx-auto max-w-6xl"><p className="text-sm font-semibold text-[var(--text-secondary-on-dark)]">Help center</p><h1 className="mt-3 max-w-2xl font-heading text-[2.75rem] font-bold leading-[1.15] tracking-[-0.015em] text-[var(--text-primary-on-dark)] md:text-[3.5rem] md:leading-[1.1] md:tracking-[-0.02em]">How can we help?</h1><p className="mt-4 max-w-[62ch] text-[1.125rem] leading-[1.6] text-[var(--text-body-on-dark)]">Get clear answers, report a problem, and help us keep Hacuba safer for people looking for property in Cebu.</p></div></section>
    <div className="mx-auto grid max-w-6xl gap-[var(--space-12)] px-6 py-[var(--space-16)] md:px-10 lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)] lg:px-20">
      <section aria-labelledby="answers-heading"><p className="text-sm font-semibold text-[var(--color-terracotta-ink)]">Answers</p><h2 id="answers-heading" className="mt-2 font-heading text-[2.25rem] font-semibold leading-[1.2] tracking-[-0.01em] text-foreground">Useful before you reach out</h2><div className="mt-[var(--space-8)] space-y-[var(--space-4)]">{faqs.map((item) => <article key={item.q} className="rounded-[var(--radius-lg)] border border-border bg-card p-[var(--space-6)] shadow-[var(--shadow-sm)]"><h3 className="font-heading text-[1.375rem] font-medium leading-[1.3] text-foreground">{item.q}</h3><p className="mt-[var(--space-3)] max-w-[62ch] text-sm leading-6 text-muted-foreground">{item.a}</p></article>)}</div><aside className="mt-[var(--space-8)] rounded-[var(--radius-lg)] border border-[var(--color-cream-border)] bg-[var(--color-sage)] p-[var(--space-6)] shadow-[var(--shadow-sm)]"><div className="flex gap-[var(--space-3)]"><AlertTriangle className="mt-1 h-5 w-5 shrink-0 text-[var(--color-forest)]" aria-hidden="true" /><div><h3 className="font-heading text-[1.375rem] font-medium text-[var(--color-forest)]">If money has been requested</h3><p className="mt-2 max-w-[54ch] text-sm leading-6 text-[var(--color-forest)]">Stop the conversation, do not share a code or make a transfer, and file a safety report below. For immediate danger, contact local emergency services.</p></div></div></aside></section>
      <section aria-labelledby="report-heading" className="rounded-[var(--radius-lg)] border border-border bg-card p-[var(--space-6)] shadow-[var(--shadow-md)] md:p-[var(--space-8)]"><p className="text-sm font-semibold text-[var(--color-terracotta-ink)]">Report an issue</p><h2 id="report-heading" className="mt-2 font-heading text-[2.25rem] font-semibold leading-[1.2] tracking-[-0.01em] text-foreground">Tell us what happened</h2><p className="mt-3 max-w-[60ch] text-sm leading-6 text-muted-foreground">Reports are reviewed privately. We record only aggregate categories in our operations dashboard.</p><form className="mt-[var(--space-8)] space-y-[var(--space-6)]" onSubmit={submit}>
        <fieldset><legend className="text-sm font-semibold text-foreground">What do you need help with?</legend><div className="mt-[var(--space-3)] grid gap-[var(--space-3)] sm:grid-cols-2">{categories.map(({ value, label, icon: Icon }) => <label key={value} className={`flex cursor-pointer items-center gap-3 rounded-[var(--radius-md)] border p-4 text-sm font-semibold transition-shadow ${category === value ? "border-[var(--color-terracotta)] bg-[var(--color-cream-surface)] text-foreground shadow-[var(--shadow-sm)]" : "border-border bg-background text-muted-foreground hover:shadow-[var(--shadow-sm)]"}`}><input className="sr-only" type="radio" name="category" value={value} checked={category === value} onChange={() => setCategory(value)} /><Icon className="h-4 w-4" aria-hidden="true" />{label}</label>)}</div>{fields.category && <p className="mt-2 text-sm text-[var(--color-terracotta-ink)]">{fields.category}</p>}</fieldset>
        <label className="block text-sm font-semibold text-foreground">Listing link or reference <span className="font-normal text-muted-foreground">(optional)</span><input value={listingReference} onChange={(event) => setListingReference(event.target.value)} maxLength={500} className="mt-2 w-full rounded-[var(--radius-sm)] border border-border bg-background px-3 py-3 text-sm font-normal text-foreground outline-none ring-[var(--color-terracotta)] focus:ring-2" placeholder="Paste the listing link if this is about a property" /></label>
        <label className="block text-sm font-semibold text-foreground">What happened?<textarea value={description} onChange={(event) => setDescription(event.target.value)} required minLength={20} maxLength={2000} className="mt-2 min-h-36 w-full resize-y rounded-[var(--radius-sm)] border border-border bg-background px-3 py-3 text-sm font-normal leading-6 text-foreground outline-none ring-[var(--color-terracotta)] focus:ring-2" placeholder="Share enough detail for our team to understand the issue." /></label>{fields.description && <p className="-mt-4 text-sm text-[var(--color-terracotta-ink)]">{fields.description}</p>}
        <label className="block text-sm font-semibold text-foreground">Email for a reply <span className="font-normal text-muted-foreground">(optional)</span><input value={email} onChange={(event) => setEmail(event.target.value)} type="email" maxLength={254} className="mt-2 w-full rounded-[var(--radius-sm)] border border-border bg-background px-3 py-3 text-sm font-normal text-foreground outline-none ring-[var(--color-terracotta)] focus:ring-2" placeholder="you@example.com" /></label>{fields.contact_email && <p className="-mt-4 text-sm text-[var(--color-terracotta-ink)]">{fields.contact_email}</p>}
        {status !== "idle" && <p role="status" className={status === "sent" ? "text-sm leading-6 text-[var(--text-secondary-on-light)]" : "text-sm leading-6 text-[var(--color-terracotta-ink)]"}>{message}</p>}
        <button type="submit" disabled={!initialized || status === "sending"} className="inline-flex min-h-12 items-center justify-center rounded-[var(--radius-md)] bg-[var(--color-terracotta)] px-5 text-sm font-semibold text-black transition-colors hover:bg-[var(--color-terracotta-hover)] disabled:cursor-not-allowed disabled:opacity-60">{status === "sending" ? "Sending report…" : "Send report"}</button>
      </form></section>
    </div>
  </main>;
}
