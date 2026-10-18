import { runtimeConfigProblems } from "@/lib/server-config";

export const dynamic = "force-dynamic";

export function GET() {
  const missing = runtimeConfigProblems();
  if (missing.length > 0) {
    return Response.json({ status: "unavailable", service: "client" }, { status: 503 });
  }
  return Response.json({ status: "ok", service: "client" });
}
