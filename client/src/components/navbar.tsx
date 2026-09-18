"use client";

import { useEffect, useRef, useState } from "react";
import { Globe, LifeBuoy, LogIn, LogOut, Menu } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { AnimatePresence, motion } from "framer-motion";
import HacubaLogo from "@/components/hacuba-logo";
import AuthDialog from "@/components/auth-dialog";
import { useAppDispatch, useAppSelector } from "@/lib/hooks";
import { loggedOut } from "@/lib/slices/auth";

const navItems = [
  { label: "All", emoji: "🌎", href: "/" },
  { label: "Homes", emoji: "🏡", href: "/homes" },
  { label: "Lots", emoji: "🌳", href: "/lots" },
  { label: "Commercial", emoji: "🏢", href: "/commercial" },
];

const GITHUB_URL = "https://github.com/Andiewitz/Hacuba";

function GitHubMark() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="currentColor"
      aria-hidden="true"
    >
      <path d="M12 .5C5.7.5.5 5.7.5 12c0 5.1 3.3 9.4 7.9 10.9.6.1.8-.2.8-.6v-2c-3.2.7-3.9-1.4-3.9-1.4-.5-1.3-1.3-1.7-1.3-1.7-1-.7.1-.7.1-.7 1.2.1 1.8 1.2 1.8 1.2 1 1.8 2.7 1.3 3.4 1 .1-.8.4-1.3.7-1.6-2.6-.3-5.3-1.3-5.3-5.7 0-1.3.4-2.3 1.2-3.1-.1-.3-.5-1.5.1-3.1 0 0 1-.3 3.2 1.2a11 11 0 0 1 5.8 0C16.4 4.9 17.4 5.2 17.4 5.2c.6 1.6.2 2.8.1 3.1.7.8 1.2 1.8 1.2 3.1 0 4.4-2.7 5.4-5.3 5.7.4.4.8 1.1.8 2.2v3.2c0 .4.2.7.8.6 4.6-1.5 7.9-5.8 7.9-10.9C23.5 5.7 18.3.5 12 .5z" />
    </svg>
  );
}

export default function Navbar() {
  const pathname = usePathname();
  const dispatch = useAppDispatch();
  const userEmail = useAppSelector((s) => s.auth.userEmail);
  const [menuOpen, setMenuOpen] = useState(false);
  const [authOpen, setAuthOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!menuOpen) return;
    const onPointerDown = (e: PointerEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false);
      }
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setMenuOpen(false);
    };
    window.addEventListener("pointerdown", onPointerDown);
    window.addEventListener("keydown", onKey);
    return () => {
      window.removeEventListener("pointerdown", onPointerDown);
      window.removeEventListener("keydown", onKey);
    };
  }, [menuOpen]);

  const openAuth = () => {
    setMenuOpen(false);
    setAuthOpen(true);
  };

  return (
    <header className="sticky top-0 z-[100] bg-background">
      <div className="relative mx-auto flex h-20 max-w-[1760px] items-center justify-between px-6 md:px-10 lg:px-20">
        <Link href="/" className="shrink-0">
          <HacubaLogo className="text-primary" />
        </Link>

        <nav className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 hidden md:flex items-center gap-1 rounded-full border border-border bg-muted/50 px-2 py-1.5 shadow-sm">
          {navItems.map((item) => {
            const isActive = pathname === item.href;
            return (
              <Link
                key={item.label}
                href={item.href}
                className="relative flex items-center gap-2 rounded-full px-4 py-2 text-sm font-medium transition-colors hover:bg-muted/30"
              >
                {isActive && (
                  <motion.div
                    layoutId="nav-pill"
                    className="absolute inset-0 rounded-full bg-foreground"
                    transition={{ type: "spring", stiffness: 400, damping: 30 }}
                  />
                )}
                <span className="relative text-base">{item.emoji}</span>
                <span
                  className={`relative ${
                    isActive ? "text-background" : "text-foreground"
                  }`}
                >
                  {item.label}
                </span>
              </Link>
            );
          })}
        </nav>

        <div className="flex items-center gap-2">
          <button className="hidden lg:flex items-center gap-2 rounded-full border border-border px-4 py-2.5 text-sm font-medium text-foreground transition-colors hover:shadow-md">
            Become a seller
          </button>
          <button className="flex items-center justify-center rounded-full border border-border p-2.5 text-foreground transition-colors hover:shadow-md">
            <Globe className="h-4 w-4" />
          </button>
          <div ref={menuRef} className="relative">
            <button
              onClick={() => setMenuOpen((v) => !v)}
              aria-haspopup="menu"
              aria-expanded={menuOpen}
              aria-label="Account menu"
              className="flex items-center gap-3 rounded-full border border-border py-1.5 pl-3 pr-1.5 text-foreground transition-colors hover:shadow-md"
            >
              <Menu className="h-4 w-4" />
              <div className="flex h-8 w-8 items-center justify-center rounded-full bg-muted text-sm font-semibold text-muted-foreground">
                {userEmail ? (
                  <span aria-hidden="true">
                    {userEmail.charAt(0).toUpperCase()}
                  </span>
                ) : (
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
                )}
              </div>
            </button>

            <AnimatePresence>
              {menuOpen && (
                <motion.div
                  role="menu"
                  aria-label="Account"
                  className="absolute top-full right-0 mt-2 w-60 overflow-hidden rounded-md border border-border bg-popover py-2 text-popover-foreground shadow-lg"
                  initial={{ opacity: 0, y: -6, scale: 0.98 }}
                  animate={{ opacity: 1, y: 0, scale: 1 }}
                  exit={{ opacity: 0, y: -6, scale: 0.98 }}
                  transition={{ duration: 0.15 }}
                >
                  {userEmail ? (
                    <>
                      <p className="truncate px-4 pt-1 pb-2 text-xs text-muted-foreground">
                        Signed in as {userEmail}
                      </p>
                      <div className="mb-1 h-px bg-border" />
                      <button
                        role="menuitem"
                        onClick={() => {
                          dispatch(loggedOut());
                          setMenuOpen(false);
                        }}
                        className="flex w-full items-center gap-3 px-4 py-2.5 text-sm font-medium transition-colors hover:bg-muted"
                      >
                        <LogOut className="h-4 w-4" />
                        Log out
                      </button>
                      <div className="my-1 h-px bg-border" />
                    </>
                  ) : (
                    <>
                      <button
                        role="menuitem"
                        onClick={openAuth}
                        className="flex w-full items-center gap-3 px-4 py-2.5 text-sm font-semibold transition-colors hover:bg-muted"
                      >
                        <LogIn className="h-4 w-4" />
                        Log in / Sign up
                      </button>
                      <div className="my-1 h-px bg-border" />
                    </>
                  )}
                  <Link
                    role="menuitem"
                    href="/help"
                    onClick={() => setMenuOpen(false)}
                    className="flex items-center gap-3 px-4 py-2.5 text-sm font-medium transition-colors hover:bg-muted"
                  >
                    <LifeBuoy className="h-4 w-4" />
                    Help center
                  </Link>
                  <a
                    role="menuitem"
                    href={GITHUB_URL}
                    target="_blank"
                    rel="noreferrer"
                    onClick={() => setMenuOpen(false)}
                    className="flex items-center gap-3 px-4 py-2.5 text-sm font-medium transition-colors hover:bg-muted"
                  >
                    <GitHubMark />
                    GitHub
                  </a>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </div>
      </div>

      <AuthDialog open={authOpen} onClose={() => setAuthOpen(false)} />
    </header>
  );
}
