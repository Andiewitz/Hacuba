import ListingSection from "@/components/listing-section";
import SearchBar from "@/components/search-bar";
import { listingFiltersFromSearchParams } from "@/lib/listing-filters";
import { getPublicListings } from "@/lib/listings";

type HomeProps = {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
};

export default async function Home({ searchParams }: HomeProps) {
  const filters = listingFiltersFromSearchParams(await searchParams);
  const listings = await getPublicListings(filters);

  return (
    <main className="min-h-screen bg-background pb-16">
      <section className="bg-[var(--color-forest)] px-6 py-12 md:px-10 md:py-16 lg:px-20">
        <div className="mx-auto max-w-[1760px]">
          <p className="text-sm font-semibold text-[var(--text-secondary-on-dark)]">Cebu property marketplace</p>
          <h1 className="mt-3 max-w-2xl font-heading text-[2.75rem] font-bold leading-[1.15] tracking-[-0.015em] text-[var(--text-primary-on-dark)] md:text-[3.5rem] md:leading-[1.1] md:tracking-[-0.02em]">
            Find a place that fits the life you&apos;re building.
          </h1>
          <p className="mt-4 max-w-[62ch] text-[1.125rem] leading-[1.6] text-[var(--text-body-on-dark)]">
            Browse homes, lots, and commercial spaces for sale or rent across Cebu.
          </p>
          <div className="mt-8">
            <SearchBar />
          </div>
        </div>
      </section>

      <ListingSection title="Properties in Cebu" listings={listings} sectionIndex={0} />
    </main>
  );
}
