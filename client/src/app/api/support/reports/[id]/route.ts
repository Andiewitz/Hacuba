import { proxySupportRequest } from "../../proxy";

export async function PATCH(request: Request, { params }: { params: Promise<{ id: string }> }) {
  return proxySupportRequest(request, `/staff/reports/${encodeURIComponent((await params).id)}`);
}
