"use client";

import { AlertTriangle, BadgeAlert, Building2, CheckCircle2, CopyX, Flag, ShieldAlert } from "lucide-react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useState } from "react";
import { useAppSelector } from "@/lib/hooks";

const reasons = [
  { value: "misleading_information", category: "listing", label: "Misleading information", detail: "The photos, price, location, or details do not match the property.", icon: Building2 },
  { value: "suspicious_payment", category: "safety", label: "Suspicious payment request", detail: "Someone asked for money, a deposit, or sensitive details before a viewing.", icon: ShieldAlert },
  { value: "duplicate_or_unavailable", category: "listing", label: "Duplicate or unavailable", detail: "This property is repeated, already sold, rented, or no longer available.", icon: CopyX },
  { value: "inappropriate_content", category: "listing", label: "Inappropriate content", detail: "The listing contains abusive, unlawful, or unrelated material.", icon: BadgeAlert },
  { value: "other", category: "listing", label: "Something else", detail: "Tell us what looks wrong or needs attention.", icon: Flag },
] as const;

type Fields = { reason?: string; description?: string; contact_email?: string; listing_reference?: string };

export default function HelpPage() {
  const { accessToken, initialized } = useAppSelector((state) => state.auth);
  const searchParams = useSearchParams();
  const linkedListingID = searchParams.get("listing_id") ?? "";
  const [reason, setReason] = useState<(typeof reasons)[number]["value"]>("misleading_information");
  const [listingReference, setListingReference] = useState("");
  const [description, setDescription] = useState("");
  const [email, setEmail] = useState("");
  const [fields, setFields] = useState<Fields>({});
  const [status, setStatus] = useState<"idle" | "sending" | "sent" | "error">("idle");
  const [message, setMessage] = useState("");
  const selected = reasons.find((item) => item.value === reason) ?? reasons[0];

  const submit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setStatus("sending");
    setFields({});
    setMessage("");
    const response = await fetch("/api/support/reports", {
      method: "POST",
      headers: { "Content-Type": "application/json", ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}) },
      body: JSON.stringify({
        category: selected.category,
        reason: selected.value,
        listing_id: linkedListingID || undefined,
        listing_reference: linkedListingID ? undefined : listingReference,
        contact_email: email,
        description,
      }),
    });
    const body = (await response.json().catch(() => null)) as { error?: string; fields?: Fields } | null;
    if (!response.ok) {
      setFields(body?.fields ?? {});
      setMessage(body?.error ?? "We could not send your report. Please try again.");
      setStatus("error");
      return;
    }
    setDescription("");
    setListingReference("");
    setEmail("");
    setMessage("Your report has been sent. Our team will review it privately.");
    setStatus("sent");
  };

  if (status === "sent") {
    return (
      <main className="min-h-screen bg-background px-6 py-[var(--space-16)] md:px-10 lg:px-20">
        <section className="mx-auto max-w-2xl rounded-[var(--radius-lg)] border border-border bg-card p-[var(--space-8)] text-center shadow-[var(--shadow-md)]">
          <span className="mx-auto inline-flex h-12 w-12 items-center justify-center rounded-full bg-[var(--color-sage)] text-[var(--color-forest)]"><CheckCircle2 className="h-6 w-6" /></span>
          <p className="mt-6 text-sm font-semibold text-[var(--color-terracotta-ink)]">Report received</p>
          <h1 className="mt-2 font-heading text-[2.25rem] font-semibold text-foreground">Thank you for flagging this</h1>
          <p className="mx-auto mt-3 max-w-[52ch] text-sm leading-6 text-muted-foreground">{message} We do not share your report or contact details with the seller.</p>
          <div className="mt-8 flex flex-wrap justify-center gap-3">
            {linkedListingID && <Link href={`/listings/${linkedListingID}`} className="rounded-[var(--radius-md)] border border-border px-4 py-3 text-sm font-semibold text-foreground">Back to listing</Link>}
            <Link href="/" className="rounded-[var(--radius-md)] bg-[var(--color-terracotta)] px-4 py-3 text-sm font-semibold text-black">Continue browsing</Link>
          </div>
        </section>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-background pb-[var(--space-24)]">
      <section className="bg-[var(--color-forest)] px-6 py-[var(--space-12)] md:px-10 md:py-[var(--space-16)] lg:px-20">
        <div className="mx-auto max-w-3xl">
          <p className="text-sm font-semibold text-[var(--text-secondary-on-dark)]">Safety and support</p>
          <h1 className="mt-3 font-heading text-[2.75rem] font-bold leading-[1.15] tracking-[-0.015em] text-[var(--text-primary-on-dark)] md:text-[3.5rem] md:leading-[1.1]">Report a listing</h1>
          <p className="mt-4 max-w-[60ch] text-[1.125rem] leading-[1.6] text-[var(--text-body-on-dark)]">Choose what happened, add the details we need, and we will review it privately.</p>
        </div>
      </section>
      <div className="mx-auto grid max-w-6xl gap-[var(--space-8)] px-6 py-[var(--space-12)] md:px-10 lg:grid-cols-[minmax(0,1fr)_360px] lg:px-20">
        <form className="rounded-[var(--radius-lg)] border border-border bg-card p-[var(--space-6)] shadow-[var(--shadow-md)] md:p-[var(--space-8)]" onSubmit={submit}>
          {linkedListingID ? <div className="mb-[var(--space-8)] rounded-[var(--radius-md)] border border-[var(--color-cream-border)] bg-[var(--color-sage)] p-[var(--space-4)]"><p className="text-sm font-semibold text-[var(--color-forest)]">You are reporting this listing</p><p className="mt-1 text-sm leading-6 text-[var(--color-forest)]">The property is linked automatically, so you do not need to paste a URL.</p></div> : <label className="mb-[var(--space-8)] block text-sm font-semibold text-foreground">Listing link <span className="font-normal text-muted-foreground">(optional)</span><input value={listingReference} onChange={(event) => setListingReference(event.target.value)} maxLength={500} className="mt-2 w-full rounded-[var(--radius-sm)] border border-border bg-background px-3 py-3 text-sm font-normal text-foreground outline-none focus:ring-2 focus:ring-[var(--color-terracotta)]" placeholder="Paste a Hacuba listing link" /></label>}
          <fieldset><legend className="font-heading text-[1.375rem] font-medium text-foreground">What is the problem?</legend><p className="mt-2 text-sm leading-6 text-muted-foreground">Pick the closest option. You can explain more in the next step.</p><div className="mt-[var(--space-6)] grid gap-[var(--space-3)]">{reasons.map(({ value, label, detail, icon: Icon }) => <label key={value} className={`flex cursor-pointer items-start gap-4 rounded-[var(--radius-md)] border p-[var(--space-4)] transition-shadow ${reason === value ? "border-[var(--color-terracotta)] bg-[var(--color-cream-surface)] shadow-[var(--shadow-sm)]" : "border-border bg-background hover:shadow-[var(--shadow-sm)]"}`}><input className="sr-only" type="radio" name="reason" value={value} checked={reason === value} onChange={() => setReason(value)} /><Icon className="mt-0.5 h-5 w-5 shrink-0 text-[var(--color-terracotta-ink)]" aria-hidden="true" /><span><span className="block text-sm font-semibold text-foreground">{label}</span><span className="mt-1 block text-sm leading-6 text-muted-foreground">{detail}</span></span></label>)}</div>{fields.reason && <p className="mt-3 text-sm text-[var(--color-terracotta-ink)]">{fields.reason}</p>}</fieldset>
          {reason === "suspicious_payment" && <aside className="mt-[var(--space-6)] rounded-[var(--radius-md)] border border-[var(--color-cream-border)] bg-[var(--color-sage)] p-[var(--space-4)]"><div className="flex gap-3"><AlertTriangle className="mt-0.5 h-5 w-5 shrink-0 text-[var(--color-forest)]" /><p className="text-sm leading-6 text-[var(--color-forest)]">Do not send money, verification codes, or document photos. For immediate danger, contact local emergency services.</p></div></aside>}
          <label className="mt-[var(--space-8)] block text-sm font-semibold text-foreground">What happened?<textarea value={description} onChange={(event) => setDescription(event.target.value)} required minLength={20} maxLength={2000} className="mt-2 min-h-36 w-full resize-y rounded-[var(--radius-sm)] border border-border bg-background px-3 py-3 text-sm font-normal leading-6 text-foreground outline-none focus:ring-2 focus:ring-[var(--color-terracotta)]" placeholder={reason === "suspicious_payment" ? "Tell us what was requested and how you were contacted." : "Describe what you noticed. Do not include passwords or financial information."} /></label>{fields.description && <p className="mt-2 text-sm text-[var(--color-terracotta-ink)]">{fields.description}</p>}
          <label className="mt-[var(--space-6)] block text-sm font-semibold text-foreground">Email for follow-up <span className="font-normal text-muted-foreground">(optional)</span><input value={email} onChange={(event) => setEmail(event.target.value)} type="email" maxLength={254} className="mt-2 w-full rounded-[var(--radius-sm)] border border-border bg-background px-3 py-3 text-sm font-normal text-foreground outline-none focus:ring-2 focus:ring-[var(--color-terracotta)]" placeholder="you@example.com" /></label>{fields.contact_email && <p className="mt-2 text-sm text-[var(--color-terracotta-ink)]">{fields.contact_email}</p>}
          {status === "error" && <p role="alert" className="mt-6 text-sm leading-6 text-[var(--color-terracotta-ink)]">{message}</p>}
          <div className="mt-[var(--space-8)] flex flex-wrap items-center gap-4"><button type="submit" disabled={!initialized || status === "sending"} className="inline-flex min-h-12 items-center justify-center rounded-[var(--radius-md)] bg-[var(--color-terracotta)] px-5 text-sm font-semibold text-black transition-colors hover:bg-[var(--color-terracotta-hover)] disabled:cursor-not-allowed disabled:opacity-60">{status === "sending" ? "Sending report…" : "Send report"}</button><p className="max-w-[40ch] text-sm leading-6 text-muted-foreground">Your details are visible only to the Hacuba support team.</p></div>
        </form>
        <aside className="h-fit rounded-[var(--radius-lg)] border border-border bg-card p-[var(--space-6)] shadow-[var(--shadow-sm)]"><p className="text-sm font-semibold text-[var(--color-terracotta-ink)]">Before you report</p><h2 className="mt-2 font-heading text-[1.75rem] font-semibold text-foreground">Stay safe while searching</h2><ul className="mt-5 space-y-4 text-sm leading-6 text-muted-foreground"><li>Arrange an in-person viewing before sending money.</li><li>Verify property documents and the seller independently.</li><li>Keep conversations and payment details private.</li></ul></aside>
      </div>
    </main>
  );
}
