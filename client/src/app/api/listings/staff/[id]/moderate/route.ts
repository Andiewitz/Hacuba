import { proxyListingsRequest } from "../../../proxy";

export async function POST(request: Request, context: RouteContext<"/api/listings/staff/[id]/moderate">) {
  const { id } = await context.params;
  return proxyListingsRequest(request, `/staff/listings/${encodeURIComponent(id)}/moderate`);
}
