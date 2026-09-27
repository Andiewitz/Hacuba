import { cookies } from "next/headers";
import { proxyListingsRequest } from "@/app/api/listings/proxy";

const viewerCookie = "hacuba_discovery_id";

export async function POST(request: Request) {
  const payload = await request.json().catch(() => null) as { events?: unknown } | null;
  if (!payload || !Array.isArray(payload.events)) return Response.json({ error: "invalid event batch" }, { status: 400 });
  const jar = await cookies();
  let viewerID = jar.get(viewerCookie)?.value;
  if (!viewerID) {
    viewerID = crypto.randomUUID();
    jar.set(viewerCookie, viewerID, { httpOnly: true, sameSite: "lax", secure: process.env.NODE_ENV === "production", path: "/", maxAge: 60 * 60 * 24 * 90 });
  }
  const upstream = new Request(request.url, { method: "POST", headers: { "content-type": "application/json" }, body: JSON.stringify({ viewer_id: viewerID, events: payload.events }) });
  return proxyListingsRequest(upstream, "/discovery/events");
}
