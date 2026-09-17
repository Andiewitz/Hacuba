"use client";

import { Heart } from "lucide-react";
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

        <button
          className="absolute top-3 right-3 flex h-8 w-8 items-center justify-center rounded-full bg-card/80 text-foreground backdrop-blur-sm transition-colors hover:bg-card"
          aria-label="Save to favorites"
        >
          <Heart className="h-4 w-4" />
        </button>
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
