"use client";

import Link from "next/link";
import { Archive, Eye, Pencil, Plus, Store, XCircle } from "lucide-react";
import { useEffect, useState } from "react";
import PropertyCard from "@/components/property-card";
import { useAppSelector } from "@/lib/hooks";
import { sellerRequest, type OwnedListing } from "@/lib/seller-listings";

function labelForStatus(status: OwnedListing["status"]) {
  return status[0].toUpperCase() + status.slice(1);
}

export default function SellerListingsPage() {
  const { user, accessToken, csrfToken, initialized } = useAppSelector((state) => state.auth);
  const [listings, setListings] = useState<OwnedListing[]>([]);
  const [error, setError] = useState<string>();
  const [loading, setLoading] = useState(true);
  const [mutatingID, setMutatingID] = useState<string>();

  useEffect(() => {
    if (!initialized) return;
    if (user?.role !== "seller" || !accessToken || !csrfToken) {
      return;
    }
    let active = true;
    sellerRequest<{ listings: OwnedListing[] }>("/api/listings/mine", accessToken, csrfToken)
      .then((response) => {
        if (active) setListings(response.listings);
      })
      .catch((requestError: unknown) => {
        if (active) setError(requestError instanceof Error ? requestError.message : "Couldn't load your listings.");
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => { active = false; };
  }, [accessToken, csrfToken, initialized, user?.role]);

  const published = listings.filter((listing) => listing.status === "published");
  const drafts = listings.filter((listing) => listing.status === "draft");

  const transitionListing = async (
    listingID: string,
    action: "unpublish" | "close" | "archive",
  ) => {
    if (!accessToken || !csrfToken) return;
    setMutatingID(listingID);
    setError(undefined);
    try {
      const listing = await sellerRequest<OwnedListing>(
        action === "archive" ? `/api/listings/${listingID}` : `/api/listings/${listingID}/${action}`,
        accessToken,
        csrfToken,
        { method: action === "archive" ? "DELETE" : "POST" },
      );
      setListings((current) => current.map((item) => item.id === listingID ? listing : item));
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Couldn't update this listing.");
    } finally {
      setMutatingID(undefined);
    }
  };

  return (
    <main className="min-h-screen bg-background px-6 py-12 md:px-10 md:py-16 lg:px-20">
      <div className="mx-auto max-w-6xl">
        <div className="flex flex-col gap-6 border-b border-border pb-8 sm:flex-row sm:items-end sm:justify-between"><div><p className="text-sm font-semibold text-[var(--color-terracotta-ink)]">Seller workspace</p><h1 className="mt-3 font-heading text-[2.75rem] font-bold leading-[1.15] tracking-[-0.015em] text-foreground">My listings</h1><p className="mt-3 max-w-[60ch] text-[1.125rem] leading-[1.6] text-muted-foreground">Create, review, and manage every property you offer through Hacuba.</p></div><Link href="/seller" className="inline-flex items-center justify-center gap-2 rounded-[var(--radius-md)] bg-primary px-5 py-3 text-sm font-semibold text-[var(--color-forest)] transition-colors hover:bg-[var(--color-terracotta-hover)]"><Plus className="h-4 w-4" aria-hidden="true" />Create listing</Link></div>
        {!initialized || (user?.role === "seller" && loading) ? <p className="mt-8 text-sm text-muted-foreground">Loading your seller listings…</p> : user?.role !== "seller" ? <section className="mt-8 rounded-[var(--radius-lg)] border border-border bg-card px-6 py-16 text-center shadow-[var(--shadow-sm)]"><h2 className="font-heading text-[1.75rem] font-semibold leading-[1.25] text-foreground">Seller access is required</h2><p className="mx-auto mt-3 max-w-[52ch] text-sm leading-6 text-muted-foreground">Activate seller access from the navigation, then return to manage your listings.</p></section> : error ? <p role="alert" className="mt-8 text-sm text-[var(--color-terracotta-ink)]">{error}</p> : listings.length === 0 ? <section className="mt-8 rounded-[var(--radius-lg)] border border-border bg-card px-6 py-16 text-center shadow-[var(--shadow-sm)]"><h2 className="font-heading text-[1.75rem] font-semibold leading-[1.25] text-foreground">Your listings will appear here</h2><p className="mx-auto mt-3 max-w-[52ch] text-sm leading-6 text-muted-foreground">Start with a draft, add the details buyers need, and publish when the property is ready to share.</p><Link href="/seller" className="mt-6 inline-flex items-center gap-2 rounded-[var(--radius-md)] bg-primary px-5 py-3 text-sm font-semibold text-[var(--color-forest)] transition-colors hover:bg-[var(--color-terracotta-hover)]"><Plus className="h-4 w-4" aria-hidden="true" />Create your first listing</Link></section> : <>
          <div className="mt-8 flex gap-6 border-b border-border text-sm font-semibold"><span className="border-b-2 border-[var(--color-terracotta)] px-1 pb-4 text-foreground">Published ({published.length})</span><span className="px-1 pb-4 text-muted-foreground">Drafts ({drafts.length})</span></div>
          <div className="mt-8 space-y-6">{listings.map((listing) => <section key={listing.id} className="grid gap-8 rounded-[var(--radius-lg)] border border-border bg-card p-6 shadow-[var(--shadow-md)] md:grid-cols-[minmax(0,280px)_1fr] md:p-8"><PropertyCard id={listing.id} href={listing.status === "published" ? `/listings/${listing.id}` : null} propertyType={listing.property_type || "Property"} location={[listing.barangay, listing.city].filter(Boolean).join(", ") || "Location pending"} priceCentavos={listing.price_centavos ?? 0} mode={listing.listing_mode} bedrooms={listing.bedrooms} bathrooms={listing.bathrooms} areaSqm={listing.floor_area_sqm ?? listing.lot_area_sqm} /><div className="flex flex-col items-start"><div className="rounded-full bg-[var(--color-sage)] px-3 py-1 text-xs font-semibold text-[var(--color-forest)]">{labelForStatus(listing.status)}</div><h2 className="mt-4 font-heading text-[1.75rem] font-semibold leading-[1.25] text-foreground">{listing.title || "Untitled draft"}</h2><p className="mt-3 max-w-[58ch] text-sm leading-6 text-muted-foreground">{listing.description || "Add property details to prepare this draft for review."}</p><p className="mt-4 text-sm text-muted-foreground">Last updated {new Date(listing.updated_at).toLocaleDateString("en-PH", { month: "short", day: "numeric", year: "numeric" })}</p><div className="mt-8 flex flex-wrap gap-3"><Link href={`/seller?listing=${listing.id}`} className="inline-flex items-center gap-2 rounded-[var(--radius-md)] border border-border bg-background px-4 py-3 text-sm font-semibold text-foreground transition-shadow hover:shadow-[var(--shadow-sm)]"><Pencil className="h-4 w-4" aria-hidden="true" />Edit listing</Link>{listing.status === "published" && <><Link href={`/listings/${listing.id}`} className="inline-flex items-center gap-2 rounded-[var(--radius-md)] bg-[var(--color-forest)] px-4 py-3 text-sm font-semibold text-[var(--text-primary-on-dark)] transition-colors hover:bg-[var(--color-forest-surface-1)]"><Eye className="h-4 w-4" aria-hidden="true" />View buyer page</Link><button type="button" disabled={mutatingID === listing.id} onClick={() => transitionListing(listing.id, "unpublish")} className="inline-flex items-center gap-2 rounded-[var(--radius-md)] border border-border px-4 py-3 text-sm font-semibold text-foreground disabled:opacity-50"><Store className="h-4 w-4" aria-hidden="true" />Unpublish</button><button type="button" disabled={mutatingID === listing.id} onClick={() => transitionListing(listing.id, "close")} className="inline-flex items-center gap-2 rounded-[var(--radius-md)] border border-border px-4 py-3 text-sm font-semibold text-foreground disabled:opacity-50"><XCircle className="h-4 w-4" aria-hidden="true" />Mark closed</button></>}<button type="button" disabled={mutatingID === listing.id || listing.status === "archived"} onClick={() => transitionListing(listing.id, "archive")} className="inline-flex items-center gap-2 rounded-[var(--radius-md)] border border-border px-4 py-3 text-sm font-semibold text-foreground disabled:opacity-50"><Archive className="h-4 w-4" aria-hidden="true" />Archive</button></div></div></section>)}</div>
        </>}
      </div>
    </main>
  );
}
