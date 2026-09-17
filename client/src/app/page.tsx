import Navbar from "@/components/navbar";
import SearchBar from "@/components/search-bar";
import PropertyCard from "@/components/property-card";

const popularListings = [
  {
    image: "https://images.unsplash.com/photo-1600585154340-be6161a56a0c?w=600&h=600&fit=crop",
    type: "Hotel",
    location: "Mabolo",
    price: 2576,
    nights: 2,
    rating: 4.84,
    isGuestFavorite: true,
  },
  {
    image: "https://images.unsplash.com/photo-1522708323590-d24dbb6b0267?w=600&h=600&fit=crop",
    type: "Apartment",
    location: "Cebu City",
    price: 2067,
    nights: 2,
    rating: 4.8,
  },
  {
    image: "https://images.unsplash.com/photo-1600607687939-ce8a6c25118c?w=600&h=600&fit=crop",
    type: "Apartment",
    location: "Kasambagan",
    price: 3540,
    nights: 2,
    rating: 5.0,
    isGuestFavorite: true,
  },
  {
    image: "https://images.unsplash.com/photo-1600566753190-17f0baa2a6c3?w=600&h=600&fit=crop",
    type: "Condo",
    location: "Cebu City",
    price: 2314,
    nights: 2,
    rating: 4.75,
  },
  {
    image: "https://images.unsplash.com/photo-1600585154526-990dced4db0d?w=600&h=600&fit=crop",
    type: "Condo",
    location: "Cebu City",
    price: 3484,
    nights: 2,
    rating: 4.88,
    isGuestFavorite: true,
  },
  {
    image: "https://images.unsplash.com/photo-1600573472592-401b489a3cdc?w=600&h=600&fit=crop",
    type: "Apartment",
    location: "Guadalupe",
    price: 2776,
    nights: 2,
    rating: 4.85,
    isGuestFavorite: true,
  },
];

export default function Home() {
  return (
    <div className="min-h-screen bg-background">
      <Navbar />

      <section className="px-6 pt-6 pb-4 md:px-10 lg:px-20">
        <SearchBar />
      </section>

      <section className="px-6 py-8 md:px-10 lg:px-20">
        <div className="mx-auto max-w-[1760px]">
          <div className="flex items-center gap-3 mb-6">
            <h2 className="text-2xl font-semibold text-foreground font-heading">
              Popular homes in Cebu City
            </h2>
            <button className="flex h-8 w-8 items-center justify-center rounded-full border border-border text-foreground transition-colors hover:shadow-sm">
              →
            </button>
          </div>

          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6">
            {popularListings.map((listing, index) => (
              <PropertyCard
                key={`${listing.type}-${listing.location}`}
                {...listing}
                priority={index === 0}
              />
            ))}
          </div>
        </div>
      </section>
    </div>
  );
}
