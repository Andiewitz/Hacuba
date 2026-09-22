"use client";

import PropertyCard from "@/components/property-card";
import { motion } from "framer-motion";
import type { ListingCardData } from "@/lib/listings";

interface ListingSectionProps {
  title: string;
  listings: ListingCardData[];
  sectionIndex?: number;
}

export default function ListingSection({
  title,
  listings,
  sectionIndex = 0,
}: ListingSectionProps) {
  return (
    <section className="px-6 py-8 md:px-10 lg:px-20">
      <div className="mx-auto max-w-[1760px]">
        <motion.div
          initial={{ opacity: 0, y: 12 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-50px" }}
          transition={{ duration: 0.4, delay: sectionIndex * 0.1 }}
          className="mb-6"
        >
          <h2 className="font-heading text-[2.25rem] font-semibold leading-[1.2] tracking-[-0.01em] text-foreground">
            {title}
          </h2>
        </motion.div>

        {listings.length === 0 ? (
          <div className="rounded-[var(--radius-lg)] border border-border bg-card px-6 py-12 text-center shadow-[var(--shadow-sm)]">
            <p className="font-heading text-[1.375rem] font-medium text-foreground">No properties match these filters</p>
            <p className="mt-2 text-sm leading-6 text-muted-foreground">Try a different location, price range, or property type.</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
            {listings.map((listing, index) => (
            <motion.div
              key={listing.id}
              initial={{ opacity: 0, y: 16 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, margin: "-30px" }}
              transition={{ duration: 0.3, delay: index * 0.05 }}
            >
              <PropertyCard
                {...listing}
                priority={sectionIndex === 0 && index === 0}
              />
            </motion.div>
            ))}
          </div>
        )}
      </div>
    </section>
  );
}
