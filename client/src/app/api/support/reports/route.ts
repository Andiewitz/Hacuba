import { proxySupportRequest } from "../proxy";

export async function POST(request: Request) {
  return proxySupportRequest(request, "/reports");
}

export async function GET(request: Request) { return proxySupportRequest(request, "/staff/reports"); }
