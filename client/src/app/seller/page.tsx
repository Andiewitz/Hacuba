"use client";

import { CheckCircle2, ChevronRight, CircleAlert, ImagePlus, Save, Send } from "lucide-react";
import { useState } from "react";
import type { ChangeEvent, FormEvent } from "react";
import Link from "next/link";
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
  const [previewOpen, setPreviewOpen] = useState(false);
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
    setPreviewOpen(true);
  };

  const propertySummary = [draft.propertyType, draft.barangay, draft.city].filter(Boolean).join(" in ") || "Your property";
  const priceCentavos = Number(draft.price.replaceAll(",", "")) * 100;
  const location = [draft.barangay, draft.city].filter(Boolean).join(", ") || "Cebu";

  return (
    <main className="min-h-screen bg-background px-6 py-12 md:px-10 md:py-16 lg:px-20">
      <div className="mx-auto grid max-w-6xl gap-12 lg:grid-cols-[minmax(0,1fr)_320px] lg:items-start">
        <div>
          <div className="flex items-center justify-between gap-4">
            <p className="text-sm font-semibold text-[var(--color-terracotta-ink)]">Seller workspace</p>
            <Link href="/seller/listings" className="text-sm font-semibold text-[var(--color-terracotta-ink)] underline-offset-4 hover:underline">My listings</Link>
          </div>
          <h1 className="mt-3 font-heading text-[2.75rem] font-bold leading-[1.15] tracking-[-0.015em] text-foreground">
            List your property
          </h1>
          <p className="mt-4 max-w-[62ch] text-[1.125rem] leading-[1.6] text-muted-foreground">
            Start with the details buyers need to decide whether your space is right for them. You can save an incomplete draft and return to it later.
          </p>

          {saved && (
            <div role="status" className="mt-8 flex items-start gap-3 rounded-[var(--radius-md)] border border-[var(--color-sage-hover)] bg-[var(--color-sage)] p-4 text-[var(--color-forest)]">
              <CheckCircle2 className="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" />
              <p className="text-sm leading-6">Your draft is saved in this browser. Add photos and complete the required fields before publishing.</p>
            </div>
          )}

          {published && (
            <div role="status" className="mt-8 flex items-start gap-3 rounded-[var(--radius-md)] border border-[var(--color-sage-hover)] bg-[var(--color-sage)] p-4 text-[var(--color-forest)]">
              <CheckCircle2 className="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" />
              <p className="text-sm leading-6">Your listing is published in this workspace. <Link href="/seller/listings" className="font-semibold underline underline-offset-4">View it in My listings</Link>.</p>
            </div>
          )}

          {Object.keys(publishErrors).length > 0 && (
            <div role="alert" className="mt-8 rounded-[var(--radius-md)] border border-primary bg-card p-4 text-foreground">
              <div className="flex items-center gap-2 font-semibold"><CircleAlert className="h-5 w-5 text-[var(--color-terracotta-ink)]" aria-hidden="true" />Complete these fields before publishing</div>
              <ul className="mt-3 list-disc space-y-1 pl-5 text-sm leading-6 text-muted-foreground">
                {Object.values(publishErrors).map((error) => <li key={error}>{error}</li>)}
              </ul>
            </div>
          )}

          {previewOpen && (
            <section className="mt-8 rounded-[var(--radius-lg)] border border-border bg-[var(--color-forest)] p-6 shadow-[var(--shadow-md)]" aria-labelledby="preview-heading">
              <p className="text-sm font-semibold text-[var(--text-secondary-on-dark)]">Buyer-facing card preview</p>
              <h2 id="preview-heading" className="mt-2 font-heading text-[1.75rem] font-semibold leading-[1.25] text-[var(--text-primary-on-dark)]">{draft.title || propertySummary}</h2>
              <div className="mt-6 max-w-sm rounded-[var(--radius-lg)] bg-background p-4">
                <PropertyCard id="seller-preview" href={null} image={photoPreview} propertyType={draft.propertyType || "Property"} location={location} priceCentavos={Number.isFinite(priceCentavos) ? priceCentavos : 0} mode={draft.mode} bedrooms={Number.isInteger(Number(draft.bedrooms)) ? Number(draft.bedrooms) : undefined} bathrooms={Number.isInteger(Number(draft.bathrooms)) ? Number(draft.bathrooms) : undefined} areaSqm={Number.isFinite(Number(draft.area)) && Number(draft.area) > 0 ? Number(draft.area) : undefined} />
              </div>
            </section>
          )}

          <form onSubmit={saveDraft} className="mt-8 rounded-[var(--radius-lg)] border border-border bg-card p-6 shadow-[var(--shadow-md)] md:p-8">
            <fieldset>
              <legend className="font-heading text-[1.75rem] font-semibold leading-[1.25] text-foreground">Property details</legend>
              <div className="mt-6 grid gap-6 md:grid-cols-2">
                <label className="text-sm font-semibold text-foreground">
                  Listing type
                  <select className={inputClassName} value={draft.mode} onChange={(event) => update("mode", event.target.value)}>
                    <option value="for_sale">For sale</option>
                    <option value="for_rent">For rent</option>
                  </select>
                </label>
                <label className="text-sm font-semibold text-foreground">
                  Property type
                  <select className={inputClassName} value={draft.propertyType} onChange={(event) => update("propertyType", event.target.value)}>
                    <option value="" disabled>Select a property type</option>
                    <option>House</option>
                    <option>Apartment</option>
                    <option>Condo</option>
                    <option>Lot</option>
                    <option>Commercial</option>
                  </select>
                </label>
                <label className="text-sm font-semibold text-foreground md:col-span-2">
                  Listing title
                  <input className={inputClassName} value={draft.title} onChange={(event) => update("title", event.target.value)} placeholder="Example: Four-bedroom house near Cebu IT Park" />
                  {publishErrors.title && <span className="mt-2 block text-xs text-[var(--color-terracotta-ink)]">{publishErrors.title}</span>}
                </label>
                <label className="text-sm font-semibold text-foreground">
                  City
                  <input className={inputClassName} value={draft.city} onChange={(event) => update("city", event.target.value)} placeholder="Cebu City" />
                  {publishErrors.city && <span className="mt-2 block text-xs text-[var(--color-terracotta-ink)]">{publishErrors.city}</span>}
                </label>
                <label className="text-sm font-semibold text-foreground">
                  Barangay
                  <input className={inputClassName} value={draft.barangay} onChange={(event) => update("barangay", event.target.value)} placeholder="Lahug" />
                </label>
                <label className="text-sm font-semibold text-foreground">
                  {draft.mode === "for_rent" ? "Monthly rent (PHP)" : "Asking price (PHP)"}
                  <input className={inputClassName} value={draft.price} onChange={(event) => update("price", event.target.value)} inputMode="numeric" placeholder={draft.mode === "for_rent" ? "35,000" : "7,500,000"} />
                  {publishErrors.price && <span className="mt-2 block text-xs text-[var(--color-terracotta-ink)]">{publishErrors.price}</span>}
                </label>
                <label className="text-sm font-semibold text-foreground">
                  Floor or lot area (m²)
                  <input className={inputClassName} value={draft.area} onChange={(event) => update("area", event.target.value)} inputMode="decimal" placeholder="120" />
                  {publishErrors.area && <span className="mt-2 block text-xs text-[var(--color-terracotta-ink)]">{publishErrors.area}</span>}
                </label>
                <label className="text-sm font-semibold text-foreground">
                  Bedrooms
                  <input className={inputClassName} value={draft.bedrooms} onChange={(event) => update("bedrooms", event.target.value)} inputMode="numeric" placeholder="3" />
                  {publishErrors.bedrooms && <span className="mt-2 block text-xs text-[var(--color-terracotta-ink)]">{publishErrors.bedrooms}</span>}
                </label>
                <label className="text-sm font-semibold text-foreground">
                  Bathrooms
                  <input className={inputClassName} value={draft.bathrooms} onChange={(event) => update("bathrooms", event.target.value)} inputMode="numeric" placeholder="2" />
                  {publishErrors.bathrooms && <span className="mt-2 block text-xs text-[var(--color-terracotta-ink)]">{publishErrors.bathrooms}</span>}
                </label>
                <label className="text-sm font-semibold text-foreground md:col-span-2">
                  Description
                  <textarea className={`${inputClassName} min-h-32 resize-y`} value={draft.description} onChange={(event) => update("description", event.target.value)} placeholder="Describe the space, its condition, and the features that matter to buyers." />
                  {publishErrors.description && <span className="mt-2 block text-xs text-[var(--color-terracotta-ink)]">{publishErrors.description}</span>}
                </label>
              </div>
            </fieldset>

            <section className="mt-10 border-t border-border pt-8" aria-labelledby="photos-heading">
              <h2 id="photos-heading" className="font-heading text-[1.375rem] font-medium leading-[1.3] text-foreground">Photos</h2>
              <p className="mt-2 text-sm leading-6 text-muted-foreground">Add clear photos of the exterior, main rooms, and important details.</p>
              <input id="listing-photos" className="sr-only" type="file" accept="image/jpeg,image/png,image/webp" multiple onChange={selectPhotos} />
              <label htmlFor="listing-photos" className="mt-4 inline-flex cursor-pointer items-center gap-2 rounded-[var(--radius-md)] border border-border bg-background px-4 py-3 text-sm font-semibold text-foreground transition-shadow hover:shadow-[var(--shadow-sm)]">
                <ImagePlus className="h-4 w-4" aria-hidden="true" />
                Add photos
              </label>
              {draft.photoNames.length > 0 && <p className="mt-3 text-sm text-muted-foreground">{draft.photoNames.join(", ")}</p>}
              {publishErrors.photos && <p className="mt-3 text-xs text-[var(--color-terracotta-ink)]">{publishErrors.photos}</p>}
            </section>

            <div className="mt-10 flex flex-col-reverse gap-3 border-t border-border pt-6 sm:flex-row sm:justify-end">
              <button type="button" onClick={() => setPreviewOpen((open) => !open)} className="rounded-[var(--radius-md)] border border-border bg-background px-5 py-3 text-sm font-semibold text-foreground transition-shadow hover:shadow-[var(--shadow-sm)]">
                {previewOpen ? "Hide card preview" : "Preview card"}
              </button>
              <button type="submit" className="inline-flex items-center justify-center gap-2 rounded-[var(--radius-md)] bg-primary px-5 py-3 text-sm font-semibold text-[var(--color-forest)] transition-colors hover:bg-[var(--color-terracotta-hover)]">
                <Save className="h-4 w-4" aria-hidden="true" />
                Save draft
              </button>
              <button type="button" onClick={publish} className="inline-flex items-center justify-center gap-2 rounded-[var(--radius-md)] bg-[var(--color-forest)] px-5 py-3 text-sm font-semibold text-[var(--text-primary-on-dark)] transition-colors hover:bg-[var(--color-forest-surface-1)]">
                <Send className="h-4 w-4" aria-hidden="true" />
                Publish listing
              </button>
            </div>
          </form>
        </div>

        <aside className="rounded-[var(--radius-lg)] bg-[var(--color-forest)] p-6 text-[var(--text-body-on-dark)] shadow-[var(--shadow-lg)] lg:sticky lg:top-28">
          <p className="text-sm font-semibold text-[var(--text-secondary-on-dark)]">Before you publish</p>
          <h2 className="mt-3 font-heading text-[1.75rem] font-semibold leading-[1.25] text-[var(--text-primary-on-dark)]">A complete listing earns more attention.</h2>
          <ol className="mt-6 space-y-4 text-sm leading-6">
            <li className="flex gap-3"><span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-[var(--color-sage)] text-xs font-semibold text-[var(--color-forest)]">1</span><span>Set a clear price and property type.</span></li>
            <li className="flex gap-3"><span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-[var(--color-sage)] text-xs font-semibold text-[var(--color-forest)]">2</span><span>Add location details and practical specifications.</span></li>
            <li className="flex gap-3"><span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-[var(--color-sage)] text-xs font-semibold text-[var(--color-forest)]">3</span><span>Upload photos that show the space accurately.</span></li>
          </ol>
          <button type="button" className="mt-8 inline-flex items-center gap-1 text-sm font-semibold text-[var(--text-body-on-dark)] underline-offset-4 hover:underline">
            Read seller guidance
            <ChevronRight className="h-4 w-4" aria-hidden="true" />
          </button>
        </aside>
      </div>
    </main>
  );
}
