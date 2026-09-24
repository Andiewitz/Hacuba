export default function ListingLoading() {
  return (
    <main className="min-h-screen animate-pulse bg-background">
      <section className="border-b border-[var(--color-forest-border-subtle)] bg-[var(--color-forest)] px-6 py-4 md:px-10 lg:px-20">
        <div className="mx-auto max-w-[1200px]">
          <div className="h-4 w-36 rounded-[var(--radius-sm)] bg-[var(--color-forest-surface-2)]" />
        </div>
      </section>
      <div className="mx-auto max-w-[1440px] px-6 pt-6 md:px-10 md:pt-8 lg:px-20">
        <div className="aspect-[16/10] rounded-[var(--radius-lg)] bg-[var(--color-forest-surface-1)] shadow-[var(--shadow-lg)]" />
      </div>
      <div className="mx-auto grid max-w-[1440px] gap-8 px-6 py-10 md:px-10 lg:grid-cols-[minmax(0,1fr)_360px] lg:px-20 lg:py-12">
        <div>
          <div className="h-5 w-44 rounded-[var(--radius-sm)] bg-card" />
          <div className="mt-4 h-16 max-w-2xl rounded-[var(--radius-md)] bg-card" />
          <div className="mt-4 h-20 max-w-3xl rounded-[var(--radius-md)] bg-card" />
          <div className="mt-10 grid gap-4 border-y border-border py-6 sm:grid-cols-3">
            {[0, 1, 2].map((item) => <div key={item} className="h-16 rounded-[var(--radius-sm)] bg-card" />)}
          </div>
          <div className="mt-10 h-8 w-56 rounded-[var(--radius-sm)] bg-card" />
          <div className="mt-4 h-20 max-w-3xl rounded-[var(--radius-md)] bg-card" />
        </div>
        <div className="h-52 rounded-[var(--radius-lg)] bg-card shadow-[var(--shadow-md)]" />
      </div>
    </main>
  );
}
