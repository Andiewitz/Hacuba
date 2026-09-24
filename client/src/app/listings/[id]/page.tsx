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

  const images = (listing.images ?? [])
    .map((image) => listingImageUrl(image.object_key))
    .filter((image): image is string => Boolean(image));
  const location = [listing.barangay, listing.city].filter(Boolean).join(", ");
  const propertyType = propertyTypeLabel(listing.property_type);
  const isLot = listing.property_type.toLowerCase() === "lot";
  const area = isLot ? listing.lot_area_sqm : listing.floor_area_sqm ?? listing.lot_area_sqm;
  const facts = [
    listing.bedrooms !== undefined ? { icon: BedDouble, label: "Bedrooms", value: String(listing.bedrooms) } : null,
    listing.bathrooms !== undefined ? { icon: Bath, label: "Bathrooms", value: String(listing.bathrooms) } : null,
    area !== undefined ? { icon: Ruler, label: isLot ? "Lot area" : "Floor area", value: `${area.toLocaleString()} m²` } : null,
  ].filter((fact): fact is NonNullable<typeof fact> => fact !== null);

  return (
    <main className="min-h-screen bg-background pb-16">
      <section className="border-b border-[var(--color-forest-border-subtle)] bg-[var(--color-forest)] px-6 py-4 md:px-10 lg:px-20">
        <div className="mx-auto max-w-[1440px]">
          <Link
            href="/"
            className="inline-flex items-center gap-2 text-sm font-semibold text-[var(--text-body-on-dark)] underline-offset-4 hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-4 focus-visible:ring-offset-[var(--color-forest)]"
          >
            <ChevronLeft className="h-4 w-4" aria-hidden="true" />
            All properties
          </Link>
        </div>
      </section>

      <div className="mx-auto max-w-[1440px] px-6 pt-6 md:px-10 md:pt-8 lg:px-20">
        <section aria-label="Property gallery" className="overflow-hidden rounded-[var(--radius-lg)] bg-[var(--color-forest-surface-1)] shadow-[var(--shadow-lg)]">
          {images.length > 0 ? (
            <div className={images.length > 1 ? "grid aspect-[16/10] grid-cols-[minmax(0,2fr)_minmax(180px,1fr)] gap-1 bg-[var(--color-forest)]" : "relative aspect-[16/10]"}>
              <div className="relative min-h-0">
                <Image src={images[0]} alt={listing.title} fill priority unoptimized className="object-cover" />
                <span className="absolute left-4 top-4 rounded-full bg-[var(--color-sage)] px-3 py-1 text-xs font-semibold text-[var(--color-forest)]">
                  {listing.listing_mode === "for_rent" ? "For rent" : "For sale"}
                </span>
              </div>
              {images.length > 1 && (
                <div className="grid min-h-0 grid-rows-2 gap-1">
                  {images.slice(1, 3).map((image, index) => (
                    <div key={image} className="relative min-h-0">
                      <Image src={image} alt={`${listing.title}, photo ${index + 2}`} fill unoptimized className="object-cover" />
                    </div>
                  ))}
                  {images.length === 2 && <div className="bg-[var(--color-forest-surface-2)]" />}
                </div>
              )}
            </div>
          ) : (
            <div className="flex aspect-[16/10] items-center justify-center px-6 text-center">
              <p className="max-w-[36ch] text-sm leading-[1.5] text-[var(--text-body-on-dark)]">No public photos have been added to this listing yet.</p>
            </div>
          )}
        </section>

        <div className="grid gap-8 py-10 lg:grid-cols-[minmax(0,1fr)_360px] lg:items-start lg:py-12">
          <article className="min-w-0">
            <div className="flex flex-wrap items-center gap-3">
              <span className="rounded-full bg-[var(--color-sage)] px-3 py-1 text-xs font-semibold text-[var(--color-forest)]">{propertyType}</span>
              <p className="text-sm text-[var(--text-secondary-on-light)]">{location}</p>
            </div>
            <h1 className="mt-4 max-w-[22ch] font-heading text-[2.75rem] font-bold leading-[1.15] tracking-[-0.015em] text-foreground md:text-[3.5rem] md:leading-[1.1] md:tracking-[-0.02em]">{listing.title}</h1>
            <p className="mt-5 flex items-center gap-2 text-sm text-[var(--text-secondary-on-light)]">
              <MapPin className="h-4 w-4 shrink-0" aria-hidden="true" />
              {location}
            </p>

            <section className="mt-10 border-y border-border py-6" aria-labelledby="at-a-glance-heading">
              <h2 id="at-a-glance-heading" className="sr-only">At a glance</h2>
              <dl className="grid gap-5 sm:grid-cols-3">
                {facts.map(({ icon: Icon, label, value }) => (
                  <div key={label}>
                    <dt className="flex items-center gap-2 text-sm text-muted-foreground"><Icon className="h-4 w-4" aria-hidden="true" />{label}</dt>
                    <dd className="mt-2 font-heading text-[1.375rem] font-medium leading-[1.3] text-foreground">{value}</dd>
                  </div>
                ))}
              </dl>
            </section>

            <section className="mt-10 max-w-[72ch]" aria-labelledby="overview-heading">
              <p className="text-sm font-semibold text-[var(--color-terracotta-ink)]">Property overview</p>
              <h2 id="overview-heading" className="mt-2 font-heading text-[2.25rem] font-semibold leading-[1.2] tracking-[-0.01em] text-foreground">A closer look</h2>
              <p className="mt-4 text-base leading-[1.6] text-[var(--text-secondary-on-light)]">{listing.description}</p>
            </section>
          </article>

          <aside className="rounded-[var(--radius-lg)] border border-border bg-card p-6 shadow-[var(--shadow-md)] lg:sticky lg:top-24">
            <p className="text-sm font-semibold text-[var(--text-secondary-on-light)]">Listing summary</p>
            <p className="mt-2 font-heading text-[2.25rem] font-semibold leading-[1.2] tracking-[-0.01em] text-foreground">
              {peso.format(listing.price_centavos / 100)}
            </p>
            {listing.listing_mode === "for_rent" && <p className="mt-1 text-sm text-muted-foreground">per month</p>}
            <div className="my-6 h-px bg-border" />
            <dl className="space-y-5 text-sm">
              <div>
                <dt className="text-muted-foreground">Property type</dt>
                <dd className="mt-1 font-semibold text-foreground">{propertyType}</dd>
              </div>
              <div>
                <dt className="text-muted-foreground">Availability</dt>
                <dd className="mt-1 font-semibold text-foreground">{listing.listing_mode === "for_rent" ? "For rent" : "For sale"}</dd>
              </div>
              <div>
                <dt className="text-muted-foreground">Location</dt>
                <dd className="mt-1 font-semibold text-foreground">{location}</dd>
              </div>
            </dl>
          </aside>
        </div>
      </div>
    </main>
  );
}
