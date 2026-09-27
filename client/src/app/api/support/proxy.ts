const supportServiceURL = process.env.SUPPORT_API_URL ?? "http://localhost:8082";
const supportProxySecret = process.env.SUPPORT_PROXY_SECRET;

export async function proxySupportRequest(request: Request, path: string) {
  const headers = new Headers();
  for (const name of ["authorization", "content-type"] as const) {
    const value = request.headers.get(name);
    if (value) headers.set(name, value);
  }
  if (supportProxySecret) {
    headers.set("X-Hacuba-Support-Proxy", supportProxySecret);
    const forwarded = request.headers.get("x-forwarded-for") ?? request.headers.get("x-real-ip");
    if (forwarded) headers.set("X-Forwarded-For", forwarded);
  }
  try {
    const upstream = await fetch(`${supportServiceURL}${path}`, {
      method: request.method,
      headers,
      body: request.method === "GET" || request.method === "HEAD" ? undefined : request.body,
      cache: "no-store",
      duplex: request.body ? "half" : undefined,
    } as RequestInit & { duplex?: "half" });
    const responseHeaders = new Headers();
    const contentType = upstream.headers.get("content-type");
    if (contentType) responseHeaders.set("content-type", contentType);
    return new Response(upstream.body, { status: upstream.status, headers: responseHeaders });
  } catch {
    return Response.json({ error: "Support is temporarily unavailable. Please try again." }, { status: 503 });
  }
}
