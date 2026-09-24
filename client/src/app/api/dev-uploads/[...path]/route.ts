const listingsServiceURL = process.env.LISTINGS_API_URL ?? "http://localhost:8081";

async function proxyUpload(request: Request, context: RouteContext<"/api/dev-uploads/[...path]">) {
  const { path } = await context.params;
  const target = `${listingsServiceURL}/dev-uploads/${path.map(encodeURIComponent).join("/")}`;
  const headers = new Headers();
  const contentType = request.headers.get("content-type");
  if (contentType) headers.set("content-type", contentType);
  const response = await fetch(target, {
    method: request.method,
    headers,
    body: request.method === "GET" ? undefined : request.body,
    cache: "no-store",
    duplex: request.body ? "half" : undefined,
  } as RequestInit & { duplex?: "half" });
  const responseHeaders = new Headers();
  const upstreamContentType = response.headers.get("content-type");
  if (upstreamContentType) responseHeaders.set("content-type", upstreamContentType);
  return new Response(response.body, { status: response.status, headers: responseHeaders });
}

export async function GET(request: Request, context: RouteContext<"/api/dev-uploads/[...path]">) {
  return proxyUpload(request, context);
}

export async function PUT(request: Request, context: RouteContext<"/api/dev-uploads/[...path]">) {
  return proxyUpload(request, context);
}
