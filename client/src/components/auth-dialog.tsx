"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { AnimatePresence, motion } from "framer-motion";
import { Eye, EyeOff, Pencil } from "lucide-react";
import { useAppDispatch, useAppSelector } from "@/lib/hooks";
import {
  authErrorCleared,
  loginWithPassword,
  registerWithPassword,
} from "@/lib/slices/auth";

function isValidEmail(value: string) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value.trim());
}

function GoogleMark() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" aria-hidden="true">
      <path
        fill="#4285F4"
        d="M23.5 12.3c0-.9-.1-1.5-.3-2.3H12v4.5h6.5c-.1 1.1-.8 2.7-2.4 3.8l-.1.1 3.5 2.7.2.1c2.2-2 3.8-5 3.8-8.9z"
      />
      <path
        fill="#34A853"
        d="M12 24c3.2 0 6-1.1 7.9-2.9l-3.8-2.9c-1 .7-2.4 1.2-4.1 1.2-3.2 0-5.9-2.1-6.8-5.1l-.1.1-3.6 2.8v.1C3.5 21.3 7.5 24 12 24z"
      />
      <path
        fill="#FBBC05"
        d="M5.2 14.3c-.2-.7-.4-1.5-.4-2.3s.1-1.6.4-2.3l-.1-.1-3.6-2.8-.1.1C.5 8.7 0 10.3 0 12s.5 3.3 1.4 4.7l3.8-2.4z"
      />
      <path
        fill="#EA4335"
        d="M12 4.7c1.8 0 3 .8 3.7 1.4l3.3-3.2C17.9 1.1 15.2 0 12 0 7.5 0 3.5 2.7 1.4 6.9l3.8 2.9c1-3 3.7-5.1 6.8-5.1z"
      />
    </svg>
  );
}

export default function AuthDialog({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const dispatch = useAppDispatch();
  const { status, error } = useAppSelector((s) => s.auth);
  const [email, setEmail] = useState("");
  const [emailCommitted, setEmailCommitted] = useState<string | null>(null);
  const [mode, setMode] = useState<"login" | "register">("login");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [emailError, setEmailError] = useState<string | null>(null);
  const passwordRef = useRef<HTMLInputElement>(null);

  const loading = status === "loading";

  const resetAndClose = useCallback(() => {
    setEmailError(null);
    setPassword("");
    setEmailCommitted(null);
    setMode("login");
    dispatch(authErrorCleared());
    onClose();
  }, [dispatch, onClose]);

  useEffect(() => {
    if (emailCommitted) passwordRef.current?.focus();
  }, [emailCommitted]);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") resetAndClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open, resetAndClose]);

  const continueWithEmail = () => {
    if (!isValidEmail(email)) {
      setEmailError("Enter a valid email address to continue.");
      return;
    }
    setEmailError(null);
    setEmailCommitted(email.trim());
  };

  const submitPassword = async () => {
    if (!emailCommitted || password.length === 0) return;
    const action = mode === "login" ? loginWithPassword : registerWithPassword;
    const result = await dispatch(action({ email: emailCommitted, password }));
    if (action.fulfilled.match(result)) resetAndClose();
  };

  return (
    <AnimatePresence>
      {open && (
        <motion.div
          className="fixed inset-0 z-[200] flex items-center justify-center p-4"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.18 }}
        >
          <button
            aria-label="Close login dialog"
            className="absolute inset-0 cursor-default bg-[rgba(28,53,45,0.6)]"
            onClick={resetAndClose}
          />
          <motion.div
            role="dialog"
            aria-modal="true"
            aria-labelledby="auth-dialog-title"
            className="relative w-full max-w-sm rounded-lg border border-border bg-popover p-6 text-popover-foreground shadow-lg md:p-8"
            initial={{ opacity: 0, scale: 0.96, y: 12 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.97, y: 8 }}
            transition={{ type: "spring", stiffness: 380, damping: 30 }}
          >
            <h2
              id="auth-dialog-title"
              className="font-heading text-[22px] leading-[1.3] font-medium"
            >
              {mode === "login" ? "Welcome back" : "Create your account"}
            </h2>
            <p className="mt-2 text-sm leading-6 text-muted-foreground">
              {mode === "login"
                ? "Log in to continue browsing and manage your seller access."
                : "Create an account to start managing your seller access."}
            </p>

            <button
              type="button"
              disabled
              aria-disabled="true"
              title="Coming soon"
              className="mt-6 flex h-11 w-full cursor-not-allowed items-center justify-center gap-2 rounded-md border border-border bg-background px-4 text-sm font-semibold opacity-60"
            >
              <GoogleMark />
              Continue with Google
            </button>
            <p className="mt-2 text-center text-xs text-muted-foreground">
              Google sign-in is coming soon.
            </p>

            <div className="my-6 flex items-center gap-3">
              <span className="h-px flex-1 bg-border" />
              <span className="text-xs tracking-wide text-muted-foreground">
                or
              </span>
              <span className="h-px flex-1 bg-border" />
            </div>

            <div className="mb-6 grid grid-cols-2 rounded-md border border-border bg-muted/40 p-1">
              {(
                [
                  ["login", "Log in"],
                  ["register", "Create account"],
                ] as const
              ).map(([value, label]) => (
                <button
                  key={value}
                  type="button"
                  onClick={() => {
                    setMode(value);
                    setPassword("");
                    dispatch(authErrorCleared());
                  }}
                  className={`rounded-sm px-3 py-2 text-sm font-semibold transition-colors ${
                    mode === value
                      ? "bg-background text-foreground shadow-sm"
                      : "text-muted-foreground hover:text-foreground"
                  }`}
                >
                  {label}
                </button>
              ))}
            </div>

            <AnimatePresence mode="wait" initial={false}>
              {emailCommitted === null ? (
                <motion.div
                  key="email-step"
                  initial={{ opacity: 0, x: -12 }}
                  animate={{ opacity: 1, x: 0 }}
                  exit={{ opacity: 0, x: -12 }}
                  transition={{ duration: 0.16 }}
                >
                  <label
                    htmlFor="auth-email"
                    className="text-sm font-semibold"
                  >
                    Email
                  </label>
                  <input
                    id="auth-email"
                    type="email"
                    autoComplete="email"
                    autoFocus
                    placeholder="you@example.com"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter") continueWithEmail();
                    }}
                    aria-invalid={emailError ? true : undefined}
                    className="mt-2 h-11 w-full rounded-sm border border-input bg-background px-3 text-sm outline-none placeholder:text-muted-foreground focus:border-ring focus:ring-2 focus:ring-ring/40 aria-invalid:border-destructive"
                  />
                  {emailError && (
                    <p className="mt-2 text-sm font-medium text-destructive">
                      {emailError}
                    </p>
                  )}
                  <button
                    type="button"
                    onClick={continueWithEmail}
                    className="mt-4 h-11 w-full rounded-md bg-primary px-4 text-sm font-semibold text-primary-foreground transition-colors hover:bg-primary/85"
                  >
                    Continue
                  </button>
                </motion.div>
              ) : (
                <motion.div
                  key="password-step"
                  initial={{ opacity: 0, x: 12 }}
                  animate={{ opacity: 1, x: 0 }}
                  exit={{ opacity: 0, x: 12 }}
                  transition={{ duration: 0.16 }}
                >
                  <div className="flex items-center justify-between rounded-md border border-border bg-background px-3 py-2.5">
                    <span className="truncate text-sm">{emailCommitted}</span>
                    <button
                      type="button"
                      onClick={() => {
                        setEmailCommitted(null);
                        setPassword("");
                        dispatch(authErrorCleared());
                      }}
                      className="flex shrink-0 items-center gap-1 rounded-full px-2 py-1 text-xs font-semibold text-muted-foreground transition-colors hover:text-foreground"
                    >
                      <Pencil className="h-3 w-3" />
                      Edit
                    </button>
                  </div>

                  <label
                    htmlFor="auth-password"
                    className="mt-4 block text-sm font-semibold"
                  >
                    Password
                  </label>
                  <div className="relative mt-2">
                    <input
                      id="auth-password"
                      ref={passwordRef}
                      type={showPassword ? "text" : "password"}
                      autoComplete={mode === "login" ? "current-password" : "new-password"}
                      placeholder={mode === "login" ? "Your password" : "8+ characters, letters and numbers"}
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === "Enter") submitPassword();
                      }}
                      className="h-11 w-full rounded-sm border border-input bg-background px-3 pr-11 text-sm outline-none placeholder:text-muted-foreground focus:border-ring focus:ring-2 focus:ring-ring/40"
                    />
                    <button
                      type="button"
                      aria-label={showPassword ? "Hide password" : "Show password"}
                      onClick={() => setShowPassword((v) => !v)}
                      className="absolute top-1/2 right-2 -translate-y-1/2 rounded-full p-1.5 text-muted-foreground transition-colors hover:text-foreground"
                    >
                      {showPassword ? (
                        <EyeOff className="h-4 w-4" />
                      ) : (
                        <Eye className="h-4 w-4" />
                      )}
                    </button>
                  </div>
                  {error && (
                    <p className="mt-2 text-sm font-medium text-destructive">
                      {error}
                    </p>
                  )}
                  {mode === "register" && (
                    <p className="mt-2 text-xs leading-5 text-muted-foreground">
                      Use 8–72 characters with at least one letter and number.
                    </p>
                  )}
                  <button
                    type="button"
                    onClick={submitPassword}
                    disabled={loading || password.length === 0}
                    className="mt-4 h-11 w-full rounded-md bg-primary px-4 text-sm font-semibold text-primary-foreground transition-colors hover:bg-primary/85 disabled:pointer-events-none disabled:opacity-50"
                  >
                    {loading
                      ? mode === "login"
                        ? "Logging in…"
                        : "Creating account…"
                      : mode === "login"
                        ? "Log in"
                        : "Create account"}
                  </button>
                </motion.div>
              )}
            </AnimatePresence>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
}
