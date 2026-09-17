export interface Listing {
  image: string;
  type: string;
  location: string;
  price: number;
  nights: number;
  rating: number;
  isGuestFavorite?: boolean;
  isNew?: boolean;
}

export const homes: Listing[] = [
  {
    image: "https://images.unsplash.com/photo-1600585154340-be6161a56a0c?w=600&h=600&fit=crop",
    type: "House",
    location: "Mabolo",
    price: 2576,
    nights: 2,
    rating: 4.84,
    isGuestFavorite: true,
    isNew: true,
  },
  {
    image: "https://images.unsplash.com/photo-1522708323590-d24dbb6b0267?w=600&h=600&fit=crop",
    type: "House",
    location: "Lahug",
    price: 2067,
    nights: 2,
    rating: 4.8,
  },
  {
    image: "https://images.unsplash.com/photo-1600607687939-ce8a6c25118c?w=600&h=600&fit=crop",
    type: "House",
    location: "Kasambagan",
    price: 3540,
    nights: 2,
    rating: 5.0,
    isGuestFavorite: true,
    isNew: true,
  },
  {
    image: "https://images.unsplash.com/photo-1600566753190-17f0baa2a6c3?w=600&h=600&fit=crop",
    type: "House",
    location: "Banilad",
    price: 2314,
    nights: 2,
    rating: 4.75,
  },
  {
    image: "https://images.unsplash.com/photo-1600585154526-990dced4db0d?w=600&h=600&fit=crop",
    type: "House",
    location: "Talamban",
    price: 3484,
    nights: 2,
    rating: 4.88,
    isGuestFavorite: true,
  },
  {
    image: "https://images.unsplash.com/photo-1600573472592-401b489a3cdc?w=600&h=600&fit=crop",
    type: "House",
    location: "Guadalupe",
    price: 2776,
    nights: 2,
    rating: 4.85,
    isGuestFavorite: true,
  },
  {
    image: "https://images.unsplash.com/photo-1564013799919-ab600027ffc6?w=600&h=600&fit=crop",
    type: "House",
    location: "AS Fortuna",
    price: 3100,
    nights: 2,
    rating: 4.72,
    isNew: true,
  },
  {
    image: "https://images.unsplash.com/photo-1582407947304-fd86f028f716?w=600&h=600&fit=crop",
    type: "House",
    location: "Camputhaw",
    price: 2450,
    nights: 2,
    rating: 4.8,
  },
];

export const lots: Listing[] = [
  {
    image: "https://images.unsplash.com/photo-1500382017468-9049fed747ef?w=600&h=600&fit=crop",
    type: "Lot",
    location: "Busay",
    price: 4200,
    nights: 2,
    rating: 4.92,
    isNew: true,
  },
  {
    image: "https://images.unsplash.com/photo-1416879595882-3373a0480b5b?w=600&h=600&fit=crop",
    type: "Lot",
    location: "Balamban",
    price: 3100,
    nights: 2,
    rating: 4.7,
  },
  {
    image: "https://images.unsplash.com/photo-1588880331179-bc9b93a8cb5e?w=600&h=600&fit=crop",
    type: "Lot",
    location: "Liloan",
    price: 5600,
    nights: 2,
    rating: 4.95,
    isGuestFavorite: true,
    isNew: true,
  },
  {
    image: "https://images.unsplash.com/photo-1501854140801-50d01698950b?w=600&h=600&fit=crop",
    type: "Lot",
    location: "Cordova",
    price: 2800,
    nights: 2,
    rating: 4.6,
  },
  {
    image: "https://images.unsplash.com/photo-1473973266408-ed4e27abdd47?w=600&h=600&fit=crop",
    type: "Lot",
    location: "Catmon",
    price: 3450,
    nights: 2,
    rating: 4.78,
  },
  {
    image: "https://images.unsplash.com/photo-1500076656116-558758c991c1?w=600&h=600&fit=crop",
    type: "Lot",
    location: "Danao",
    price: 2900,
    nights: 2,
    rating: 4.65,
  },
  {
    image: "https://images.unsplash.com/photo-1588880331179-bc9b93a8cb5e?w=600&h=600&fit=crop",
    type: "Lot",
    location: "Consolacion",
    price: 3200,
    nights: 2,
    rating: 4.7,
    isNew: true,
  },
  {
    image: "https://images.unsplash.com/photo-1501854140801-50d01698950b?w=600&h=600&fit=crop",
    type: "Lot",
    location: "Compostela",
    price: 2600,
    nights: 2,
    rating: 4.55,
  },
];

export const commercial: Listing[] = [
  {
    image: "https://images.unsplash.com/photo-1486406146926-c627a92ad1ab?w=600&h=600&fit=crop",
    type: "Commercial",
    location: "Cebu IT Park",
    price: 8500,
    nights: 2,
    rating: 4.9,
    isGuestFavorite: true,
    isNew: true,
  },
  {
    image: "https://images.unsplash.com/photo-1497366216548-37526070297c?w=600&h=600&fit=crop",
    type: "Commercial",
    location: "Ayala Center",
    price: 7200,
    nights: 2,
    rating: 4.85,
  },
  {
    image: "https://images.unsplash.com/photo-1497366811353-6870744d04b2?w=600&h=600&fit=crop",
    type: "Commercial",
    location: "Mandaue",
    price: 6300,
    nights: 2,
    rating: 4.72,
  },
  {
    image: "https://images.unsplash.com/photo-1564013799919-ab600027ffc6?w=600&h=600&fit=crop",
    type: "Commercial",
    location: "Lapu-Lapu",
    price: 9100,
    nights: 2,
    rating: 4.95,
    isNew: true,
  },
  {
    image: "https://images.unsplash.com/photo-1582407947304-fd86f028f716?w=600&h=600&fit=crop",
    type: "Commercial",
    location: "SRP",
    price: 7800,
    nights: 2,
    rating: 4.88,
  },
  {
    image: "https://images.unsplash.com/photo-1503387762-592deb58ef4e?w=600&h=600&fit=crop",
    type: "Commercial",
    location: "Talisay",
    price: 5400,
    nights: 2,
    rating: 4.7,
  },
  {
    image: "https://images.unsplash.com/photo-1486406146926-c627a92ad1ab?w=600&h=600&fit=crop",
    type: "Commercial",
    location: "Banilad",
    price: 6800,
    nights: 2,
    rating: 4.78,
    isNew: true,
  },
  {
    image: "https://images.unsplash.com/photo-1497366216548-37526070297c?w=600&h=600&fit=crop",
    type: "Commercial",
    location: "Kasambagan",
    price: 7100,
    nights: 2,
    rating: 4.82,
  },
];

export const nearbyAreas = [
  "Mandaue",
  "Lapu-Lapu",
  "Talisay",
  "Minglanilla",
  "Consolacion",
];

export function getNewListings(type?: string) {
  const all = type === "House" ? homes : type === "Lot" ? lots : type === "Commercial" ? commercial : [...homes, ...lots, ...commercial];
  return all.filter((l) => l.isNew);
}

export function getListingsByType(type: string) {
  return (type === "House" ? homes : type === "Lot" ? lots : commercial);
}

export function getAvailableNextMonth(type?: string) {
  const all = type === "House" ? homes : type === "Lot" ? lots : type === "Commercial" ? commercial : [...homes, ...lots, ...commercial];
  return all.filter((l) => !l.isNew);
}
