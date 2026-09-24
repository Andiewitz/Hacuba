export default function ListingLoading() {
  return (
    <main className="min-h-screen animate-pulse bg-background">
      <section className="bg-[var(--color-forest)] px-6 py-10 md:px-10 md:py-12 lg:px-20">
        <div className="mx-auto max-w-[1200px]">
          <div className="h-4 w-36 rounded-[var(--radius-sm)] bg-[var(--color-forest-surface-2)]" />
          <div className="mt-8 h-6 w-20 rounded-full bg-[var(--color-forest-surface-2)]" />
          <div className="mt-4 h-12 max-w-2xl rounded-[var(--radius-md)] bg-[var(--color-forest-surface-2)] md:h-16" />
          <div className="mt-4 h-5 w-48 rounded-[var(--radius-sm)] bg-[var(--color-forest-surface-2)]" />
        </div>
      </section>
      <div className="mx-auto grid max-w-[1200px] gap-12 px-6 py-12 md:px-10 lg:grid-cols-[minmax(0,1fr)_320px] lg:px-20 lg:py-16">
        <div>
          <div className="aspect-[4/3] rounded-[var(--radius-lg)] bg-card shadow-[var(--shadow-md)]" />
          <div className="mt-12 h-8 w-64 rounded-[var(--radius-sm)] bg-card" />
          <div className="mt-4 h-20 max-w-3xl rounded-[var(--radius-md)] bg-card" />
          <div className="mt-12 grid gap-4 sm:grid-cols-3">
            {[0, 1, 2].map((item) => <div key={item} className="h-24 rounded-[var(--radius-md)] bg-card shadow-[var(--shadow-sm)]" />)}
          </div>
        </div>
        <div className="h-64 rounded-[var(--radius-lg)] bg-[var(--color-forest-surface-1)] shadow-[var(--shadow-md)]" />
      </div>
    </main>
  );
}
