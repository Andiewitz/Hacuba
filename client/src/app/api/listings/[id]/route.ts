import { proxyListingsRequest } from "../proxy";

export async function PATCH(request: Request, context: RouteContext<"/api/listings/[id]">) {
  const { id } = await context.params;
  return proxyListingsRequest(request, `/listings/${encodeURIComponent(id)}`);
}

export async function DELETE(request: Request, context: RouteContext<"/api/listings/[id]">) {
  const { id } = await context.params;
  return proxyListingsRequest(request, `/listings/${encodeURIComponent(id)}`);
}
