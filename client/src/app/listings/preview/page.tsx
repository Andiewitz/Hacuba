"use client";

import Link from "next/link";
import { Bath, BedDouble, ChevronLeft, MapPin, Ruler } from "lucide-react";

type SellerListing = {
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

function listingPreview() {
  if (typeof window === "undefined") return null;

  try {
    const value = window.localStorage.getItem("hacuba:published-listing");
    return value ? (JSON.parse(value) as SellerListing) : null;
  } catch {
    return null;
  }
}

export default function ListingPreviewPage() {
  const listing = listingPreview();
  if (!listing) {
    return (
      <main className="min-h-screen bg-background px-6 py-16 md:px-10 lg:px-20">
        <section className="mx-auto max-w-xl rounded-[var(--radius-lg)] border border-border bg-card px-6 py-16 text-center shadow-[var(--shadow-sm)]">
          <h1 className="font-heading text-[2.25rem] font-semibold leading-[1.2] tracking-[-0.01em] text-foreground">No published listing yet</h1>
          <p className="mt-4 text-sm leading-6 text-muted-foreground">Publish a property from the seller workspace to see its buyer-facing page.</p>
          <Link href="/seller" className="mt-6 inline-flex rounded-[var(--radius-md)] bg-primary px-5 py-3 text-sm font-semibold text-[var(--color-forest)] transition-colors hover:bg-[var(--color-terracotta-hover)]">Go to seller workspace</Link>
        </section>
      </main>
    );
  }

  const location = [listing.barangay, listing.city].filter(Boolean).join(", ");
  const price = Number(listing.price.replaceAll(",", ""));
  const specs = [
    listing.bedrooms && { label: `${listing.bedrooms} bedrooms`, icon: BedDouble },
    listing.bathrooms && { label: `${listing.bathrooms} bathrooms`, icon: Bath },
    listing.area && { label: `${listing.area} m²`, icon: Ruler },
  ].filter(Boolean) as Array<{ label: string; icon: typeof BedDouble }>;

  return (
    <main className="min-h-screen bg-background pb-16">
      <section className="bg-[var(--color-forest)] px-6 py-10 md:px-10 lg:px-20">
        <div className="mx-auto max-w-6xl">
          <Link href="/seller/listings" className="inline-flex items-center gap-2 text-sm font-semibold text-[var(--text-body-on-dark)] underline-offset-4 hover:underline"><ChevronLeft className="h-4 w-4" aria-hidden="true" />Back to My listings</Link>
          <div className="mt-8 flex flex-col justify-between gap-6 md:flex-row md:items-end">
            <div>
              <span className="rounded-full bg-[var(--color-sage)] px-3 py-1 text-xs font-semibold text-[var(--color-forest)]">{listing.mode === "for_rent" ? "For rent" : "For sale"}</span>
              <h1 className="mt-4 font-heading text-[2.75rem] font-bold leading-[1.15] tracking-[-0.015em] text-[var(--text-primary-on-dark)]">{listing.title}</h1>
              <p className="mt-3 flex items-center gap-2 text-[1.125rem] leading-[1.6] text-[var(--text-body-on-dark)]"><MapPin className="h-5 w-5" aria-hidden="true" />{location}</p>
            </div>
            <p className="font-heading text-[2.25rem] font-semibold leading-[1.2] text-[var(--text-primary-on-dark)]">₱{Number.isFinite(price) ? price.toLocaleString() : "—"}{listing.mode === "for_rent" ? <span className="font-sans text-sm font-semibold text-[var(--text-body-on-dark)]"> / month</span> : null}</p>
          </div>
        </div>
      </section>

      <div className="mx-auto grid max-w-6xl gap-12 px-6 py-12 md:px-10 lg:grid-cols-[minmax(0,1fr)_300px] lg:px-20">
        <article>
          <section className="overflow-hidden rounded-[var(--radius-lg)] border border-border bg-card shadow-[var(--shadow-md)]">
            <div className="flex aspect-[16/8] items-center justify-center bg-[var(--color-cream-surface)] px-6 text-center">
              <div><p className="font-heading text-[1.75rem] font-semibold text-foreground">{listing.photoNames.length} property image{listing.photoNames.length === 1 ? "" : "s"}</p><p className="mt-2 text-sm text-muted-foreground">Photos are ready for secure upload when the listing service is connected.</p></div>
            </div>
          </section>

          <section className="mt-10 border-b border-border pb-10">
            <h2 className="font-heading text-[2.25rem] font-semibold leading-[1.2] tracking-[-0.01em] text-foreground">About this property</h2>
            <p className="mt-4 max-w-[70ch] whitespace-pre-wrap text-[1rem] leading-[1.6] text-muted-foreground">{listing.description}</p>
          </section>

          <section className="mt-10">
            <h2 className="font-heading text-[2.25rem] font-semibold leading-[1.2] tracking-[-0.01em] text-foreground">Property details</h2>
            <div className="mt-6 grid gap-4 sm:grid-cols-3">
              {specs.map(({ label, icon: Icon }) => <div key={label} className="flex items-center gap-3 rounded-[var(--radius-md)] border border-border bg-card p-4 text-sm font-semibold text-foreground"><Icon className="h-5 w-5 text-[var(--color-terracotta-ink)]" aria-hidden="true" />{label}</div>)}
            </div>
          </section>
        </article>

        <aside className="h-fit rounded-[var(--radius-lg)] bg-[var(--color-forest)] p-6 text-[var(--text-body-on-dark)] shadow-[var(--shadow-lg)] lg:sticky lg:top-28">
          <p className="text-sm font-semibold text-[var(--text-secondary-on-dark)]">Interested in this property?</p>
          <h2 className="mt-3 font-heading text-[1.75rem] font-semibold leading-[1.25] text-[var(--text-primary-on-dark)]">Contact the seller</h2>
          <p className="mt-3 text-sm leading-6">Contact details will appear here after the buyer and seller messaging flow is connected.</p>
          <button type="button" className="mt-6 w-full rounded-[var(--radius-md)] bg-primary px-4 py-3 text-sm font-semibold text-[var(--color-forest)] transition-colors hover:bg-[var(--color-terracotta-hover)]">Request information</button>
        </aside>
      </div>
    </main>
  );
}
