import { proxyAuthRequest } from "../proxy";

export async function POST(request: Request) {
  return proxyAuthRequest(request, "/auth/register");
}
