import { apiFetch } from "@/lib/api-client";

export async function GET() {
  try {
    const res = await apiFetch("/integrations");
    const data = await res.json();
    return Response.json(data, { status: res.status });
  } catch {
    return Response.json({ error: "Unauthorized" }, { status: 401 });
  }
}
