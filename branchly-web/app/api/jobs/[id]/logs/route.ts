import { authOptions } from "@/lib/auth";
import { getServerSession } from "next-auth";

export const dynamic = "force-dynamic";
export const runtime = "nodejs";

export async function GET(
  request: Request,
  context: { params: Promise<{ id: string }> }
) {
  const session = await getServerSession(authOptions);
  if (!session?.internalToken) {
    return new Response("Unauthorized", { status: 401 });
  }

  const base = process.env.API_URL?.replace(/\/$/, "");
  if (!base) {
    return new Response("API_URL not configured", { status: 500 });
  }

  const { id } = await context.params;
  const upstream = await fetch(
    `${base}/jobs/${encodeURIComponent(id)}/logs`,
    {
      headers: { Authorization: `Bearer ${session.internalToken}` },
      cache: "no-store",
      signal: request.signal,
    }
  );

  if (!upstream.ok) {
    const ct = upstream.headers.get("content-type") ?? "application/json";
    return new Response(upstream.body, {
      status: upstream.status,
      headers: { "Content-Type": ct },
    });
  }

  if (!upstream.body) {
    return new Response("Upstream has no body", { status: 502 });
  }

  return new Response(upstream.body, {
    status: upstream.status,
    headers: {
      "Content-Type": "text/event-stream; charset=utf-8",
      "Cache-Control": "no-cache, no-transform",
      "X-Accel-Buffering": "no",
    },
  });
}
