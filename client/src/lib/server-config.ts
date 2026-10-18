// Server-only deployment configuration. APP_ENV is deliberately explicit so
// local `next build` commands can retain their development conveniences while
// Railway production instances reject missing private-service settings.
export const isProduction = process.env.APP_ENV?.toLowerCase() === "production";

export function serviceURL(name: "AUTH_API_URL" | "LISTINGS_API_URL" | "SUPPORT_API_URL", fallback: string) {
  const value = process.env[name]?.trim();
  if (value) return value.replace(/\/$/, "");
  if (isProduction) throw new Error(`${name} is required when APP_ENV=production`);
  return fallback;
}

export function optionalServerValue(name: string) {
  const value = process.env[name]?.trim();
  if (!value && isProduction) throw new Error(`${name} is required when APP_ENV=production`);
  return value;
}

export function runtimeConfigProblems() {
  if (!isProduction) return [];
  return [
    "AUTH_API_URL",
    "LISTINGS_API_URL",
    "SUPPORT_API_URL",
    "LISTINGS_IMAGE_BASE_URL",
    "SUPPORT_PROXY_SECRET",
    "DISCOVERY_PROXY_SECRET",
  ].filter((name) => !process.env[name]?.trim());
}
