"use client";

import { Globe, Menu } from "lucide-react";
import Link from "next/link";

const navItems = [
  { label: "All", emoji: "🌎", active: true },
  { label: "Homes", emoji: "🏡", active: false },
  { label: "Experiences", emoji: "🎈", active: false },
  { label: "Services", emoji: "🛎️", active: false },
];

export default function Navbar() {
  return (
    <header className="sticky top-0 z-50 bg-background border-b border-border">
      <div className="mx-auto flex h-20 max-w-[1760px] items-center justify-between px-6 md:px-10 lg:px-20">
        <Link href="/" className="flex items-center gap-1 shrink-0">
          <svg
            width="32"
            height="32"
            viewBox="0 0 32 32"
            fill="none"
            className="text-primary"
          >
            <path
              d="M16 2C8.268 2 2 8.268 2 16s6.268 14 14 14 14-6.268 14-14S23.732 2 16 2zm0 2c2.8 0 5.36 1.04 7.32 2.76L16 16V4zm-9.32 4.76A11.94 11.94 0 0116 4v12L6.68 6.76zM16 28a11.94 11.94 0 01-9.32-4.24L16 16v12zm2-12l9.32-7.24A11.94 11.94 0 0116 28V16z"
              fill="currentColor"
            />
          </svg>
          <span className="text-xl font-semibold tracking-tight text-foreground font-heading">
            Paruba
          </span>
        </Link>

        <nav className="hidden md:flex items-center gap-1 rounded-full border border-border bg-card px-2 py-1.5 shadow-sm">
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
            Become a host
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
