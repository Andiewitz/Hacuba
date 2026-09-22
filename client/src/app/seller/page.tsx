"use client";

import { CheckCircle2, ChevronRight, ImagePlus, Save } from "lucide-react";
import { useState } from "react";
import type { FormEvent } from "react";

const inputClassName = "mt-2 w-full rounded-[var(--radius-sm)] border border-border bg-background px-3 py-3 text-sm text-foreground outline-none placeholder:text-muted-foreground focus:border-ring focus:ring-2 focus:ring-ring/40";

export default function SellerPage() {
  const [saved, setSaved] = useState(false);
  const [mode, setMode] = useState("for_sale");

  const saveDraft = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setSaved(true);
  };

  return (
    <main className="min-h-screen bg-background px-6 py-12 md:px-10 md:py-16 lg:px-20">
      <div className="mx-auto grid max-w-6xl gap-12 lg:grid-cols-[minmax(0,1fr)_320px] lg:items-start">
        <div>
          <p className="text-sm font-semibold text-[var(--color-terracotta-ink)]">Seller workspace</p>
          <h1 className="mt-3 font-heading text-[2.75rem] font-bold leading-[1.15] tracking-[-0.015em] text-foreground">
            List your property
          </h1>
          <p className="mt-4 max-w-[62ch] text-[1.125rem] leading-[1.6] text-muted-foreground">
            Start with the details buyers need to decide whether your space is right for them. You can save an incomplete draft and return to it later.
          </p>

          {saved && (
            <div className="mt-8 flex items-start gap-3 rounded-[var(--radius-md)] border border-[var(--color-sage-hover)] bg-[var(--color-sage)] p-4 text-[var(--color-forest)]">
              <CheckCircle2 className="mt-0.5 h-5 w-5 shrink-0" aria-hidden="true" />
              <p className="text-sm leading-6">Your draft details are ready to review. Add photos and complete the required fields before publishing.</p>
            </div>
          )}

          <form onSubmit={saveDraft} className="mt-8 rounded-[var(--radius-lg)] border border-border bg-card p-6 shadow-[var(--shadow-md)] md:p-8">
            <fieldset>
              <legend className="font-heading text-[1.75rem] font-semibold leading-[1.25] text-foreground">Property details</legend>
              <div className="mt-6 grid gap-6 md:grid-cols-2">
                <label className="text-sm font-semibold text-foreground">
                  Listing type
                  <select className={inputClassName} value={mode} onChange={(event) => setMode(event.target.value)}>
                    <option value="for_sale">For sale</option>
                    <option value="for_rent">For rent</option>
                  </select>
                </label>
                <label className="text-sm font-semibold text-foreground">
                  Property type
                  <select className={inputClassName} defaultValue="">
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
                  <input className={inputClassName} placeholder="Example: Four-bedroom house near Cebu IT Park" />
                </label>
                <label className="text-sm font-semibold text-foreground">
                  City
                  <input className={inputClassName} placeholder="Cebu City" />
                </label>
                <label className="text-sm font-semibold text-foreground">
                  Barangay
                  <input className={inputClassName} placeholder="Lahug" />
                </label>
                <label className="text-sm font-semibold text-foreground">
                  {mode === "for_rent" ? "Monthly rent (PHP)" : "Asking price (PHP)"}
                  <input className={inputClassName} inputMode="numeric" placeholder={mode === "for_rent" ? "35,000" : "7,500,000"} />
                </label>
                <label className="text-sm font-semibold text-foreground">
                  Floor or lot area (m²)
                  <input className={inputClassName} inputMode="decimal" placeholder="120" />
                </label>
                <label className="text-sm font-semibold text-foreground">
                  Bedrooms
                  <input className={inputClassName} inputMode="numeric" placeholder="3" />
                </label>
                <label className="text-sm font-semibold text-foreground">
                  Bathrooms
                  <input className={inputClassName} inputMode="numeric" placeholder="2" />
                </label>
                <label className="text-sm font-semibold text-foreground md:col-span-2">
                  Description
                  <textarea className={`${inputClassName} min-h-32 resize-y`} placeholder="Describe the space, its condition, and the features that matter to buyers." />
                </label>
              </div>
            </fieldset>

            <section className="mt-10 border-t border-border pt-8" aria-labelledby="photos-heading">
              <h2 id="photos-heading" className="font-heading text-[1.375rem] font-medium leading-[1.3] text-foreground">Photos</h2>
              <p className="mt-2 text-sm leading-6 text-muted-foreground">Add clear photos of the exterior, main rooms, and important details.</p>
              <button type="button" className="mt-4 inline-flex items-center gap-2 rounded-[var(--radius-md)] border border-border bg-background px-4 py-3 text-sm font-semibold text-foreground transition-shadow hover:shadow-[var(--shadow-sm)]">
                <ImagePlus className="h-4 w-4" aria-hidden="true" />
                Add photos
              </button>
            </section>

            <div className="mt-10 flex flex-col-reverse gap-3 border-t border-border pt-6 sm:flex-row sm:justify-end">
              <button type="button" className="rounded-[var(--radius-md)] border border-border bg-background px-5 py-3 text-sm font-semibold text-foreground transition-shadow hover:shadow-[var(--shadow-sm)]">
                Preview listing
              </button>
              <button type="submit" className="inline-flex items-center justify-center gap-2 rounded-[var(--radius-md)] bg-primary px-5 py-3 text-sm font-semibold text-[var(--color-forest)] transition-colors hover:bg-[var(--color-terracotta-hover)]">
                <Save className="h-4 w-4" aria-hidden="true" />
                Save draft
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
