import ListingSection from "@/components/listing-section";
import SearchBar from "@/components/search-bar";
import { listingFiltersFromSearchParams } from "@/lib/listing-filters";
import { getPublicListings } from "@/lib/listings";

type LotsPageProps = {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
};

export default async function LotsPage({ searchParams }: LotsPageProps) {
  const filters = listingFiltersFromSearchParams(await searchParams);
  const listings = await getPublicListings({ ...filters, propertyType: "lot" });

  return (
    <main className="min-h-screen bg-background pb-16">
      <section className="px-6 pb-4 pt-6 md:px-10 lg:px-20">
        <SearchBar />
      </section>
      <ListingSection title="Lots in Cebu" listings={listings} sectionIndex={0} />
    </main>
  );
}
