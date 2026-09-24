import { proxyListingsRequest } from "../proxy";

export async function GET(request: Request) {
  return proxyListingsRequest(request, "/me/listings");
}
