"use client";

import { useState, useRef, useEffect } from "react";
import { Search, MapPin } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { useRouter } from "next/navigation";

const destinations = [
  { name: "Cebu City, Philippines", description: "For sights like Magellan's Cross" },
  { name: "Mandaue, Philippines", description: "Near you" },
  { name: "Lapu-Lapu, Philippines", description: "For a beach getaway" },
  { name: "Mactan, Philippines", description: "For island hopping" },
  { name: "Bohol, Philippines", description: "For the Chocolate Hills" },
];

const priceRanges = [
  { label: "Any price" },
  { label: "Under ₱3M", maxPrice: 300_000_000 },
  { label: "₱3M – ₱6M", minPrice: 300_000_000, maxPrice: 600_000_000 },
  { label: "₱6M – ₱10M", minPrice: 600_000_000, maxPrice: 1_000_000_000 },
  { label: "₱10M+", minPrice: 1_000_000_000 },
];

const propertyTypes = [
  "Any type",
  "House",
  "Apartment",
  "Condo",
  "Lot",
  "Commercial",
];

export default function SearchBar() {
  const router = useRouter();
  const [activeSegment, setActiveSegment] = useState<string | null>(null);
  const [where, setWhere] = useState("");
  const [price, setPrice] = useState(priceRanges[0].label);
  const [propertyType, setPropertyType] = useState("Any type");
  const [searchQuery, setSearchQuery] = useState("");
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setActiveSegment(null);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  useEffect(() => {
    if (!activeSegment) return;
    const el = containerRef.current;
    if (!el) return;

    function handleWheel(e: WheelEvent) {
      const dropdown = el?.querySelector("[data-dropdown]");
      if (dropdown && dropdown.contains(e.target as Node)) {
        e.stopPropagation();
      }
    }

    document.addEventListener("wheel", handleWheel, { passive: false, capture: true });
    return () => document.removeEventListener("wheel", handleWheel, { capture: true });
  }, [activeSegment]);

  const filteredDestinations = destinations.filter((d) =>
    d.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const dropdownVariants = {
    hidden: { opacity: 0, y: -8, scale: 0.96 },
    visible: { opacity: 1, y: 0, scale: 1, transition: { duration: 0.18, ease: "easeOut" as const } },
    exit: { opacity: 0, y: -6, scale: 0.97, transition: { duration: 0.12, ease: "easeIn" as const } },
  };

  const applyFilters = () => {
    const params = new URLSearchParams();
    const selectedPrice = priceRanges.find((range) => range.label === price);

    if (where) params.set("city", where.split(",")[0]);
    if (propertyType !== "Any type") params.set("type", propertyType.toLowerCase());
    if (selectedPrice?.minPrice !== undefined) params.set("min_price", String(selectedPrice.minPrice));
    if (selectedPrice?.maxPrice !== undefined) params.set("max_price", String(selectedPrice.maxPrice));

    router.push(`/?${params.toString()}`);
    setActiveSegment(null);
  };

  return (
    <div ref={containerRef} className="relative mx-auto w-full max-w-3xl">
      <div className="relative flex items-center rounded-full border border-border bg-muted/50 shadow-md transition-shadow hover:shadow-lg">
        {/* Where */}
        <button
          onClick={() => setActiveSegment(activeSegment === "where" ? null : "where")}
          className="relative flex-1 rounded-l-full px-6 py-4 text-left transition-colors hover:bg-muted/30"
        >
          {activeSegment === "where" && (
            <motion.div
              layoutId="segment-pill"
              className="absolute inset-0 rounded-full bg-card shadow-[var(--shadow-md)]"
              transition={{ type: "spring", stiffness: 400, damping: 30 }}
            />
          )}
          <span className="relative block text-xs font-semibold text-foreground">Where</span>
          <span className="relative block text-sm text-muted-foreground">
            {where || "Search locations"}
          </span>
        </button>

        <div className="relative z-10 h-8 w-px bg-border/60" />

        {/* Price */}
        <button
          onClick={() => setActiveSegment(activeSegment === "price" ? null : "price")}
          className="relative flex-1 px-6 py-4 text-left transition-colors hover:bg-muted/30"
        >
          {activeSegment === "price" && (
            <motion.div
              layoutId="segment-pill"
              className="absolute inset-0 rounded-full bg-card shadow-[var(--shadow-md)]"
              transition={{ type: "spring", stiffness: 400, damping: 30 }}
            />
          )}
          <span className="relative block text-xs font-semibold text-foreground">Price</span>
          <span className="relative block text-sm text-muted-foreground">{price}</span>
        </button>

        <div className="relative z-10 h-8 w-px bg-border/60" />

        {/* Property Type */}
        <button
          onClick={() => setActiveSegment(activeSegment === "type" ? null : "type")}
          className="relative flex-1 px-6 py-4 text-left transition-colors hover:bg-muted/30"
        >
          {activeSegment === "type" && (
            <motion.div
              layoutId="segment-pill"
              className="absolute inset-0 rounded-full bg-card shadow-[var(--shadow-md)]"
              transition={{ type: "spring", stiffness: 400, damping: 30 }}
            />
          )}
          <span className="relative block text-xs font-semibold text-foreground">Property Type</span>
          <span className="relative block text-sm text-muted-foreground">{propertyType}</span>
        </button>

        <button
          type="button"
          onClick={applyFilters}
          aria-label="Search properties"
          className="relative z-10 mr-2 flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-primary text-[var(--color-forest)] transition-colors hover:bg-[var(--color-terracotta-hover)]"
        >
          <Search className="h-4 w-4" />
        </button>
      </div>

      <AnimatePresence>
        {/* Where Dropdown */}
        {activeSegment === "where" && (
          <motion.div
            key="where"
            variants={dropdownVariants}
            initial="hidden"
            animate="visible"
            exit="exit"
            className="absolute top-full left-0 z-50 mt-3 w-80 rounded-2xl border border-border bg-card p-4 shadow-lg overflow-y-auto max-h-80"
            data-dropdown
          >
            <p className="mb-3 text-xs font-semibold text-muted-foreground">
              Suggested locations
            </p>
            <input
              type="text"
              placeholder="Search locations"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="mb-3 w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
            />
            <div className="flex flex-col gap-1">
              {filteredDestinations.map((dest, i) => (
                <motion.button
                  key={dest.name}
                  initial={{ opacity: 0, x: -8 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: i * 0.04, duration: 0.2 }}
                  onClick={() => {
                    setWhere(dest.name);
                    setActiveSegment(null);
                    setSearchQuery("");
                  }}
                  className="relative flex items-center gap-3 rounded-xl px-3 py-3 text-left transition-colors hover:bg-muted"
                >
                  {where === dest.name && (
                    <motion.div
                      layoutId="where-highlight"
                      className="absolute inset-0 rounded-xl bg-accent/20"
                      transition={{ type: "spring", stiffness: 500, damping: 30 }}
                    />
                  )}
                  <div className="relative flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-muted">
                    <MapPin className="h-5 w-5 text-muted-foreground" />
                  </div>
                  <div className="relative">
                    <p className="text-sm font-medium text-foreground">{dest.name}</p>
                    <p className="text-xs text-muted-foreground">{dest.description}</p>
                  </div>
                </motion.button>
              ))}
            </div>
          </motion.div>
        )}

        {/* Price Dropdown */}
        {activeSegment === "price" && (
          <motion.div
            key="price"
            variants={dropdownVariants}
            initial="hidden"
            animate="visible"
            exit="exit"
            className="absolute top-full left-1/3 z-50 mt-3 w-64 rounded-2xl border border-border bg-card p-2 shadow-lg"
            data-dropdown
          >
            {priceRanges.map((range, i) => (
              <motion.button
                key={range.label}
                initial={{ opacity: 0, y: -4 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: i * 0.03, duration: 0.15 }}
                onClick={() => {
                  setPrice(range.label);
                  setActiveSegment(null);
                }}
                className="relative flex w-full items-center rounded-xl px-4 py-3 text-sm text-left transition-colors hover:bg-muted"
              >
                {price === range.label && (
                  <motion.div
                    layoutId="price-highlight"
                    className="absolute inset-0 rounded-xl bg-accent/20"
                    transition={{ type: "spring", stiffness: 500, damping: 30 }}
                  />
                )}
                <span
                  className={`relative ${
                    price === range.label ? "font-semibold text-foreground" : "text-muted-foreground"
                  }`}
                >
                  {range.label}
                </span>
              </motion.button>
            ))}
          </motion.div>
        )}

        {/* Property Type Dropdown */}
        {activeSegment === "type" && (
          <motion.div
            key="type"
            variants={dropdownVariants}
            initial="hidden"
            animate="visible"
            exit="exit"
            className="absolute top-full right-16 z-50 mt-3 w-56 rounded-2xl border border-border bg-card p-2 shadow-lg"
            data-dropdown
          >
            {propertyTypes.map((type, i) => (
              <motion.button
                key={type}
                initial={{ opacity: 0, y: -4 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: i * 0.03, duration: 0.15 }}
                onClick={() => {
                  setPropertyType(type);
                  setActiveSegment(null);
                }}
                className="relative flex w-full items-center rounded-xl px-4 py-3 text-sm text-left transition-colors hover:bg-muted"
              >
                {propertyType === type && (
                  <motion.div
                    layoutId="type-highlight"
                    className="absolute inset-0 rounded-xl bg-accent/20"
                    transition={{ type: "spring", stiffness: 500, damping: 30 }}
                  />
                )}
                <span
                  className={`relative ${
                    propertyType === type ? "font-semibold text-foreground" : "text-muted-foreground"
                  }`}
                >
                  {type}
                </span>
              </motion.button>
            ))}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
