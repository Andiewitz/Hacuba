"use client";

import Link from "next/link";
import { Eye, Pencil, Plus } from "lucide-react";
import PropertyCard from "@/components/property-card";

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

function publishedListing() {
  if (typeof window === "undefined") return null;

  try {
    const value = window.localStorage.getItem("hacuba:published-listing");
    return value ? (JSON.parse(value) as SellerListing) : null;
  } catch {
    return null;
  }
}

export default function SellerListingsPage() {
  const listing = publishedListing();
  const location = listing ? [listing.barangay, listing.city].filter(Boolean).join(", ") : "";
  const priceCentavos = listing ? Number(listing.price.replaceAll(",", "")) * 100 : 0;

  return (
    <main className="min-h-screen bg-background px-6 py-12 md:px-10 md:py-16 lg:px-20">
      <div className="mx-auto max-w-6xl">
        <div className="flex flex-col gap-6 border-b border-border pb-8 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <p className="text-sm font-semibold text-[var(--color-terracotta-ink)]">Seller workspace</p>
            <h1 className="mt-3 font-heading text-[2.75rem] font-bold leading-[1.15] tracking-[-0.015em] text-foreground">My listings</h1>
            <p className="mt-3 max-w-[60ch] text-[1.125rem] leading-[1.6] text-muted-foreground">Create, review, and manage every property you offer through Hacuba.</p>
          </div>
          <Link href="/seller" className="inline-flex items-center justify-center gap-2 rounded-[var(--radius-md)] bg-primary px-5 py-3 text-sm font-semibold text-[var(--color-forest)] transition-colors hover:bg-[var(--color-terracotta-hover)]">
            <Plus className="h-4 w-4" aria-hidden="true" />
            Create listing
          </Link>
        </div>

        <div className="mt-8 flex gap-6 border-b border-border text-sm font-semibold">
          <span className="border-b-2 border-[var(--color-terracotta)] px-1 pb-4 text-foreground">Published {listing ? "(1)" : "(0)"}</span>
          <span className="px-1 pb-4 text-muted-foreground">Drafts {listing ? "(1)" : "(0)"}</span>
        </div>

        {listing ? (
          <section className="mt-8 grid gap-8 rounded-[var(--radius-lg)] border border-border bg-card p-6 shadow-[var(--shadow-md)] md:grid-cols-[minmax(0,360px)_1fr] md:p-8">
            <PropertyCard
              id="published-listing"
              href={null}
              propertyType={listing.propertyType}
              location={location}
              priceCentavos={Number.isFinite(priceCentavos) ? priceCentavos : 0}
              mode={listing.mode}
              bedrooms={listing.bedrooms ? Number(listing.bedrooms) : undefined}
              bathrooms={listing.bathrooms ? Number(listing.bathrooms) : undefined}
              areaSqm={listing.area ? Number(listing.area) : undefined}
            />
            <div className="flex flex-col items-start">
              <div className="rounded-full bg-[var(--color-sage)] px-3 py-1 text-xs font-semibold text-[var(--color-forest)]">Published</div>
              <h2 className="mt-4 font-heading text-[1.75rem] font-semibold leading-[1.25] text-foreground">{listing.title}</h2>
              <p className="mt-3 max-w-[58ch] text-sm leading-6 text-muted-foreground">{listing.description}</p>
              <p className="mt-4 text-sm text-muted-foreground">{listing.photoNames.length} image{listing.photoNames.length === 1 ? "" : "s"} selected</p>
              <div className="mt-8 flex flex-wrap gap-3">
                <Link href="/seller" className="inline-flex items-center gap-2 rounded-[var(--radius-md)] border border-border bg-background px-4 py-3 text-sm font-semibold text-foreground transition-shadow hover:shadow-[var(--shadow-sm)]">
                  <Pencil className="h-4 w-4" aria-hidden="true" />
                  Edit listing
                </Link>
                <Link href="/listings/preview" className="inline-flex items-center gap-2 rounded-[var(--radius-md)] bg-[var(--color-forest)] px-4 py-3 text-sm font-semibold text-[var(--text-primary-on-dark)] transition-colors hover:bg-[var(--color-forest-surface-1)]">
                  <Eye className="h-4 w-4" aria-hidden="true" />
                  View buyer page
                </Link>
              </div>
            </div>
          </section>
        ) : (
          <section className="mt-8 rounded-[var(--radius-lg)] border border-border bg-card px-6 py-16 text-center shadow-[var(--shadow-sm)]">
            <h2 className="font-heading text-[1.75rem] font-semibold leading-[1.25] text-foreground">Your listings will appear here</h2>
            <p className="mx-auto mt-3 max-w-[52ch] text-sm leading-6 text-muted-foreground">Start with a draft, add the details buyers need, and publish when the property is ready to share.</p>
            <Link href="/seller" className="mt-6 inline-flex items-center gap-2 rounded-[var(--radius-md)] bg-primary px-5 py-3 text-sm font-semibold text-[var(--color-forest)] transition-colors hover:bg-[var(--color-terracotta-hover)]">
              <Plus className="h-4 w-4" aria-hidden="true" />
              Create your first listing
            </Link>
          </section>
        )}
      </div>
    </main>
  );
}
