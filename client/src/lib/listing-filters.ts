import type { ListingFilters, ListingMode } from "@/lib/listings";

type SearchParams = Record<string, string | string[] | undefined>;

function value(params: SearchParams, name: string) {
  const item = params[name];
  return Array.isArray(item) ? item[0] : item;
}

function price(value: string | undefined) {
  if (!value || !/^\d+$/.test(value)) return undefined;
  return Number(value);
}

export function listingFiltersFromSearchParams(params: SearchParams): ListingFilters {
  const mode = value(params, "mode");
  const propertyType = value(params, "type");

  return {
    city: value(params, "city") || undefined,
    minPrice: price(value(params, "min_price")),
    maxPrice: price(value(params, "max_price")),
    propertyType: propertyType && propertyType !== "any" ? propertyType : undefined,
    mode: mode === "for_sale" || mode === "for_rent" ? (mode as ListingMode) : undefined,
  };
}
