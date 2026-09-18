"use client";

import { motion, useReducedMotion } from "framer-motion";
import { usePathname } from "next/navigation";

// Route enter transition: bridges the content swap between category pages
// so navigation doesn't hard-cut. Enter-only by design — App Router
// unmounts the old route immediately, so there is no exit to animate.
// Transform + opacity only, under 300ms, disabled-motion aware.
export default function Template({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const reduce = useReducedMotion();

  return (
    <motion.main
      key={pathname}
      initial={{ opacity: 0, y: reduce ? 0 : 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.22, ease: [0.23, 1, 0.32, 1] }}
    >
      {children}
    </motion.main>
  );
}
