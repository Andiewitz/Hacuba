export type ListingMode = "for_sale" | "for_rent";

export interface ListingCardData {
  id: string;
  image?: string;
  propertyType: string;
  location: string;
  priceCentavos: number;
  mode: ListingMode;
  bedrooms?: number;
  bathrooms?: number;
  areaSqm?: number;
}

export interface ListingFilters {
  city?: string;
  minPrice?: number;
  maxPrice?: number;
  propertyType?: string;
  mode?: ListingMode;
}

interface ListingsResponse {
  listings: PublicListing[];
}

interface PublicListing {
  id: string;
  listing_mode: ListingMode;
  property_type: string;
  price_centavos: number;
  city: string;
  barangay?: string;
  bedrooms?: number;
  bathrooms?: number;
  floor_area_sqm?: number;
  lot_area_sqm?: number;
  images?: Array<{ object_key: string }>;
}

const catalogue: ListingCardData[] = [
  { id: "catalogue-home-mabolo", image: "https://images.unsplash.com/photo-1600585154340-be6161a56a0c?w=900&h=675&fit=crop", propertyType: "House", location: "Mabolo, Cebu City", priceCentavos: 485_000_000, mode: "for_sale", bedrooms: 4, bathrooms: 3, areaSqm: 210 },
  { id: "catalogue-home-lahug", image: "https://images.unsplash.com/photo-1600607687939-ce8a6c25118c?w=900&h=675&fit=crop", propertyType: "Condo", location: "Lahug, Cebu City", priceCentavos: 3_800_000, mode: "for_rent", bedrooms: 2, bathrooms: 2, areaSqm: 74 },
  { id: "catalogue-home-banilad", image: "https://images.unsplash.com/photo-1600566753190-17f0baa2a6c3?w=900&h=675&fit=crop", propertyType: "House", location: "Banilad, Cebu City", priceCentavos: 720_000_000, mode: "for_sale", bedrooms: 5, bathrooms: 4, areaSqm: 340 },
  { id: "catalogue-home-it-park", image: "https://images.unsplash.com/photo-1522708323590-d24dbb6b0267?w=900&h=675&fit=crop", propertyType: "Apartment", location: "Cebu IT Park", priceCentavos: 2_400_000, mode: "for_rent", bedrooms: 1, bathrooms: 1, areaSqm: 42 },
  { id: "catalogue-lot-busay", image: "https://images.unsplash.com/photo-1500382017468-9049fed747ef?w=900&h=675&fit=crop", propertyType: "Lot", location: "Busay, Cebu City", priceCentavos: 128_000_000, mode: "for_sale", areaSqm: 800 },
  { id: "catalogue-lot-liloan", image: "https://images.unsplash.com/photo-1416879595882-3373a0480b5b?w=900&h=675&fit=crop", propertyType: "Lot", location: "Liloan, Cebu", priceCentavos: 65_000_000, mode: "for_sale", areaSqm: 500 },
  { id: "catalogue-lot-cordova", image: "https://images.unsplash.com/photo-1501854140801-50d01698950b?w=900&h=675&fit=crop", propertyType: "Lot", location: "Cordova, Cebu", priceCentavos: 45_000_000, mode: "for_sale", areaSqm: 350 },
  { id: "catalogue-commercial-it-park", image: "https://images.unsplash.com/photo-1486406146926-c627a92ad1ab?w=900&h=675&fit=crop", propertyType: "Commercial", location: "Cebu IT Park", priceCentavos: 9_500_000, mode: "for_rent", areaSqm: 160 },
  { id: "catalogue-commercial-mandaue", image: "https://images.unsplash.com/photo-1497366216548-37526070297c?w=900&h=675&fit=crop", propertyType: "Commercial", location: "Mandaue City", priceCentavos: 110_000_000, mode: "for_sale", areaSqm: 220 },
  { id: "catalogue-commercial-srp", image: "https://images.unsplash.com/photo-1497366811353-6870744d04b2?w=900&h=675&fit=crop", propertyType: "Commercial", location: "South Road Properties", priceCentavos: 7_200_000, mode: "for_rent", areaSqm: 110 },
];

function imageUrl(objectKey?: string) {
  const baseUrl = process.env.LISTINGS_IMAGE_BASE_URL?.replace(/\/$/, "");
  return baseUrl && objectKey ? `${baseUrl}/${objectKey}` : undefined;
}

function toCard(listing: PublicListing): ListingCardData {
  return {
    id: listing.id,
    image: imageUrl(listing.images?.[0]?.object_key),
    propertyType: listing.property_type,
    location: [listing.barangay, listing.city].filter(Boolean).join(", "),
    priceCentavos: listing.price_centavos,
    mode: listing.listing_mode,
    bedrooms: listing.bedrooms,
    bathrooms: listing.bathrooms,
    areaSqm: listing.floor_area_sqm ?? listing.lot_area_sqm,
  };
}

function matchesFilters(listing: ListingCardData, filters: ListingFilters) {
  const price = listing.priceCentavos;
  const city = filters.city?.toLowerCase();
  const propertyType = filters.propertyType?.toLowerCase();

  return (
    (!city || listing.location.toLowerCase().includes(city)) &&
    (!propertyType || listing.propertyType.toLowerCase() === propertyType) &&
    (!filters.mode || listing.mode === filters.mode) &&
    (filters.minPrice === undefined || price >= filters.minPrice) &&
    (filters.maxPrice === undefined || price <= filters.maxPrice)
  );
}

function fixtureListings(filters: ListingFilters) {
  return catalogue.filter((listing) => matchesFilters(listing, filters));
}

export async function getPublicListings(filters: ListingFilters = {}) {
  const apiUrl = process.env.LISTINGS_API_URL;
  if (!apiUrl) return fixtureListings(filters);

  const query = new URLSearchParams();
  if (filters.city) query.set("city", filters.city);
  if (filters.propertyType) query.set("type", filters.propertyType.toLowerCase());
  if (filters.mode) query.set("mode", filters.mode);
  if (filters.minPrice !== undefined) query.set("min_price", String(filters.minPrice));
  if (filters.maxPrice !== undefined) query.set("max_price", String(filters.maxPrice));

  try {
    const response = await fetch(`${apiUrl.replace(/\/$/, "")}/listings?${query}`, {
      next: { revalidate: 60 },
    });
    if (!response.ok) return [];
    const data = (await response.json()) as ListingsResponse;
    return data.listings.map(toCard);
  } catch {
    return [];
  }
}
