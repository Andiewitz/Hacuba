"use client";

import { useState, useRef, useEffect } from "react";
import { Search, MapPin } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";

const destinations = [
  { name: "Cebu City, Philippines", description: "For sights like Magellan's Cross" },
  { name: "Mandaue, Philippines", description: "Near you" },
  { name: "Lapu-Lapu, Philippines", description: "For a beach getaway" },
  { name: "Mactan, Philippines", description: "For island hopping" },
  { name: "Bohol, Philippines", description: "For the Chocolate Hills" },
];

const priceRanges = [
  "Any price",
  "₱1,000 – ₱2,500",
  "₱2,500 – ₱5,000",
  "₱5,000 – ₱10,000",
  "₱10,000+",
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
  const [activeDropdown, setActiveDropdown] = useState<string | null>(null);
  const [where, setWhere] = useState("");
  const [price, setPrice] = useState(priceRanges[0]);
  const [propertyType, setPropertyType] = useState("Any type");
  const [searchQuery, setSearchQuery] = useState("");
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setActiveDropdown(null);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const filteredDestinations = destinations.filter((d) =>
    d.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const dropdownVariants = {
    hidden: { opacity: 0, y: -8, scale: 0.96 },
    visible: { opacity: 1, y: 0, scale: 1, transition: { duration: 0.18, ease: "easeOut" as const } },
    exit: { opacity: 0, y: -6, scale: 0.97, transition: { duration: 0.12, ease: "easeIn" as const } },
  };

  return (
    <div ref={containerRef} className="relative mx-auto w-full max-w-3xl">
      <div className="flex items-center rounded-full border border-border bg-card shadow-md transition-shadow hover:shadow-lg">
        {/* Where */}
        <button
          onClick={() => setActiveDropdown(activeDropdown === "where" ? null : "where")}
          className="flex-1 rounded-l-full px-6 py-4 text-left transition-colors hover:bg-muted"
        >
          <span className="block text-xs font-semibold text-foreground">Where</span>
          <span className="block text-sm text-muted-foreground">
            {where || "Search destinations"}
          </span>
        </button>

        <div className="h-8 w-px bg-border" />

        {/* Price */}
        <button
          onClick={() => setActiveDropdown(activeDropdown === "price" ? null : "price")}
          className="flex-1 px-6 py-4 text-left transition-colors hover:bg-muted"
        >
          <span className="block text-xs font-semibold text-foreground">Price</span>
          <span className="block text-sm text-muted-foreground">{price}</span>
        </button>

        <div className="h-8 w-px bg-border" />

        {/* Property Type */}
        <button
          onClick={() => setActiveDropdown(activeDropdown === "type" ? null : "type")}
          className="flex-1 px-6 py-4 text-left transition-colors hover:bg-muted"
        >
          <span className="block text-xs font-semibold text-foreground">Property Type</span>
          <span className="block text-sm text-muted-foreground">{propertyType}</span>
        </button>

        <button className="mr-2 flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-primary text-primary-foreground transition-colors hover:bg-primary/90">
          <Search className="h-4 w-4" />
        </button>
      </div>

      <AnimatePresence>
        {/* Where Dropdown */}
        {activeDropdown === "where" && (
          <motion.div
            key="where"
            variants={dropdownVariants}
            initial="hidden"
            animate="visible"
            exit="exit"
            className="absolute top-full left-0 z-50 mt-3 w-80 rounded-2xl border border-border bg-card p-4 shadow-lg"
          >
            <p className="mb-3 text-xs font-semibold text-muted-foreground">
              Suggested destinations
            </p>
            <input
              type="text"
              placeholder="Search destinations"
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
                    setActiveDropdown(null);
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
        {activeDropdown === "price" && (
          <motion.div
            key="price"
            variants={dropdownVariants}
            initial="hidden"
            animate="visible"
            exit="exit"
            className="absolute top-full left-1/3 z-50 mt-3 w-64 rounded-2xl border border-border bg-card p-2 shadow-lg"
          >
            {priceRanges.map((range, i) => (
              <motion.button
                key={range}
                initial={{ opacity: 0, y: -4 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: i * 0.03, duration: 0.15 }}
                onClick={() => {
                  setPrice(range);
                  setActiveDropdown(null);
                }}
                className="relative flex w-full items-center rounded-xl px-4 py-3 text-sm text-left transition-colors hover:bg-muted"
              >
                {price === range && (
                  <motion.div
                    layoutId="price-highlight"
                    className="absolute inset-0 rounded-xl bg-accent/20"
                    transition={{ type: "spring", stiffness: 500, damping: 30 }}
                  />
                )}
                <span
                  className={`relative ${
                    price === range ? "font-semibold text-foreground" : "text-muted-foreground"
                  }`}
                >
                  {range}
                </span>
              </motion.button>
            ))}
          </motion.div>
        )}

        {/* Property Type Dropdown */}
        {activeDropdown === "type" && (
          <motion.div
            key="type"
            variants={dropdownVariants}
            initial="hidden"
            animate="visible"
            exit="exit"
            className="absolute top-full right-16 z-50 mt-3 w-56 rounded-2xl border border-border bg-card p-2 shadow-lg"
          >
            {propertyTypes.map((type, i) => (
              <motion.button
                key={type}
                initial={{ opacity: 0, y: -4 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: i * 0.03, duration: 0.15 }}
                onClick={() => {
                  setPropertyType(type);
                  setActiveDropdown(null);
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
