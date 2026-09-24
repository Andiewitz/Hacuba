import { proxyListingsRequest } from "./proxy";

export async function POST(request: Request) {
  return proxyListingsRequest(request, "/listings");
}
