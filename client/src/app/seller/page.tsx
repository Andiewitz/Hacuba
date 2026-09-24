"use client";

import { Check, CheckCircle2, CircleAlert, ImagePlus, MapPin, Save, Send, Sparkles } from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import type { ChangeEvent, FormEvent } from "react";
import PropertyCard from "@/components/property-card";

const draftStorageKey = "hacuba:seller-draft";
const inputClassName = "mt-2 w-full rounded-[var(--radius-sm)] border border-border bg-background px-3 py-3 text-sm text-foreground outline-none placeholder:text-muted-foreground focus:border-ring focus:ring-2 focus:ring-ring/40";

type Draft = {
  mode: "for_sale" | "for_rent";
  propertyType: string;
  title: string;
  city: string;
  barangay: string;
  price: string;
  area: string;
  bedrooms: string;
  bathrooms: string;
  description: string;
  sellerName: string;
  contactPhone: string;
  contactEmail: string;
  photoNames: string[];
};

const initialDraft: Draft = {
  mode: "for_sale",
  propertyType: "",
  title: "",
  city: "",
  barangay: "",
  price: "",
  area: "",
  bedrooms: "",
  bathrooms: "",
  description: "",
  sellerName: "",
  contactPhone: "",
  contactEmail: "",
  photoNames: [],
};

function storedDraft() {
  if (typeof window === "undefined") return initialDraft;

  try {
    const savedDraft = window.localStorage.getItem(draftStorageKey);
    return savedDraft ? { ...initialDraft, ...JSON.parse(savedDraft) } : initialDraft;
  } catch {
    window.localStorage.removeItem(draftStorageKey);
    return initialDraft;
  }
}

export default function SellerPage() {
  const [draft, setDraft] = useState<Draft>(storedDraft);
  const [saved, setSaved] = useState(false);
  const [photoPreview, setPhotoPreview] = useState<string>();
  const [publishErrors, setPublishErrors] = useState<Record<string, string>>({});
  const [published, setPublished] = useState(false);

  const update = (field: Exclude<keyof Draft, "photoNames">, value: string) => {
    setDraft((current) => ({ ...current, [field]: value }));
    setSaved(false);
    setPublished(false);
    setPublishErrors({});
  };

  const selectPhotos = (event: ChangeEvent<HTMLInputElement>) => {
    const photos = Array.from(event.target.files ?? []);
    const photoNames = photos.map((file) => file.name);
    setDraft((current) => ({ ...current, photoNames }));
    setPhotoPreview(photos[0] ? URL.createObjectURL(photos[0]) : undefined);
    setSaved(false);
    setPublished(false);
    setPublishErrors({});
  };

  const saveDraft = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    window.localStorage.setItem(draftStorageKey, JSON.stringify(draft));
    setSaved(true);
  };

  const publish = () => {
    const price = Number(draft.price.replaceAll(",", ""));
    const area = Number(draft.area);
    const bedrooms = Number(draft.bedrooms);
    const bathrooms = Number(draft.bathrooms);
    const errors: Record<string, string> = {};

    if (!draft.propertyType) errors.propertyType = "Choose a property type.";
    if (draft.title.trim().length < 5) errors.title = "Add a title with at least 5 characters.";
    if (!draft.city.trim()) errors.city = "Add the city.";
    if (!Number.isFinite(price) || price <= 0) errors.price = "Enter a valid price.";
    if (!draft.description.trim()) errors.description = "Add a property description.";
    if (draft.sellerName.trim().length < 2) errors.sellerName = "Add the seller or agent name.";
    if (!draft.contactPhone.trim() && !draft.contactEmail.trim()) errors.contact = "Add a phone number or email address.";
    if (draft.photoNames.length === 0) errors.photos = "Add at least one property image.";

    if (["House", "Apartment", "Condo"].includes(draft.propertyType)) {
      if (!Number.isInteger(bedrooms) || bedrooms < 0) errors.bedrooms = "Add the number of bedrooms.";
      if (!Number.isInteger(bathrooms) || bathrooms < 0) errors.bathrooms = "Add the number of bathrooms.";
      if (!Number.isFinite(area) || area <= 0) errors.area = "Add the floor area.";
    }
    if (draft.propertyType === "Lot" && (!Number.isFinite(area) || area <= 0)) errors.area = "Add the lot area.";
    if (draft.propertyType === "Commercial" && (!Number.isFinite(area) || area <= 0)) errors.area = "Add the floor area.";

    if (Object.keys(errors).length > 0) {
      setPublishErrors(errors);
      setPublished(false);
      return;
    }

    window.localStorage.setItem("hacuba:published-listing", JSON.stringify(draft));
    setPublishErrors({});
    setPublished(true);
  };

  const priceCentavos = Number(draft.price.replaceAll(",", "")) * 100;
  const location = [draft.barangay, draft.city].filter(Boolean).join(", ") || "Cebu";
  const completion = [
    { label: "Property basics", ready: Boolean(draft.propertyType && draft.title.trim() && draft.city.trim() && draft.price.trim() && draft.description.trim()) },
    { label: "Contact details", ready: Boolean(draft.sellerName.trim() && (draft.contactPhone.trim() || draft.contactEmail.trim())) },
    { label: "Photos", ready: draft.photoNames.length > 0 },
  ];
  const completedSteps = completion.filter((step) => step.ready).length;

  return (
    <main className="min-h-screen bg-[var(--color-forest)] pb-16">
      <section className="border-b border-[var(--color-forest-border-subtle)] px-6 py-10 md:px-10 md:py-12 lg:px-20">
        <div className="mx-auto flex max-w-[1440px] flex-col justify-between gap-8 lg:flex-row lg:items-end">
          <div>
            <div className="flex items-center gap-4 text-sm font-semibold">
              <span className="text-[var(--text-secondary-on-dark)]">Seller workspace</span>
              <Link href="/seller/listings" className="text-[var(--text-body-on-dark)] underline-offset-4 hover:underline">My listings</Link>
            </div>
            <h1 className="mt-4 font-heading text-[2.75rem] font-bold leading-[1.15] tracking-[-0.015em] text-[var(--text-primary-on-dark)] md:text-[3.5rem] md:leading-[1.1] md:tracking-[-0.02em]">Make a listing buyers can act on.</h1>
            <p className="mt-4 max-w-[62ch] text-[1.125rem] leading-[1.6] text-[var(--text-body-on-dark)]">Give buyers the property facts, photos, and direct contact details they need to reach you.</p>
          </div>
          <div className="rounded-[var(--radius-lg)] border border-[var(--color-forest-border-strong)] bg-[var(--color-forest-surface-1)] p-4 lg:min-w-[292px]">
            <div className="flex items-center justify-between gap-4">
              <span className="text-sm font-semibold text-[var(--text-body-on-dark)]">Listing readiness</span>
              <span className="rounded-full bg-[var(--color-sage)] px-3 py-1 text-xs font-semibold text-[var(--color-forest)]">{completedSteps} of 3</span>
            </div>
            <div className="mt-4 flex gap-2" aria-label={`${completedSteps} of 3 listing sections complete`}>
              {completion.map((step) => <span key={step.label} className={`h-2 flex-1 rounded-full ${step.ready ? "bg-[var(--color-sage)]" : "bg-[var(--color-forest-surface-2)]"}`} />)}
            </div>
          </div>
        </div>
      </section>

      <div className="mx-auto grid max-w-[1440px] gap-8 px-6 py-8 md:px-10 md:py-12 lg:grid-cols-[minmax(0,1fr)_360px] lg:px-20">
        <div>
          {saved && (
            <div role="status" className="mb-6 flex items-start gap-3 rounded-[var(--radius-md)] border border-[var(--color-sage-hover)] bg-[var(--color-sage)] p-4 text-[var(--color-forest)]">
              <CheckCircle2 className="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" />
              <p className="text-sm leading-6">Your draft is saved in this browser.</p>
            </div>
          )}
          {published && (
            <div role="status" className="mb-6 flex items-start gap-3 rounded-[var(--radius-md)] border border-[var(--color-sage-hover)] bg-[var(--color-sage)] p-4 text-[var(--color-forest)]">
              <CheckCircle2 className="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" />
              <p className="text-sm leading-6">Your listing is published in this workspace. <Link href="/seller/listings" className="font-semibold underline underline-offset-4">View it in My listings</Link>.</p>
            </div>
          )}
          {Object.keys(publishErrors).length > 0 && (
            <div role="alert" className="mb-6 rounded-[var(--radius-md)] border border-[var(--color-terracotta)] bg-card p-4 text-foreground">
              <div className="flex items-center gap-2 text-sm font-semibold"><CircleAlert className="h-5 w-5 text-[var(--color-terracotta-ink)]" aria-hidden="true" />Finish these items before publishing</div>
              <ul className="mt-3 list-disc space-y-1 pl-5 text-sm leading-6 text-muted-foreground">
                {Object.values(publishErrors).map((error) => <li key={error}>{error}</li>)}
              </ul>
            </div>
          )}

          <form onSubmit={saveDraft} className="space-y-6">
            <section className="rounded-[var(--radius-lg)] bg-card p-6 shadow-[var(--shadow-md)] md:p-8" aria-labelledby="basics-heading">
              <div className="flex items-start justify-between gap-4">
                <div>
                  <p className="text-sm font-semibold text-[var(--color-terracotta-ink)]">01 — Property</p>
                  <h2 id="basics-heading" className="mt-2 font-heading text-[1.75rem] font-semibold leading-[1.25] tracking-[-0.005em] text-foreground">The essentials</h2>
                </div>
                <MapPin className="h-5 w-5 text-[var(--color-terracotta-ink)]" aria-hidden="true" />
              </div>
              <div className="mt-6 grid gap-6 md:grid-cols-2">
                <fieldset>
                  <legend className="text-sm font-semibold text-foreground">Listing type</legend>
                  <div className="mt-2 grid grid-cols-2 gap-2">
                    {(["for_sale", "for_rent"] as const).map((mode) => <button key={mode} type="button" onClick={() => update("mode", mode)} className={`rounded-[var(--radius-md)] border px-3 py-3 text-sm font-semibold transition-colors ${draft.mode === mode ? "border-[var(--color-forest)] bg-[var(--color-forest)] text-[var(--text-primary-on-dark)]" : "border-border bg-background text-foreground hover:border-[var(--color-forest-border-strong)]"}`}>{mode === "for_sale" ? "For sale" : "For rent"}</button>)}
                  </div>
                </fieldset>
                <label className="text-sm font-semibold text-foreground">
                  Property type
                  <select className={inputClassName} value={draft.propertyType} onChange={(event) => update("propertyType", event.target.value)}>
                    <option value="" disabled>Select a property type</option>
                    <option>House</option><option>Apartment</option><option>Condo</option><option>Lot</option><option>Commercial</option>
                  </select>
                  {publishErrors.propertyType && <span className="mt-2 block text-xs text-[var(--color-terracotta-ink)]">{publishErrors.propertyType}</span>}
                </label>
                <label className="text-sm font-semibold text-foreground md:col-span-2">
                  Listing title
                  <input className={inputClassName} value={draft.title} onChange={(event) => update("title", event.target.value)} placeholder="Four-bedroom house near Cebu IT Park" />
                  {publishErrors.title && <span className="mt-2 block text-xs text-[var(--color-terracotta-ink)]">{publishErrors.title}</span>}
                </label>
                <label className="text-sm font-semibold text-foreground"><span>City</span><input className={inputClassName} value={draft.city} onChange={(event) => update("city", event.target.value)} placeholder="Cebu City" />{publishErrors.city && <span className="mt-2 block text-xs text-[var(--color-terracotta-ink)]">{publishErrors.city}</span>}</label>
                <label className="text-sm font-semibold text-foreground"><span>Barangay</span><input className={inputClassName} value={draft.barangay} onChange={(event) => update("barangay", event.target.value)} placeholder="Lahug" /></label>
              </div>
            </section>

            <section className="rounded-[var(--radius-lg)] bg-card p-6 shadow-[var(--shadow-md)] md:p-8" aria-labelledby="details-heading">
              <p className="text-sm font-semibold text-[var(--color-terracotta-ink)]">02 — Details</p>
              <h2 id="details-heading" className="mt-2 font-heading text-[1.75rem] font-semibold leading-[1.25] tracking-[-0.005em] text-foreground">Price and specifications</h2>
              <div className="mt-6 grid gap-6 md:grid-cols-2">
                <label className="text-sm font-semibold text-foreground">{draft.mode === "for_rent" ? "Monthly rent (PHP)" : "Asking price (PHP)"}<input className={inputClassName} value={draft.price} onChange={(event) => update("price", event.target.value)} inputMode="numeric" placeholder={draft.mode === "for_rent" ? "35,000" : "7,500,000"} />{publishErrors.price && <span className="mt-2 block text-xs text-[var(--color-terracotta-ink)]">{publishErrors.price}</span>}</label>
                <label className="text-sm font-semibold text-foreground">Floor or lot area (m²)<input className={inputClassName} value={draft.area} onChange={(event) => update("area", event.target.value)} inputMode="decimal" placeholder="120" />{publishErrors.area && <span className="mt-2 block text-xs text-[var(--color-terracotta-ink)]">{publishErrors.area}</span>}</label>
                <label className="text-sm font-semibold text-foreground">Bedrooms<input className={inputClassName} value={draft.bedrooms} onChange={(event) => update("bedrooms", event.target.value)} inputMode="numeric" placeholder="3" />{publishErrors.bedrooms && <span className="mt-2 block text-xs text-[var(--color-terracotta-ink)]">{publishErrors.bedrooms}</span>}</label>
                <label className="text-sm font-semibold text-foreground">Bathrooms<input className={inputClassName} value={draft.bathrooms} onChange={(event) => update("bathrooms", event.target.value)} inputMode="numeric" placeholder="2" />{publishErrors.bathrooms && <span className="mt-2 block text-xs text-[var(--color-terracotta-ink)]">{publishErrors.bathrooms}</span>}</label>
                <label className="text-sm font-semibold text-foreground md:col-span-2">Description<textarea className={`${inputClassName} min-h-36 resize-y`} value={draft.description} onChange={(event) => update("description", event.target.value)} placeholder="Describe the space, its condition, and the features that matter to buyers." />{publishErrors.description && <span className="mt-2 block text-xs text-[var(--color-terracotta-ink)]">{publishErrors.description}</span>}</label>
              </div>
            </section>

            <section className="rounded-[var(--radius-lg)] bg-card p-6 shadow-[var(--shadow-md)] md:p-8" aria-labelledby="contact-heading">
              <p className="text-sm font-semibold text-[var(--color-terracotta-ink)]">03 — Contact</p>
              <h2 id="contact-heading" className="mt-2 font-heading text-[1.75rem] font-semibold leading-[1.25] tracking-[-0.005em] text-foreground">How buyers reach you</h2>
              <p className="mt-3 max-w-[64ch] text-sm leading-6 text-muted-foreground">Hacuba helps buyers discover your listing. You arrange viewings, terms, and payment directly with them.</p>
              <div className="mt-6 grid gap-6 md:grid-cols-2">
                <label className="text-sm font-semibold text-foreground md:col-span-2">Seller or agent name<input className={inputClassName} value={draft.sellerName} onChange={(event) => update("sellerName", event.target.value)} placeholder="Your name or agency" />{publishErrors.sellerName && <span className="mt-2 block text-xs text-[var(--color-terracotta-ink)]">{publishErrors.sellerName}</span>}</label>
                <label className="text-sm font-semibold text-foreground">Phone number<input className={inputClassName} value={draft.contactPhone} onChange={(event) => update("contactPhone", event.target.value)} inputMode="tel" placeholder="+63 917 123 4567" /></label>
                <label className="text-sm font-semibold text-foreground">Email address<input className={inputClassName} value={draft.contactEmail} onChange={(event) => update("contactEmail", event.target.value)} inputMode="email" placeholder="you@example.com" /></label>
              </div>
              {publishErrors.contact && <p className="mt-3 text-xs text-[var(--color-terracotta-ink)]">{publishErrors.contact}</p>}
            </section>

            <section className="rounded-[var(--radius-lg)] bg-card p-6 shadow-[var(--shadow-md)] md:p-8" aria-labelledby="photos-heading">
              <p className="text-sm font-semibold text-[var(--color-terracotta-ink)]">04 — Photos</p>
              <h2 id="photos-heading" className="mt-2 font-heading text-[1.75rem] font-semibold leading-[1.25] tracking-[-0.005em] text-foreground">Show the property clearly</h2>
              <p className="mt-3 text-sm leading-6 text-muted-foreground">Use clear photos of the exterior, main rooms, and the details that help buyers decide.</p>
              <input id="listing-photos" className="sr-only" type="file" accept="image/jpeg,image/png,image/webp" multiple onChange={selectPhotos} />
              <label htmlFor="listing-photos" className="mt-6 inline-flex cursor-pointer items-center gap-2 rounded-[var(--radius-md)] border border-border bg-background px-4 py-3 text-sm font-semibold text-foreground transition-shadow hover:shadow-[var(--shadow-sm)]"><ImagePlus className="h-4 w-4" aria-hidden="true" />Add photos</label>
              {draft.photoNames.length > 0 && <p className="mt-3 text-sm text-muted-foreground">{draft.photoNames.length} photo{draft.photoNames.length === 1 ? "" : "s"} selected</p>}
              {publishErrors.photos && <p className="mt-3 text-xs text-[var(--color-terracotta-ink)]">{publishErrors.photos}</p>}
            </section>

            <div className="flex flex-col-reverse gap-3 rounded-[var(--radius-lg)] bg-[var(--color-forest-surface-1)] p-4 sm:flex-row sm:items-center sm:justify-between">
              <p className="text-sm text-[var(--text-body-on-dark)]">You can save an incomplete draft and return to it later.</p>
              <div className="flex flex-col gap-3 sm:flex-row">
                <button type="submit" className="inline-flex items-center justify-center gap-2 rounded-[var(--radius-md)] border border-[var(--color-forest-border-strong)] px-5 py-3 text-sm font-semibold text-[var(--text-body-on-dark)] transition-colors hover:bg-[var(--color-forest-surface-2)]"><Save className="h-4 w-4" aria-hidden="true" />Save draft</button>
                <button type="button" onClick={publish} className="inline-flex items-center justify-center gap-2 rounded-[var(--radius-md)] bg-[var(--color-terracotta)] px-5 py-3 text-sm font-semibold text-black transition-colors hover:bg-[var(--color-terracotta-hover)]"><Send className="h-4 w-4" aria-hidden="true" />Publish listing</button>
              </div>
            </div>
          </form>
        </div>

        <aside className="space-y-6 lg:sticky lg:top-28 lg:h-fit">
          <section className="rounded-[var(--radius-lg)] bg-[var(--color-forest-surface-1)] p-6 shadow-[var(--shadow-lg)]" aria-labelledby="preview-heading">
            <div className="flex items-center gap-2 text-sm font-semibold text-[var(--text-secondary-on-dark)]"><Sparkles className="h-4 w-4" aria-hidden="true" />Buyer-facing preview</div>
            <h2 id="preview-heading" className="mt-3 font-heading text-[1.75rem] font-semibold leading-[1.25] tracking-[-0.005em] text-[var(--text-primary-on-dark)]">Your card, as buyers see it</h2>
            <div className="mt-6 rounded-[var(--radius-lg)] bg-background p-4 shadow-[var(--shadow-sm)]">
              <PropertyCard id="seller-preview" href={null} image={photoPreview} propertyType={draft.propertyType || "Property"} location={location} priceCentavos={Number.isFinite(priceCentavos) ? priceCentavos : 0} mode={draft.mode} bedrooms={Number.isInteger(Number(draft.bedrooms)) ? Number(draft.bedrooms) : undefined} bathrooms={Number.isInteger(Number(draft.bathrooms)) ? Number(draft.bathrooms) : undefined} areaSqm={Number.isFinite(Number(draft.area)) && Number(draft.area) > 0 ? Number(draft.area) : undefined} />
            </div>
          </section>

          <section className="rounded-[var(--radius-lg)] border border-[var(--color-forest-border-subtle)] bg-[var(--color-forest-surface-1)] p-6" aria-labelledby="checklist-heading">
            <p className="text-sm font-semibold text-[var(--text-secondary-on-dark)]">Before you publish</p>
            <h2 id="checklist-heading" className="mt-2 font-heading text-[1.375rem] font-medium leading-[1.3] text-[var(--text-primary-on-dark)]">A complete listing earns attention.</h2>
            <ul className="mt-6 space-y-4">
              {completion.map((step) => <li key={step.label} className="flex items-center gap-3 text-sm text-[var(--text-body-on-dark)]"><span className={`flex h-6 w-6 shrink-0 items-center justify-center rounded-full ${step.ready ? "bg-[var(--color-sage)] text-[var(--color-forest)]" : "border border-[var(--color-forest-border-strong)] text-[var(--text-secondary-on-dark)]"}`}>{step.ready && <Check className="h-4 w-4" aria-hidden="true" />}</span>{step.label}</li>)}
            </ul>
          </section>
        </aside>
      </div>
    </main>
  );
}
