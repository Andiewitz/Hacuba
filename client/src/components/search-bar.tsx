"use client";

import { Search } from "lucide-react";

export default function SearchBar() {
  return (
    <div className="mx-auto w-full max-w-3xl">
      <div className="flex items-center rounded-full border border-border bg-card shadow-md transition-shadow hover:shadow-lg">
        <button className="flex-1 rounded-l-full px-6 py-4 text-left transition-colors hover:bg-muted">
          <span className="block text-xs font-semibold text-foreground">
            Where
          </span>
          <span className="block text-sm text-muted-foreground">
            Search destinations
          </span>
        </button>

        <div className="h-8 w-px bg-border" />

        <button className="flex-1 px-6 py-4 text-left transition-colors hover:bg-muted">
          <span className="block text-xs font-semibold text-foreground">
            When
          </span>
          <span className="block text-sm text-muted-foreground">
            Add dates
          </span>
        </button>

        <div className="h-8 w-px bg-border" />

        <button className="flex-1 px-6 py-4 text-left transition-colors hover:bg-muted">
          <span className="block text-xs font-semibold text-foreground">
            Who
          </span>
          <span className="block text-sm text-muted-foreground">
            Add guests
          </span>
        </button>

        <button className="mr-2 flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-primary text-primary-foreground transition-colors hover:bg-primary/90">
          <Search className="h-4 w-4" />
        </button>
      </div>
    </div>
  );
}
