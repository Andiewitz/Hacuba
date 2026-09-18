import SearchBar from "@/components/search-bar";
import ListingSection from "@/components/listing-section";
import {
  getListingsByType,
  getNewListings,
  getAvailableNextMonth,
  nearbyAreas,
} from "@/data/listings";

export default function CommercialPage() {
  const newListings = getNewListings("Commercial");
  const listings = getListingsByType("Commercial");
  const nextMonth = getAvailableNextMonth("Commercial");

  return (
    <div className="min-h-screen bg-background">
      <section className="px-6 pt-6 pb-4 md:px-10 lg:px-20">
        <SearchBar />
      </section>
      <ListingSection title="New listings" listings={newListings} sectionIndex={0} />
      <ListingSection title="Relevant commercial properties in Cebu City" listings={listings} sectionIndex={1} />
      {nearbyAreas.map((area, i) => (
        <ListingSection
          key={area}
          title={`Available in ${area}`}
          listings={listings.slice(i * 2, i * 2 + 4)}
          sectionIndex={i + 2}
        />
      ))}
      <ListingSection title="Available next month" listings={nextMonth} sectionIndex={nearbyAreas.length + 2} />
    </div>
  );
}
