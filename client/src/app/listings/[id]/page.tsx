import { Bath, BedDouble, ChevronLeft, MapPin, Ruler } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { notFound } from "next/navigation";
import { getPublicListing, listingImageUrl } from "@/lib/listings";

type ListingPageProps = {
  params: Promise<{ id: string }>;
};

const peso = new Intl.NumberFormat("en-PH", {
  style: "currency",
  currency: "PHP",
  maximumFractionDigits: 0,
});

function propertyTypeLabel(propertyType: string) {
  return propertyType.replace(/[_-]/g, " ").replace(/\b\w/g, (letter) => letter.toUpperCase());
}

export default async function ListingPage({ params }: ListingPageProps) {
  const { id } = await params;
  const listing = await getPublicListing(id);

  if (!listing) notFound();

  const imageUrl = listingImageUrl(listing.images?.[0]?.object_key);
  const location = [listing.barangay, listing.city].filter(Boolean).join(", ");
  const isLot = listing.property_type.toLowerCase() === "lot";
  const area = isLot ? listing.lot_area_sqm : listing.floor_area_sqm ?? listing.lot_area_sqm;
  const details = [
    listing.bedrooms !== undefined
      ? { icon: BedDouble, label: "Bedrooms", value: String(listing.bedrooms) }
      : null,
    listing.bathrooms !== undefined
      ? { icon: Bath, label: "Bathrooms", value: String(listing.bathrooms) }
      : null,
    area !== undefined
      ? { icon: Ruler, label: isLot ? "Lot area" : "Floor area", value: `${area.toLocaleString()} m²` }
      : null,
  ].filter((detail): detail is NonNullable<typeof detail> => detail !== null);

  return (
    <main className="min-h-screen bg-background">
      <section className="bg-[var(--color-forest)] px-6 py-10 md:px-10 md:py-12 lg:px-20">
        <div className="mx-auto max-w-[1200px]">
          <Link
            href="/"
            className="inline-flex items-center gap-2 text-sm font-semibold text-[var(--text-body-on-dark)] underline-offset-4 hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-4 focus-visible:ring-offset-[var(--color-forest)]"
          >
            <ChevronLeft className="h-4 w-4" aria-hidden="true" />
            Back to properties
          </Link>
          <div className="mt-8 max-w-3xl">
            <span className="inline-flex rounded-full bg-[var(--color-sage)] px-3 py-1 text-xs font-semibold text-[var(--color-forest)]">
              {listing.listing_mode === "for_rent" ? "For rent" : "For sale"}
            </span>
            <h1 className="mt-4 font-heading text-[2.75rem] font-bold leading-[1.15] tracking-[-0.015em] text-[var(--text-primary-on-dark)] md:text-[3.5rem] md:leading-[1.1] md:tracking-[-0.02em]">
              {listing.title}
            </h1>
            <p className="mt-4 text-[1.125rem] font-medium leading-[1.6] text-[var(--text-body-on-dark)]">
              {peso.format(listing.price_centavos / 100)}
              {listing.listing_mode === "for_rent" ? " / month" : ""}
            </p>
            <p className="mt-2 flex items-center gap-2 text-sm text-[var(--text-secondary-on-dark)]">
              <MapPin className="h-4 w-4" aria-hidden="true" />
              {location}
            </p>
          </div>
        </div>
      </section>

      <div className="mx-auto grid max-w-[1200px] gap-12 px-6 py-12 md:px-10 lg:grid-cols-[minmax(0,1fr)_320px] lg:px-20 lg:py-16">
        <div className="min-w-0">
          <section aria-label="Property photo" className="overflow-hidden rounded-[var(--radius-lg)] bg-card shadow-[var(--shadow-md)]">
            {imageUrl ? (
              <div className="relative aspect-[4/3]">
                <Image
                  src={imageUrl}
                  alt={listing.title}
                  fill
                  priority
                  unoptimized
                  className="object-cover"
                />
              </div>
            ) : (
              <div className="flex aspect-[4/3] flex-col items-center justify-center gap-2 px-6 text-center">
                <span className="font-heading text-[1.375rem] font-medium leading-[1.3] text-foreground">Photos are being prepared</span>
                <span className="max-w-[38ch] text-sm leading-[1.5] text-muted-foreground">This listing does not have a public image URL yet.</span>
              </div>
            )}
          </section>

          <section className="mt-12">
            <p className="text-sm font-semibold text-[var(--color-terracotta-ink)]">{propertyTypeLabel(listing.property_type)}</p>
            <h2 className="mt-2 font-heading text-[2.25rem] font-semibold leading-[1.2] tracking-[-0.01em] text-foreground">About this property</h2>
            <p className="mt-4 max-w-[70ch] text-base leading-[1.6] text-[var(--text-secondary-on-light)]">{listing.description}</p>
          </section>

          {details.length > 0 && (
            <section className="mt-12" aria-labelledby="property-details-heading">
              <h2 id="property-details-heading" className="font-heading text-[1.75rem] font-semibold leading-[1.25] tracking-[-0.005em] text-foreground">Property details</h2>
              <dl className="mt-6 grid gap-4 sm:grid-cols-3">
                {details.map(({ icon: Icon, label, value }) => (
                  <div key={label} className="rounded-[var(--radius-md)] border border-border bg-card p-4 shadow-[var(--shadow-sm)]">
                    <dt className="flex items-center gap-2 text-sm text-muted-foreground">
                      <Icon className="h-4 w-4" aria-hidden="true" />
                      {label}
                    </dt>
                    <dd className="mt-2 font-heading text-[1.375rem] font-medium leading-[1.3] text-foreground">{value}</dd>
                  </div>
                ))}
              </dl>
            </section>
          )}
        </div>

        <aside className="h-fit rounded-[var(--radius-lg)] bg-[var(--color-forest-surface-1)] p-6 shadow-[var(--shadow-md)] lg:sticky lg:top-6">
          <p className="text-sm font-semibold text-[var(--text-secondary-on-dark)]">Listing support</p>
          <h2 className="mt-2 font-heading text-[1.75rem] font-semibold leading-[1.25] tracking-[-0.005em] text-[var(--text-primary-on-dark)]">Interested in this property?</h2>
          <p className="mt-4 text-sm leading-[1.5] text-[var(--text-body-on-dark)]">Direct seller messaging will appear here when it is available.</p>
          <Link
            href="/help"
            className="mt-6 inline-flex rounded-[var(--radius-md)] border border-[var(--color-forest-border-strong)] px-4 py-3 text-sm font-semibold text-[var(--text-body-on-dark)] underline-offset-4 hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-4 focus-visible:ring-offset-[var(--color-forest-surface-1)]"
          >
            Read how Hacuba works
          </Link>
        </aside>
      </div>
    </main>
  );
}
