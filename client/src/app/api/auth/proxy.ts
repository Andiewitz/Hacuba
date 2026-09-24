const authServiceURL = process.env.AUTH_API_URL ?? "http://localhost:8080";

const forwardedRequestHeaders = [
  "authorization",
  "content-type",
  "cookie",
  "x-csrf-token",
] as const;

function backendUnavailable() {
  return Response.json(
    { error: "Authentication is temporarily unavailable. Please try again." },
    { status: 503 },
  );
}

export async function proxyAuthRequest(request: Request, path: string) {
  const headers = new Headers();
  for (const name of forwardedRequestHeaders) {
    const value = request.headers.get(name);
    if (value) headers.set(name, value);
  }

  let upstream: Response;
  try {
    upstream = await fetch(`${authServiceURL}${path}`, {
      method: request.method,
      headers,
      body: request.method === "GET" || request.method === "HEAD" ? undefined : request.body,
      cache: "no-store",
      duplex: request.body ? "half" : undefined,
    } as RequestInit & { duplex?: "half" });
  } catch {
    return backendUnavailable();
  }

  const responseHeaders = new Headers();
  const contentType = upstream.headers.get("content-type");
  if (contentType) responseHeaders.set("content-type", contentType);

  // Keep each cookie separate so both the HttpOnly refresh token and the
  // readable CSRF token reach the browser.
  for (const cookie of upstream.headers.getSetCookie()) {
    responseHeaders.append("set-cookie", cookie);
  }

  return new Response(upstream.body, {
    status: upstream.status,
    headers: responseHeaders,
  });
}
