const listingsServiceURL = process.env.LISTINGS_API_URL ?? "http://localhost:8081";

const forwardedRequestHeaders = [
  "authorization",
  "content-type",
  "cookie",
  "x-csrf-token",
] as const;

export async function proxyListingsRequest(request: Request, path: string) {
  const headers = new Headers();
  for (const name of forwardedRequestHeaders) {
    const value = request.headers.get(name);
    if (value) headers.set(name, value);
  }

  let upstream: Response;
  try {
    upstream = await fetch(`${listingsServiceURL}${path}`, {
      method: request.method,
      headers,
      body: request.method === "GET" || request.method === "HEAD" ? undefined : request.body,
      cache: "no-store",
      duplex: request.body ? "half" : undefined,
    } as RequestInit & { duplex?: "half" });
  } catch {
    return Response.json(
      { error: "Listings are temporarily unavailable. Please try again." },
      { status: 503 },
    );
  }

  const responseHeaders = new Headers();
  const contentType = upstream.headers.get("content-type");
  if (contentType) responseHeaders.set("content-type", contentType);
  return new Response(upstream.body, { status: upstream.status, headers: responseHeaders });
}
