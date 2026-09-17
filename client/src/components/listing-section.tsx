"use client";

import PropertyCard from "@/components/property-card";
import { ArrowRight } from "lucide-react";
import { motion } from "framer-motion";
import type { Listing } from "@/data/listings";

interface ListingSectionProps {
  title: string;
  listings: Listing[];
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
          className="flex items-center gap-3 mb-6"
        >
          <h2 className="text-2xl font-semibold text-foreground font-heading">
            {title}
          </h2>
          <ArrowRight className="h-5 w-5 text-foreground" />
        </motion.div>

        <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6">
          {listings.map((listing, index) => (
            <motion.div
              key={`${listing.type}-${listing.location}-${sectionIndex}-${index}`}
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
      </div>
    </section>
  );
}
