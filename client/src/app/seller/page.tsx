"use client";

import { Check, CheckCircle2, CircleAlert, ImagePlus, MapPin, Save, Send, Sparkles } from "lucide-react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";
import type { ChangeEvent, FormEvent } from "react";
import PropertyCard from "@/components/property-card";
import { useAppSelector } from "@/lib/hooks";
import {
  apiListingToDraft,
  draftPayload,
  emptySellerDraft,
  sellerRequest,
  type OwnedListing,
  type SellerAPIError,
  type SellerDraft,
} from "@/lib/seller-listings";

const inputClassName = "mt-2 w-full rounded-[var(--radius-sm)] border border-border bg-background px-3 py-3 text-sm text-foreground outline-none placeholder:text-muted-foreground focus:border-ring focus:ring-2 focus:ring-ring/40";

function errorFields(error: unknown) {
  return (error as SellerAPIError).fields ?? {};
}

export default function SellerPage() {
  const { user, accessToken, csrfToken, initialized } = useAppSelector((state) => state.auth);
  const searchParams = useSearchParams();
  const [draft, setDraft] = useState<SellerDraft>(emptySellerDraft);
  const [listingID, setListingID] = useState<string>();
  const [saved, setSaved] = useState(false);
  const [published, setPublished] = useState<OwnedListing>();
  const [photoPreview, setPhotoPreview] = useState<string>();
  const [photoNames, setPhotoNames] = useState<string[]>([]);
  const [formError, setFormError] = useState<string>();
  const [publishErrors, setPublishErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);

  const update = (field: keyof SellerDraft, value: string) => {
    setDraft((current) => ({ ...current, [field]: value }));
    setSaved(false);
    setPublished(undefined);
    setFormError(undefined);
    setPublishErrors({});
  };

  const selectPhotos = (event: ChangeEvent<HTMLInputElement>) => {
    const photos = Array.from(event.target.files ?? []);
    setPhotoNames(photos.map((file) => file.name));
    setPhotoPreview(photos[0] ? URL.createObjectURL(photos[0]) : undefined);
    setPublished(undefined);
  };

  const sessionReady = Boolean(user?.role === "seller" && accessToken && csrfToken);

  useEffect(() => {
    const requestedID = searchParams.get("listing");
    if (!requestedID || !sessionReady || !accessToken || !csrfToken) return;
    let active = true;
    sellerRequest<{ listings: OwnedListing[] }>("/api/listings/mine", accessToken, csrfToken)
      .then((response) => {
        const listing = response.listings.find((item) => item.id === requestedID);
        if (!active) return;
        if (!listing) {
          setFormError("That listing is no longer available in your seller account.");
          return;
        }
        setListingID(listing.id);
        setDraft(apiListingToDraft(listing));
      })
      .catch(() => {
        if (active) setFormError("Couldn't load that seller listing.");
      });
    return () => { active = false; };
  }, [accessToken, csrfToken, searchParams, sessionReady]);

  const saveDraft = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!sessionReady || !accessToken || !csrfToken) {
      setFormError("Log in with a seller account before saving a listing.");
      return;
    }
    setSubmitting(true);
    setFormError(undefined);
    setPublishErrors({});
    try {
      const response = await sellerRequest<OwnedListing>(
        listingID ? `/api/listings/${listingID}` : "/api/listings",
        accessToken,
        csrfToken,
        { method: listingID ? "PATCH" : "POST", body: JSON.stringify(draftPayload(draft)) },
      );
      setListingID(response.id);
      setDraft(apiListingToDraft(response));
      setSaved(true);
    } catch (error) {
      setPublishErrors(errorFields(error));
      setFormError(error instanceof Error ? error.message : "Couldn't save the draft.");
    } finally {
      setSubmitting(false);
    }
  };

  const publish = async () => {
    if (!listingID || !accessToken || !csrfToken) {
      setFormError("Save this draft to Hacuba before publishing it.");
      return;
    }
    setSubmitting(true);
    setFormError(undefined);
    setPublishErrors({});
    try {
      const response = await sellerRequest<OwnedListing>(
        `/api/listings/${listingID}/publish`,
        accessToken,
        csrfToken,
        { method: "POST" },
      );
      setPublished(response);
      setSaved(false);
    } catch (error) {
      setPublishErrors(errorFields(error));
      setFormError(error instanceof Error ? error.message : "Couldn't publish the listing.");
    } finally {
      setSubmitting(false);
    }
  };

  const priceCentavos = Number(draft.price.replaceAll(",", "")) * 100;
  const location = [draft.barangay, draft.city].filter(Boolean).join(", ") || "Location";
  const completion = [
    { label: "Property basics", ready: Boolean(draft.propertyType && draft.title.trim() && draft.city.trim() && draft.price.trim() && draft.description.trim()) },
    { label: "Contact details", ready: Boolean(draft.sellerName.trim() && (draft.contactPhone.trim() || draft.contactEmail.trim())) },
    { label: "Uploaded photos", ready: false },
  ];
  const completedSteps = completion.filter((step) => step.ready).length;

  return (
    <main className="min-h-screen bg-[var(--color-forest)] pb-16">
      <section className="border-b border-[var(--color-forest-border-subtle)] px-6 py-10 md:px-10 md:py-12 lg:px-20">
        <div className="mx-auto flex max-w-[1440px] flex-col justify-between gap-8 lg:flex-row lg:items-end">
          <div>
            <div className="flex items-center gap-4 text-sm font-semibold"><span className="text-[var(--text-secondary-on-dark)]">Seller workspace</span><Link href="/seller/listings" className="text-[var(--text-body-on-dark)] underline-offset-4 hover:underline">My listings</Link></div>
            <h1 className="mt-4 font-heading text-[2.75rem] font-bold leading-[1.15] tracking-[-0.015em] text-[var(--text-primary-on-dark)] md:text-[3.5rem] md:leading-[1.1] md:tracking-[-0.02em]">Make a listing buyers can act on.</h1>
            <p className="mt-4 max-w-[62ch] text-[1.125rem] leading-[1.6] text-[var(--text-body-on-dark)]">Save property details to your seller account, then publish only when the listing meets Hacuba&apos;s requirements.</p>
          </div>
          <div className="rounded-[var(--radius-lg)] border border-[var(--color-forest-border-strong)] bg-[var(--color-forest-surface-1)] p-4 lg:min-w-[292px]"><div className="flex items-center justify-between gap-4"><span className="text-sm font-semibold text-[var(--text-body-on-dark)]">Listing readiness</span><span className="rounded-full bg-[var(--color-sage)] px-3 py-1 text-xs font-semibold text-[var(--color-forest)]">{completedSteps} of 3</span></div><div className="mt-4 flex gap-2" aria-label={`${completedSteps} of 3 listing sections complete`}>{completion.map((step) => <span key={step.label} className={`h-2 flex-1 rounded-full ${step.ready ? "bg-[var(--color-sage)]" : "bg-[var(--color-forest-surface-2)]"}`} />)}</div></div>
        </div>
      </section>

      <div className="mx-auto grid max-w-[1440px] gap-8 px-6 py-8 md:px-10 md:py-12 lg:grid-cols-[minmax(0,1fr)_360px] lg:px-20">
        <div>
          {!initialized && <p className="mb-6 text-sm text-[var(--text-body-on-dark)]">Checking your account…</p>}
          {initialized && !sessionReady && <div role="alert" className="mb-6 rounded-[var(--radius-md)] border border-[var(--color-forest-border-strong)] bg-[var(--color-forest-surface-1)] p-4 text-sm leading-6 text-[var(--text-body-on-dark)]">Log in and activate seller access from the navigation before saving a listing.</div>}
          {saved && <div role="status" className="mb-6 flex items-start gap-3 rounded-[var(--radius-md)] border border-[var(--color-sage-hover)] bg-[var(--color-sage)] p-4 text-[var(--color-forest)]"><CheckCircle2 className="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" /><p className="text-sm leading-6">Your draft is saved to your Hacuba seller account.</p></div>}
          {published && <div role="status" className="mb-6 flex items-start gap-3 rounded-[var(--radius-md)] border border-[var(--color-sage-hover)] bg-[var(--color-sage)] p-4 text-[var(--color-forest)]"><CheckCircle2 className="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" /><p className="text-sm leading-6">Your listing is published. <Link href={`/listings/${published.id}`} className="font-semibold underline underline-offset-4">Open the public listing</Link>.</p></div>}
          {(formError || Object.keys(publishErrors).length > 0) && <div role="alert" className="mb-6 rounded-[var(--radius-md)] border border-[var(--color-terracotta)] bg-card p-4 text-foreground"><div className="flex items-center gap-2 text-sm font-semibold"><CircleAlert className="h-5 w-5 text-[var(--color-terracotta-ink)]" aria-hidden="true" />{formError ?? "Finish these items before publishing"}</div>{Object.keys(publishErrors).length > 0 && <ul className="mt-3 list-disc space-y-1 pl-5 text-sm leading-6 text-muted-foreground">{Object.entries(publishErrors).map(([field, message]) => <li key={field}>{message}</li>)}</ul>}</div>}

          <form onSubmit={saveDraft} className="space-y-6">
            <section className="rounded-[var(--radius-lg)] bg-card p-6 shadow-[var(--shadow-md)] md:p-8"><div className="flex items-start justify-between gap-4"><div><p className="text-sm font-semibold text-[var(--color-terracotta-ink)]">01 — Property</p><h2 className="mt-2 font-heading text-[1.75rem] font-semibold leading-[1.25] tracking-[-0.005em] text-foreground">The essentials</h2></div><MapPin className="h-5 w-5 text-[var(--color-terracotta-ink)]" aria-hidden="true" /></div><div className="mt-6 grid gap-6 md:grid-cols-2">
              <fieldset><legend className="text-sm font-semibold text-foreground">Listing type</legend><div className="mt-2 grid grid-cols-2 gap-2">{(["for_sale", "for_rent"] as const).map((mode) => <button key={mode} type="button" onClick={() => update("mode", mode)} className={`rounded-[var(--radius-md)] border px-3 py-3 text-sm font-semibold transition-colors ${draft.mode === mode ? "border-[var(--color-forest)] bg-[var(--color-forest)] text-[var(--text-primary-on-dark)]" : "border-border bg-background text-foreground hover:border-[var(--color-forest-border-strong)]"}`}>{mode === "for_sale" ? "For sale" : "For rent"}</button>)}</div></fieldset>
              <label className="text-sm font-semibold text-foreground">Property type<select className={inputClassName} value={draft.propertyType} onChange={(event) => update("propertyType", event.target.value)}><option value="" disabled>Select a property type</option><option>House</option><option>Apartment</option><option>Condo</option><option>Lot</option><option>Commercial</option></select></label>
              <label className="text-sm font-semibold text-foreground md:col-span-2">Listing title<input className={inputClassName} value={draft.title} onChange={(event) => update("title", event.target.value)} placeholder="Four-bedroom house near Cebu IT Park" /></label>
              <label className="text-sm font-semibold text-foreground">City<input className={inputClassName} value={draft.city} onChange={(event) => update("city", event.target.value)} placeholder="Cebu City" /></label><label className="text-sm font-semibold text-foreground">Barangay<input className={inputClassName} value={draft.barangay} onChange={(event) => update("barangay", event.target.value)} placeholder="Lahug" /></label>
            </div></section>
            <section className="rounded-[var(--radius-lg)] bg-card p-6 shadow-[var(--shadow-md)] md:p-8"><p className="text-sm font-semibold text-[var(--color-terracotta-ink)]">02 — Details</p><h2 className="mt-2 font-heading text-[1.75rem] font-semibold leading-[1.25] tracking-[-0.005em] text-foreground">Price and specifications</h2><div className="mt-6 grid gap-6 md:grid-cols-2"><label className="text-sm font-semibold text-foreground">{draft.mode === "for_rent" ? "Monthly rent (PHP)" : "Asking price (PHP)"}<input className={inputClassName} value={draft.price} onChange={(event) => update("price", event.target.value)} inputMode="numeric" placeholder={draft.mode === "for_rent" ? "35,000" : "7,500,000"} /></label><label className="text-sm font-semibold text-foreground">Floor or lot area (m²)<input className={inputClassName} value={draft.area} onChange={(event) => update("area", event.target.value)} inputMode="decimal" placeholder="120" /></label><label className="text-sm font-semibold text-foreground">Bedrooms<input className={inputClassName} value={draft.bedrooms} onChange={(event) => update("bedrooms", event.target.value)} inputMode="numeric" placeholder="3" /></label><label className="text-sm font-semibold text-foreground">Bathrooms<input className={inputClassName} value={draft.bathrooms} onChange={(event) => update("bathrooms", event.target.value)} inputMode="numeric" placeholder="2" /></label><label className="text-sm font-semibold text-foreground md:col-span-2">Description<textarea className={`${inputClassName} min-h-36 resize-y`} value={draft.description} onChange={(event) => update("description", event.target.value)} placeholder="Describe the space, its condition, and the features that matter to buyers." /></label></div></section>
            <section className="rounded-[var(--radius-lg)] bg-card p-6 shadow-[var(--shadow-md)] md:p-8"><p className="text-sm font-semibold text-[var(--color-terracotta-ink)]">03 — Contact</p><h2 className="mt-2 font-heading text-[1.75rem] font-semibold leading-[1.25] tracking-[-0.005em] text-foreground">How buyers reach you</h2><p className="mt-3 max-w-[64ch] text-sm leading-6 text-muted-foreground">Hacuba helps buyers discover your listing. You arrange viewings, terms, and payment directly with them.</p><div className="mt-6 grid gap-6 md:grid-cols-2"><label className="text-sm font-semibold text-foreground md:col-span-2">Seller or agent name<input className={inputClassName} value={draft.sellerName} onChange={(event) => update("sellerName", event.target.value)} placeholder="Your name or agency" /></label><label className="text-sm font-semibold text-foreground">Phone number<input className={inputClassName} value={draft.contactPhone} onChange={(event) => update("contactPhone", event.target.value)} inputMode="tel" placeholder="+63 917 123 4567" /></label><label className="text-sm font-semibold text-foreground">Email address<input className={inputClassName} value={draft.contactEmail} onChange={(event) => update("contactEmail", event.target.value)} inputMode="email" placeholder="you@example.com" /></label></div></section>
            <section className="rounded-[var(--radius-lg)] bg-card p-6 shadow-[var(--shadow-md)] md:p-8"><p className="text-sm font-semibold text-[var(--color-terracotta-ink)]">04 — Photos</p><h2 className="mt-2 font-heading text-[1.75rem] font-semibold leading-[1.25] tracking-[-0.005em] text-foreground">Preview your property photos</h2><p className="mt-3 text-sm leading-6 text-muted-foreground">Selected files improve this card preview only. Secure upload and registration are the next seller workflow step; Hacuba will not publish them before then.</p><input id="listing-photos" className="sr-only" type="file" accept="image/jpeg,image/png,image/webp" multiple onChange={selectPhotos} /><label htmlFor="listing-photos" className="mt-6 inline-flex cursor-pointer items-center gap-2 rounded-[var(--radius-md)] border border-border bg-background px-4 py-3 text-sm font-semibold text-foreground transition-shadow hover:shadow-[var(--shadow-sm)]"><ImagePlus className="h-4 w-4" aria-hidden="true" />Select photos for preview</label>{photoNames.length > 0 && <p className="mt-3 text-sm text-muted-foreground">{photoNames.length} photo{photoNames.length === 1 ? "" : "s"} selected for this browser preview.</p>}</section>
            <div className="flex flex-col-reverse gap-3 rounded-[var(--radius-lg)] bg-[var(--color-forest-surface-1)] p-4 sm:flex-row sm:items-center sm:justify-between"><p className="text-sm text-[var(--text-body-on-dark)]">Draft data is saved to your seller account. Publishing waits for server validation.</p><div className="flex flex-col gap-3 sm:flex-row"><button type="submit" disabled={submitting} className="inline-flex items-center justify-center gap-2 rounded-[var(--radius-md)] border border-[var(--color-forest-border-strong)] px-5 py-3 text-sm font-semibold text-[var(--text-body-on-dark)] transition-colors hover:bg-[var(--color-forest-surface-2)] disabled:opacity-50"><Save className="h-4 w-4" aria-hidden="true" />{submitting ? "Saving…" : "Save draft"}</button><button type="button" disabled={submitting || !listingID} onClick={publish} className="inline-flex items-center justify-center gap-2 rounded-[var(--radius-md)] bg-[var(--color-terracotta)] px-5 py-3 text-sm font-semibold text-black transition-colors hover:bg-[var(--color-terracotta-hover)] disabled:opacity-50"><Send className="h-4 w-4" aria-hidden="true" />Publish listing</button></div></div>
          </form>
        </div>
        <aside className="space-y-6 lg:sticky lg:top-28 lg:h-fit"><section className="rounded-[var(--radius-lg)] bg-[var(--color-forest-surface-1)] p-6 shadow-[var(--shadow-lg)]"><div className="flex items-center gap-2 text-sm font-semibold text-[var(--text-secondary-on-dark)]"><Sparkles className="h-4 w-4" aria-hidden="true" />Buyer-facing card preview</div><h2 className="mt-3 font-heading text-[1.75rem] font-semibold leading-[1.25] tracking-[-0.005em] text-[var(--text-primary-on-dark)]">Your card, as buyers see it</h2><div className="mt-6 rounded-[var(--radius-lg)] bg-background p-4 shadow-[var(--shadow-sm)]"><PropertyCard id="seller-preview" href={null} image={photoPreview} propertyType={draft.propertyType || "Property"} location={location} priceCentavos={Number.isFinite(priceCentavos) ? priceCentavos : 0} mode={draft.mode} bedrooms={Number.isInteger(Number(draft.bedrooms)) ? Number(draft.bedrooms) : undefined} bathrooms={Number.isInteger(Number(draft.bathrooms)) ? Number(draft.bathrooms) : undefined} areaSqm={Number.isFinite(Number(draft.area)) && Number(draft.area) > 0 ? Number(draft.area) : undefined} /></div></section><section className="rounded-[var(--radius-lg)] border border-[var(--color-forest-border-subtle)] bg-[var(--color-forest-surface-1)] p-6"><p className="text-sm font-semibold text-[var(--text-secondary-on-dark)]">Before you publish</p><h2 className="mt-2 font-heading text-[1.375rem] font-medium leading-[1.3] text-[var(--text-primary-on-dark)]">A complete listing earns attention.</h2><ul className="mt-6 space-y-4">{completion.map((step) => <li key={step.label} className="flex items-center gap-3 text-sm text-[var(--text-body-on-dark)]"><span className={`flex h-6 w-6 shrink-0 items-center justify-center rounded-full ${step.ready ? "bg-[var(--color-sage)] text-[var(--color-forest)]" : "border border-[var(--color-forest-border-strong)] text-[var(--text-secondary-on-dark)]"}`}>{step.ready && <Check className="h-4 w-4" aria-hidden="true" />}</span>{step.label}</li>)}</ul></section></aside>
      </div>
    </main>
  );
}
