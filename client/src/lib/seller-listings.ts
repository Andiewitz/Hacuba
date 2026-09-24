import type { ListingMode } from "@/lib/listings";

export type OwnedListing = {
  id: string;
  listing_mode: ListingMode;
  property_type: string;
  title: string;
  description: string;
  price_centavos?: number;
  city: string;
  barangay?: string;
  bedrooms?: number;
  bathrooms?: number;
  floor_area_sqm?: number;
  lot_area_sqm?: number;
  seller_name: string;
  contact_phone: string;
  contact_email: string;
  status: "draft" | "published" | "closed" | "archived";
  updated_at: string;
};

export type SellerDraft = {
  mode: ListingMode;
  propertyType: string;
  title: string;
  city: string;
  barangay: string;
  price: string;
  area: string;
  bedrooms: string;
  bathrooms: string;
  description: string;
  sellerName: string;
  contactPhone: string;
  contactEmail: string;
};

export type SellerAPIError = Error & { fields?: Record<string, string> };

export const emptySellerDraft: SellerDraft = {
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
  sellerName: "",
  contactPhone: "",
  contactEmail: "",
};

function stringField(value: string) {
  return value.trim() === "" ? undefined : value.trim();
}

function wholeNumber(value: string) {
  if (value.trim() === "") return undefined;
  const number = Number(value);
  return Number.isInteger(number) ? number : undefined;
}

function decimal(value: string) {
  if (value.trim() === "") return undefined;
  const number = Number(value);
  return Number.isFinite(number) ? number : undefined;
}

export function draftPayload(draft: SellerDraft) {
  const price = decimal(draft.price.replaceAll(",", ""));
  const area = decimal(draft.area);
  const propertyType = stringField(draft.propertyType)?.toLowerCase();
  const payload: Record<string, string | number> = {
    listing_mode: draft.mode,
  };
  if (propertyType) payload.property_type = propertyType;
  if (stringField(draft.title)) payload.title = draft.title.trim();
  if (stringField(draft.city)) payload.city = draft.city.trim();
  if (stringField(draft.barangay)) payload.barangay = draft.barangay.trim();
  if (stringField(draft.description)) payload.description = draft.description.trim();
  if (stringField(draft.sellerName)) payload.seller_name = draft.sellerName.trim();
  if (stringField(draft.contactPhone)) payload.contact_phone = draft.contactPhone.trim();
  if (stringField(draft.contactEmail)) payload.contact_email = draft.contactEmail.trim();
  if (price !== undefined) payload.price_centavos = Math.round(price * 100);
  if (draft.mode === "for_rent") payload.price_period = "monthly";
  if (propertyType === "lot") {
    if (area !== undefined) payload.lot_area_sqm = area;
  } else if (area !== undefined) {
    payload.floor_area_sqm = area;
  }
  const bedrooms = wholeNumber(draft.bedrooms);
  const bathrooms = wholeNumber(draft.bathrooms);
  if (bedrooms !== undefined) payload.bedrooms = bedrooms;
  if (bathrooms !== undefined) payload.bathrooms = bathrooms;
  return payload;
}

export function apiListingToDraft(listing: OwnedListing): SellerDraft {
  const area = listing.floor_area_sqm ?? listing.lot_area_sqm;
  return {
    mode: listing.listing_mode,
    propertyType: listing.property_type ? `${listing.property_type[0].toUpperCase()}${listing.property_type.slice(1)}` : "",
    title: listing.title,
    city: listing.city,
    barangay: listing.barangay ?? "",
    price: listing.price_centavos ? String(listing.price_centavos / 100) : "",
    area: area ? String(area) : "",
    bedrooms: listing.bedrooms === undefined ? "" : String(listing.bedrooms),
    bathrooms: listing.bathrooms === undefined ? "" : String(listing.bathrooms),
    description: listing.description,
    sellerName: listing.seller_name,
    contactPhone: listing.contact_phone,
    contactEmail: listing.contact_email,
  };
}

export async function sellerRequest<T>(
  path: string,
  accessToken: string,
  csrfToken: string,
  init: RequestInit = {},
): Promise<T> {
  const response = await fetch(path, {
    ...init,
    credentials: "same-origin",
    headers: {
      Authorization: `Bearer ${accessToken}`,
      "X-CSRF-Token": csrfToken,
      ...(init.body ? { "Content-Type": "application/json" } : {}),
      ...init.headers,
    },
  });
  const body = await response.json().catch(() => null) as {
    error?: string;
    fields?: Record<string, string>;
  } | null;
  if (!response.ok) {
    const error = new Error(body?.error ?? "Listing request failed.") as SellerAPIError;
    error.fields = body?.fields;
    throw error;
  }
  return body as T;
}
