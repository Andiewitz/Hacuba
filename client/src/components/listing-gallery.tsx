"use client";

import { ChevronLeft, ChevronRight } from "lucide-react";
import Image from "next/image";
import { useState } from "react";

type ListingGalleryProps = {
  images: string[];
  title: string;
  mode: "for_sale" | "for_rent";
};

export default function ListingGallery({ images, title, mode }: ListingGalleryProps) {
  const [activeIndex, setActiveIndex] = useState(0);
  const hasMultipleImages = images.length > 1;
  const currentImage = images[activeIndex];

  const move = (direction: -1 | 1) => {
    setActiveIndex((current) => (current + direction + images.length) % images.length);
  };

  if (!currentImage) {
    return (
      <section aria-label="Property gallery" className="flex aspect-[16/10] items-center justify-center overflow-hidden rounded-[var(--radius-lg)] bg-[var(--color-forest-surface-1)] px-6 text-center shadow-[var(--shadow-lg)]">
        <p className="max-w-[36ch] text-sm leading-[1.5] text-[var(--text-body-on-dark)]">No public photos have been added to this listing yet.</p>
      </section>
    );
  }

  return (
    <section
      aria-label="Property gallery"
      aria-roledescription="carousel"
      tabIndex={0}
      onKeyDown={(event) => {
        if (!hasMultipleImages) return;
        if (event.key === "ArrowLeft") move(-1);
        if (event.key === "ArrowRight") move(1);
      }}
      className="group relative aspect-[16/10] overflow-hidden rounded-[var(--radius-lg)] bg-[var(--color-forest-surface-1)] shadow-[var(--shadow-lg)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-4 focus-visible:ring-offset-background"
    >
      <Image src={currentImage} alt={`${title}, photo ${activeIndex + 1}`} fill priority={activeIndex === 0} unoptimized className="object-cover" />
      <span className="absolute left-4 top-4 rounded-full bg-[var(--color-sage)] px-3 py-1 text-xs font-semibold text-[var(--color-forest)]">
        {mode === "for_rent" ? "For rent" : "For sale"}
      </span>
      <p aria-live="polite" className="absolute bottom-4 right-4 rounded-full bg-[var(--color-forest)] px-3 py-1 text-sm font-semibold text-[var(--text-primary-on-dark)]">
        {activeIndex + 1} / {images.length}
      </p>
      {hasMultipleImages && (
        <div className="absolute inset-x-4 top-1/2 flex -translate-y-1/2 items-center justify-between">
          <button
            type="button"
            onClick={() => move(-1)}
            className="flex h-10 w-10 items-center justify-center rounded-full border border-[var(--color-forest-border-strong)] bg-[var(--color-forest)] text-[var(--text-primary-on-dark)] shadow-[var(--shadow-sm)] transition-colors hover:bg-[var(--color-forest-surface-1)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-4 focus-visible:ring-offset-[var(--color-forest)]"
            aria-label="Previous photo"
          >
            <ChevronLeft className="h-5 w-5" aria-hidden="true" />
          </button>
          <button
            type="button"
            onClick={() => move(1)}
            className="flex h-10 w-10 items-center justify-center rounded-full border border-[var(--color-forest-border-strong)] bg-[var(--color-forest)] text-[var(--text-primary-on-dark)] shadow-[var(--shadow-sm)] transition-colors hover:bg-[var(--color-forest-surface-1)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-4 focus-visible:ring-offset-[var(--color-forest)]"
            aria-label="Next photo"
          >
            <ChevronRight className="h-5 w-5" aria-hidden="true" />
          </button>
        </div>
      )}
    </section>
  );
}
