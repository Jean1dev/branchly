import type { NextAuthOptions } from "next-auth";
import GithubProvider from "next-auth/providers/github";

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

        const res = await fetch(
          `${process.env.API_URL}/internal/auth/upsert`,
          {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              "X-Internal-Secret": process.env.INTERNAL_API_SECRET!,
            },
            body: JSON.stringify({
              provider_id: pid,
              email,
              name,
              avatar_url: image,
              github_token: account.access_token,
            }),
          }
        );

        if (!res.ok) throw new Error("Failed to upsert user in branchly-api");

        const json = (await res.json()) as {
          data?: { user_id: string; internal_token: string };
        };
        const payload = json.data;
        if (!payload?.user_id || !payload?.internal_token) {
          throw new Error("Invalid upsert response from branchly-api");
        }

        token.userId = payload.user_id;
        token.internalToken = payload.internal_token;
        token.githubToken = account.access_token ?? "";
        const p = profile as {
          login?: string;
        };
        if (typeof p.login === "string") {
          token.githubLogin = p.login;
        }
      }
      return token;
    },

    async session({ session, token }) {
      session.userId = token.userId as string;
      session.internalToken = token.internalToken as string;
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
