import { Bath, BedDouble, Ruler } from "lucide-react";
import Image from "next/image";
import Link from "next/link";

interface PropertyCardProps {
  id: string;
  image?: string;
  propertyType: string;
  location: string;
  priceCentavos: number;
  mode: "for_sale" | "for_rent";
  bedrooms?: number;
  bathrooms?: number;
  areaSqm?: number;
  priority?: boolean;
  href?: string | null;
}

const peso = new Intl.NumberFormat("en-PH", {
  style: "currency",
  currency: "PHP",
  maximumFractionDigits: 0,
});

export default function PropertyCard({
  id,
  image,
  propertyType,
  location,
  priceCentavos,
  mode,
  bedrooms,
  bathrooms,
  areaSqm,
  priority = false,
  href,
}: PropertyCardProps) {
  const hasSpecs = bedrooms !== undefined || bathrooms !== undefined || areaSqm !== undefined;
  const destination = href === undefined ? `/listings/${id}` : href;

  const content = (
    <article className="flex flex-col gap-3">
      <div className="relative aspect-[4/3] overflow-hidden rounded-[var(--radius-lg)] bg-card shadow-[var(--shadow-sm)]">
        {image ? (
          <Image
            src={image}
            alt={`${propertyType} in ${location}`}
            fill
            unoptimized
            priority={priority}
            className="object-cover transition-transform duration-300 group-hover:scale-105"
          />
        ) : (
          <div className="flex h-full items-center justify-center px-6 text-center text-sm text-muted-foreground">
            Property image coming soon
          </div>
        )}
        <span className="absolute left-3 top-3 rounded-full bg-[var(--color-sage)] px-3 py-1 text-xs font-semibold text-[var(--color-forest)]">
          {mode === "for_rent" ? "For rent" : "For sale"}
        </span>
      </div>

      <div className="flex min-h-24 flex-col gap-1 px-0.5">
        <h3 className="font-heading text-[1.125rem] font-medium leading-[1.35] text-foreground">
          {propertyType} in {location}
        </h3>
        <p className="text-sm font-semibold text-foreground">
          {peso.format(priceCentavos / 100)}
          {mode === "for_rent" ? " / month" : ""}
        </p>
        {hasSpecs && (
          <div className="mt-1 flex flex-wrap gap-x-3 gap-y-1 text-xs text-muted-foreground">
            {bedrooms !== undefined && (
              <span className="flex items-center gap-1">
                <BedDouble className="h-3.5 w-3.5" aria-hidden="true" />
                {bedrooms} beds
              </span>
            )}
            {bathrooms !== undefined && (
              <span className="flex items-center gap-1">
                <Bath className="h-3.5 w-3.5" aria-hidden="true" />
                {bathrooms} baths
              </span>
            )}
            {areaSqm !== undefined && (
              <span className="flex items-center gap-1">
                <Ruler className="h-3.5 w-3.5" aria-hidden="true" />
                {areaSqm.toLocaleString()} m²
              </span>
            )}
          </div>
        )}
      </div>
    </article>
  );

  if (!destination) return content;

  return (
    <Link
      href={destination}
      className="group block rounded-[var(--radius-lg)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-4 focus-visible:ring-offset-background"
    >
      {content}
    </Link>
  );
}
