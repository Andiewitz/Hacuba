import { proxyListingsRequest } from "../../proxy";

const allowedActions = new Set(["publish", "unpublish", "close"]);

export async function POST(
  request: Request,
  context: RouteContext<"/api/listings/[id]/[action]">,
) {
  const { id, action } = await context.params;
  if (!allowedActions.has(action)) {
    return Response.json({ error: "not found" }, { status: 404 });
  }
  return proxyListingsRequest(
    request,
    `/listings/${encodeURIComponent(id)}/${encodeURIComponent(action)}`,
  );
}
