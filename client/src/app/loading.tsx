function LoadingBar({ className = "" }: { className?: string }) {
  return <div aria-hidden="true" className={`animate-pulse rounded-[var(--radius-sm)] bg-card ${className}`} />;
}

export default function Loading() {
  return (
    <main aria-busy="true" aria-label="Loading properties" className="min-h-screen bg-background pb-16">
      <section className="bg-[var(--color-forest)] px-6 py-12 md:px-10 md:py-16 lg:px-20">
        <div className="mx-auto max-w-[1760px]">
          <LoadingBar className="h-4 w-44 bg-[var(--color-forest-surface-2)]" />
          <div className="mt-4 space-y-3">
            <LoadingBar className="h-12 max-w-xl bg-[var(--color-forest-surface-2)] md:h-16" />
            <LoadingBar className="h-6 max-w-2xl bg-[var(--color-forest-surface-2)]" />
          </div>
          <div className="mt-8 flex h-[76px] max-w-3xl items-center gap-4 rounded-full border border-[var(--color-cream-border)] bg-white px-6 shadow-[var(--shadow-md)]">
            <LoadingBar className="h-8 flex-1" />
            <LoadingBar className="h-8 flex-1" />
            <LoadingBar className="h-8 flex-1" />
            <div aria-hidden="true" className="h-12 w-12 rounded-full bg-[var(--color-terracotta)]" />
          </div>
        </div>
      </section>

      <section className="px-6 py-12 md:px-10 lg:px-20">
        <div className="mx-auto max-w-[1760px]">
          <LoadingBar className="h-10 w-72" />
          <div className="mt-6 grid grid-cols-1 gap-6 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
            {Array.from({ length: 8 }, (_, index) => (
              <div key={index} className="space-y-3">
                <LoadingBar className="aspect-[4/3] w-full rounded-[var(--radius-lg)] shadow-[var(--shadow-sm)]" />
                <LoadingBar className="h-5 w-4/5" />
                <LoadingBar className="h-4 w-2/5" />
                <LoadingBar className="h-4 w-3/5" />
              </div>
            ))}
          </div>
        </div>
      </section>
    </main>
  );
}
