import { proxyListingsRequest } from "../../../proxy";

export async function POST(
  request: Request,
  context: RouteContext<"/api/listings/[id]/images/presign">,
) {
  const { id } = await context.params;
  return proxyListingsRequest(request, `/listings/${encodeURIComponent(id)}/images/presign`);
}
