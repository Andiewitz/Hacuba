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

// Save-burst sparks: fixed radiating angles so the burst is identical on
// every card and every save (deterministic, SSR-safe — no Math.random).
const SPARKS = Array.from({ length: 8 }, (_, i) => {
  const angle = ((i * 45 + 22) * Math.PI) / 180;
  const distance = 24 + (i % 3) * 4;
  return {
    x: Math.cos(angle) * distance,
    y: Math.sin(angle) * distance,
    size: [4, 6, 3][i % 3],
    tone: ["bg-amber-300", "bg-amber-400", "bg-amber-500"][i % 3],
  };
});

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
          whileTap={{ scale: 0.8 }}
          transition={{ duration: 0.12 }}
          className="absolute top-3 right-3 flex h-8 w-8 items-center justify-center text-white drop-shadow-[0_0_2px_rgba(0,0,0,0.9)] drop-shadow-[0_1px_4px_rgba(0,0,0,0.7)]"
        >
          <AnimatePresence>
            {!reduce && splashKey > 0 && (
              <motion.span
                key={splashKey}
                aria-hidden="true"
                className="pointer-events-none absolute inset-0"
              >
                <motion.span
                  aria-hidden="true"
                  className="absolute inset-[-6px] rounded-full bg-amber-300/60 blur-md"
                  initial={{ opacity: 0.7, scale: 0.5 }}
                  animate={{ opacity: 0, scale: 1.4 }}
                  transition={{ duration: 0.4, ease: [0.23, 1, 0.32, 1] }}
                />
                {SPARKS.map((spark, i) => (
                  <motion.span
                    key={i}
                    aria-hidden="true"
                    className={`absolute top-1/2 left-1/2 rounded-full ${spark.tone}`}
                    style={{
                      width: spark.size,
                      height: spark.size,
                      marginLeft: -spark.size / 2,
                      marginTop: -spark.size / 2,
                    }}
                    initial={{ opacity: 1, scale: 1, x: 0, y: 0 }}
                    animate={{ opacity: 0, scale: 0.2, x: spark.x, y: spark.y }}
                    transition={{
                      duration: 0.45,
                      delay: i * 0.02,
                      ease: [0.23, 1, 0.32, 1],
                    }}
                  />
                ))}
              </motion.span>
            )}
          </AnimatePresence>
          <motion.span
            key={saved ? "saved" : "unsaved"}
            aria-hidden="true"
            className="flex"
            initial={saved ? { scale: 0.5, rotate: -12 } : false}
            animate={{ scale: 1, rotate: 0 }}
            transition={{ type: "spring", stiffness: 500, damping: 15 }}
          >
            <Bookmark
              strokeWidth={2.5}
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
