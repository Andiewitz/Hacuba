import SearchBar from "@/components/search-bar";
import ListingSection from "@/components/listing-section";
import {
  homes,
  lots,
  commercial,
  getNewListings,
  getAvailableNextMonth,
  nearbyAreas,
} from "@/data/listings";

export default function Home() {
  const newListings = getNewListings();
  const nextMonth = getAvailableNextMonth();

  return (
    <div className="min-h-screen bg-background">

      <section className="px-6 pt-6 pb-4 md:px-10 lg:px-20">
        <SearchBar />
      </section>

      <ListingSection title="New listings" listings={newListings} sectionIndex={0} />
      <ListingSection title="Relevant homes in Cebu City" listings={homes} sectionIndex={1} />
      <ListingSection title="Relevant lots in Cebu City" listings={lots} sectionIndex={2} />
      <ListingSection title="Relevant commercial properties in Cebu City" listings={commercial} sectionIndex={3} />
      {nearbyAreas.map((area, i) => (
        <ListingSection
          key={area}
          title={`Available in ${area}`}
          listings={[...homes, ...lots, ...commercial].slice(i * 3, i * 3 + 6)}
          sectionIndex={i + 4}
        />
      ))}
      <ListingSection title="Available next month" listings={nextMonth} sectionIndex={nearbyAreas.length + 4} />
    </div>
  );
}
