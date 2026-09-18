"use client";

import { useState } from "react";
import { AnimatePresence, motion, useReducedMotion } from "framer-motion";
import { Bookmark } from "lucide-react";
import Image from "next/image";

interface PropertyCardProps {
  image: string;
  type: string;
  location: string;
  price: number;
  nights: number;
  rating: number;
  isGuestFavorite?: boolean;
  priority?: boolean;
}

export default function PropertyCard({
  image,
  type,
  location,
  price,
  nights,
  rating,
  isGuestFavorite = false,
  priority = false,
}: PropertyCardProps) {
  const [saved, setSaved] = useState(false);
  const [splashKey, setSplashKey] = useState(0);
  const reduce = useReducedMotion();

  const toggleSaved = () => {
    if (!saved) setSplashKey((k) => k + 1);
    setSaved((v) => !v);
  };

  return (
    <div className="group flex flex-col gap-2">
      <div className="relative aspect-square overflow-hidden rounded-[14px]">
        <Image
          src={image}
          alt={`${type} in ${location}`}
          fill
          unoptimized
          priority={priority}
          className="object-cover transition-transform duration-300 group-hover:scale-105"
        />

        {isGuestFavorite && (
          <span className="absolute top-3 left-3 rounded-full bg-card px-3 py-1 text-xs font-semibold text-foreground shadow-sm">
            Guest favorite
          </span>
        )}

        <motion.button
          onClick={toggleSaved}
          aria-pressed={saved}
          aria-label={saved ? "Remove from saved homes" : "Save this home"}
          whileTap={{ transform: "scale(0.8)" }}
          transition={{ duration: 0.12 }}
          className="absolute top-3 right-3 flex h-8 w-8 items-center justify-center text-white drop-shadow-[0_0_2px_rgba(0,0,0,0.9)] drop-shadow-[0_1px_4px_rgba(0,0,0,0.7)]"
        >
          <AnimatePresence>
            {!reduce && splashKey > 0 && (
              <motion.span
                key={splashKey}
                aria-hidden="true"
                className="pointer-events-none absolute inset-0 rounded-full border-[3px] border-amber-300"
                initial={{ opacity: 0.9, transform: "scale(0.5)" }}
                animate={{ opacity: 0, transform: "scale(2.2)" }}
                transition={{ duration: 0.4, ease: [0.23, 1, 0.32, 1] }}
              />
            )}
          </AnimatePresence>
          <motion.span
            key={saved ? "saved" : "unsaved"}
            aria-hidden="true"
            className="flex"
            initial={saved ? { transform: "scale(0.6)" } : false}
            animate={{ transform: "scale(1)" }}
            transition={{ type: "spring", stiffness: 500, damping: 18 }}
          >
            <Bookmark
              className={`h-5 w-5 transition-colors duration-150 ${
                saved ? "fill-amber-400 text-amber-400" : "text-white"
              }`}
            />
          </motion.span>
        </motion.button>
      </div>

      <div className="flex flex-col gap-0.5 px-0.5">
        <div className="flex items-start justify-between">
          <h3 className="text-sm font-medium text-foreground">
            {type} in {location}
          </h3>
          <span className="flex items-center gap-1 text-sm text-foreground">
            <span>★</span>
            <span>{rating}</span>
          </span>
        </div>
        <p className="text-sm text-muted-foreground">
          ₱{price.toLocaleString()} for {nights} nights
        </p>
      </div>
    </div>
  );
}
