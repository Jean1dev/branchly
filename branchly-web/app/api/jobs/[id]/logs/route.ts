import { authOptions } from "@/lib/auth";
import { getServerSession } from "next-auth";

export async function GET(
  _request: Request,
  context: { params: Promise<{ id: string }> }
) {
  const session = await getServerSession(authOptions);
  if (!session?.internalToken) {
    return new Response("Unauthorized", { status: 401 });
  }

  const { id } = await context.params;
  const upstream = await fetch(
    `${process.env.API_URL}/jobs/${encodeURIComponent(id)}/logs`,
    {
      headers: { Authorization: `Bearer ${session.internalToken}` },
      cache: "no-store",
    }
  );

  if (!upstream.ok) {
    const ct = upstream.headers.get("content-type") ?? "application/json";
    return new Response(upstream.body, {
      status: upstream.status,
      headers: { "Content-Type": ct },
    });
  }

  return new Response(upstream.body, {
    status: upstream.status,
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      Connection: "keep-alive",
    },
  });
}
