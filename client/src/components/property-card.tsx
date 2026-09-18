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

// Save splash: water-crown droplets arcing up and out with gravity, like a
// liquid burst. Fixed angles and distances so the splash is identical on
// every card and every save (deterministic, SSR-safe — no Math.random).
// Angles use screen-clockwise degrees (0° = east, -90° = straight up).
const DROPLETS = [
  { angle: -165, dist: 22, size: 5, tone: "bg-amber-400" },
  { angle: -140, dist: 28, size: 7, tone: "bg-amber-300" },
  { angle: -118, dist: 32, size: 6, tone: "bg-amber-500" },
  { angle: -96, dist: 34, size: 9, tone: "bg-amber-400" },
  { angle: -84, dist: 34, size: 7, tone: "bg-white" },
  { angle: -62, dist: 32, size: 5, tone: "bg-amber-300" },
  { angle: -40, dist: 28, size: 8, tone: "bg-amber-500" },
  { angle: -15, dist: 22, size: 6, tone: "bg-amber-400" },
  { angle: 155, dist: 16, size: 5, tone: "bg-white" },
  { angle: 25, dist: 16, size: 5, tone: "bg-amber-300" },
].map((d, i) => {
  const rad = (d.angle * Math.PI) / 180;
  return {
    ...d,
    x: Math.cos(rad) * d.dist,
    rise: -(22 + (i % 3) * 6),
    fall: 8 + (i % 2) * 5,
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
                {DROPLETS.map((drop, i) => (
                  <motion.span
                    key={i}
                    aria-hidden="true"
                    className="absolute top-1/2 left-1/2"
                    style={{
                      width: drop.size,
                      height: drop.size,
                      marginLeft: -drop.size / 2,
                      marginTop: -drop.size / 2,
                    }}
                    initial={{ opacity: 1, scale: 1, x: 0, y: 0 }}
                    animate={{
                      opacity: [1, 1, 0],
                      scale: [1, 0.9, 0.25],
                      x: [0, drop.x * 0.75, drop.x],
                      y: [0, drop.rise, drop.fall],
                    }}
                    transition={{
                      duration: 0.6,
                      delay: i * 0.025,
                      times: [0, 0.45, 1],
                      ease: [0.23, 1, 0.32, 1],
                    }}
                  >
                    <span
                      aria-hidden="true"
                      className={`block h-full w-full ${drop.tone}`}
                      style={{
                        borderRadius: "0 50% 50% 50%",
                        transform: `rotate(${drop.angle + 45}deg)`,
                      }}
                    />
                  </motion.span>
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
