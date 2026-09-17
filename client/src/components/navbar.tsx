"use client";

import { Globe, Menu } from "lucide-react";
import Link from "next/link";
import HacubaLogo from "@/components/hacuba-logo";

const navItems = [
  { label: "All", emoji: "🌎", active: true },
  { label: "Homes", emoji: "🏡", active: false },
  { label: "Lots", emoji: "🌳", active: false },
  { label: "Commercial", emoji: "🏢", active: false },
];

export default function Navbar() {
  return (
    <header className="sticky top-0 z-50 bg-background">
      <div className="relative mx-auto flex h-20 max-w-[1760px] items-center justify-between px-6 md:px-10 lg:px-20">
        <Link href="/" className="shrink-0">
          <HacubaLogo className="text-primary" />
        </Link>

        <nav className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 hidden md:flex items-center gap-1 rounded-full border border-border bg-card px-2 py-1.5 shadow-sm">
          {navItems.map((item) => (
            <button
              key={item.label}
              className={`flex items-center gap-2 rounded-full px-4 py-2 text-sm font-medium transition-colors ${
                item.active
                  ? "bg-foreground text-background"
                  : "text-foreground hover:bg-muted"
              }`}
            >
              <span className="text-base">{item.emoji}</span>
              <span>{item.label}</span>
            </button>
          ))}
        </nav>

        <div className="flex items-center gap-2">
          <button className="hidden lg:flex items-center gap-2 rounded-full border border-border px-4 py-2.5 text-sm font-medium text-foreground transition-colors hover:shadow-md">
            Become a seller
          </button>
          <button className="flex items-center justify-center rounded-full border border-border p-2.5 text-foreground transition-colors hover:shadow-md">
            <Globe className="h-4 w-4" />
          </button>
          <button className="flex items-center gap-3 rounded-full border border-border py-1.5 pl-3 pr-1.5 text-foreground transition-colors hover:shadow-md">
            <Menu className="h-4 w-4" />
            <div className="flex h-8 w-8 items-center justify-center rounded-full bg-muted text-muted-foreground">
              <svg
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
              >
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
                <circle cx="12" cy="7" r="4" />
              </svg>
            </div>
          </button>
        </div>
      </div>
    </header>
  );
}
