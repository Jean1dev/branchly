import { authOptions } from "@/lib/auth";
import { getServerSession } from "next-auth";

export async function apiFetch(path: string, init?: RequestInit) {
  const session = await getServerSession(authOptions);
  if (!session?.internalToken) throw new Error("Unauthorized");

  const headers = new Headers(init?.headers);
  if (!headers.has("Content-Type") && init?.body != null) {
    headers.set("Content-Type", "application/json");
  }
  headers.set("Authorization", `Bearer ${session.internalToken}`);

  return fetch(`${process.env.API_URL}${path}`, {
    ...init,
    headers,
    cache: "no-store",
  });
}
