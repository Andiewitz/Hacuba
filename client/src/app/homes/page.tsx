import ListingSection from "@/components/listing-section";
import SearchBar from "@/components/search-bar";
import { listingFiltersFromSearchParams } from "@/lib/listing-filters";
import { getPublicListings } from "@/lib/listings";

type HomesPageProps = {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
};

export default async function HomesPage({ searchParams }: HomesPageProps) {
  const filters = listingFiltersFromSearchParams(await searchParams);
  const listings = await getPublicListings({ ...filters, propertyType: filters.propertyType ?? "house" });

  return (
    <main className="min-h-screen bg-background pb-16">
      <section className="px-6 pb-4 pt-6 md:px-10 lg:px-20">
        <SearchBar />
      </section>
      <ListingSection title="Homes in Cebu" listings={listings} sectionIndex={0} />
    </main>
  );
}
