import { apiFetch } from "@/lib/api-client";
import { NextRequest } from "next/server";

export async function GET() {
  try {
    const res = await apiFetch("/repositories");
    const data = await res.json();
    return Response.json(data, { status: res.status });
  } catch {
    return Response.json({ error: "Unauthorized" }, { status: 401 });
  }
}

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const res = await apiFetch("/repositories", {
      method: "POST",
      body: JSON.stringify(body),
    });
    const data = await res.json();
    return Response.json(data, { status: res.status });
  } catch {
    return Response.json({ error: "Unauthorized" }, { status: 401 });
  }
}
