import { proxySupportRequest } from "../proxy";

export async function POST(request: Request) {
  return proxySupportRequest(request, "/reports");
}
