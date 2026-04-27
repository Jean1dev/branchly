import type { NextAuthOptions } from "next-auth";
import GithubProvider from "next-auth/providers/github";

function parseJwtExpiry(token: string): number | null {
  try {
    const parts = token.split(".");
    if (parts.length !== 3) return null;
    // base64url → base64, then decode (works in Node 16+ and browsers)
    const base64 = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    const payload = JSON.parse(atob(base64)) as { exp?: number };
    return typeof payload.exp === "number" ? payload.exp * 1000 : null;
  } catch {
    return null;
  }
}

async function issueInternalToken(params: {
  providerId: string;
  githubToken: string;
  email?: string | null;
  name?: string | null;
  avatarUrl?: string | null;
}): Promise<{ userId: string; internalToken: string; expiry: number } | null> {
  try {
    const res = await fetch(`${process.env.API_URL}/internal/auth/upsert`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Internal-Secret": process.env.INTERNAL_API_SECRET!,
      },
      body: JSON.stringify({
        provider_id: params.providerId,
        github_token: params.githubToken,
        email: params.email ?? null,
        name: params.name ?? null,
        avatar_url: params.avatarUrl ?? null,
      }),
    });

    if (!res.ok) return null;

    const json = (await res.json()) as {
      data?: { user_id: string; internal_token: string };
    };
    const data = json.data;
    if (!data?.user_id || !data?.internal_token) return null;

    return {
      userId: data.user_id,
      internalToken: data.internal_token,
      expiry: parseJwtExpiry(data.internal_token) ?? Date.now() + 7 * 24 * 60 * 60 * 1000,
    };
  } catch {
    return null;
  }
}

export const authOptions: NextAuthOptions = {
  secret: process.env.NEXTAUTH_SECRET,

  providers: [
    GithubProvider({
      clientId: process.env.GITHUB_CLIENT_ID!,
      clientSecret: process.env.GITHUB_CLIENT_SECRET!,
      authorization: { params: { scope: "repo user:email" } },
    }),
  ],

  callbacks: {
    async jwt({ token, account, profile }) {
      // Initial sign-in: issue internal token and store everything
      if (account && profile) {
        const pid =
          profile && typeof profile === "object" && "id" in profile
            ? String((profile as { id: string | number }).id)
            : "";
        const email =
          profile && typeof profile === "object" && "email" in profile
            ? ((profile as { email?: string | null }).email ?? null)
            : null;
        const name =
          profile && typeof profile === "object" && "name" in profile
            ? ((profile as { name?: string | null }).name ?? null)
            : null;
        const image =
          profile && typeof profile === "object" && "image" in profile
            ? ((profile as { image?: string | null }).image ?? null)
            : null;

        const issued = await issueInternalToken({
          providerId: pid,
          githubToken: account.access_token ?? "",
          email,
          name,
          avatarUrl: image,
        });

        if (!issued) throw new Error("Failed to upsert user in branchly-api");

        token.userId = issued.userId;
        token.internalToken = issued.internalToken;
        token.internalTokenExpiry = issued.expiry;
        token.githubToken = account.access_token ?? "";
        const p = profile as { login?: string };
        if (typeof p.login === "string") {
          token.githubLogin = p.login;
        }
        delete token.error;
        return token;
      }

      // Subsequent calls (triggered by client-side update()):
      // Refresh the internal token if it is expired or expiring within 5 minutes.
      const expiry =
        (token.internalTokenExpiry as number | undefined) ??
        parseJwtExpiry(token.internalToken as string) ??
        0;

      const needsRefresh = Date.now() >= expiry - 5 * 60 * 1000;

      if (needsRefresh && token.githubToken) {
        const issued = await issueInternalToken({
          providerId: token.sub ?? (token.userId as string),
          githubToken: token.githubToken as string,
        });

        if (issued) {
          token.internalToken = issued.internalToken;
          token.userId = issued.userId;
          token.internalTokenExpiry = issued.expiry;
          delete token.error;
        } else {
          token.error = "RefreshAccessTokenError";
        }
      }

      return token;
    },

    async session({ session, token }) {
      session.userId = token.userId as string;
      session.internalToken = token.internalToken as string;
      session.internalTokenExpiry = token.internalTokenExpiry as number | undefined;
      if (token.error) {
        session.error = token.error as string;
      }
      const login = token.githubLogin;
      if (typeof login === "string" && login.length > 0) {
        session.githubLogin = login;
      }
      return session;
    },
  },

  pages: {
    signIn: "/login",
    error: "/login",
  },
};
