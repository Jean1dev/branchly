import { apiFetch } from "@/lib/api-client";
import { NextRequest } from "next/server";

export async function DELETE(
  _request: NextRequest,
  context: { params: Promise<{ id: string }> }
) {
  try {
    const { id } = await context.params;
    const res = await apiFetch(`/repositories/${encodeURIComponent(id)}`, {
      method: "DELETE",
    });
    const data = await res.json();
    return Response.json(data, { status: res.status });
  } catch {
    return Response.json({ error: "Unauthorized" }, { status: 401 });
  }
}
