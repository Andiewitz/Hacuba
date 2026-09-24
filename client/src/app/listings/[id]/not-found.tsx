import Link from "next/link";

export default function ListingNotFound() {
  return (
    <main className="min-h-screen bg-background px-6 py-16 md:px-10 lg:px-20">
      <section className="mx-auto max-w-[760px] rounded-[var(--radius-lg)] bg-[var(--color-forest)] p-8 shadow-[var(--shadow-lg)] md:p-12">
        <p className="text-sm font-semibold text-[var(--text-secondary-on-dark)]">Listing unavailable</p>
        <h1 className="mt-3 font-heading text-[2.75rem] font-bold leading-[1.15] tracking-[-0.015em] text-[var(--text-primary-on-dark)]">This property is no longer available.</h1>
        <p className="mt-4 max-w-[54ch] text-base leading-[1.6] text-[var(--text-body-on-dark)]">It may have been removed, unpublished, or the link may be incorrect.</p>
        <Link href="/" className="mt-8 inline-flex rounded-[var(--radius-md)] bg-[var(--color-terracotta)] px-4 py-3 text-sm font-semibold text-black focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-4 focus-visible:ring-offset-[var(--color-forest)]">Browse properties</Link>
      </section>
    </main>
  );
}
