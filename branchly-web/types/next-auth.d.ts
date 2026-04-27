import "next-auth";
import "next-auth/jwt";

declare module "next-auth" {
  interface Session {
    userId: string;
    internalToken: string;
    internalTokenExpiry?: number;
    error?: string;
    githubLogin?: string;
  }
}

declare module "next-auth/jwt" {
  interface JWT {
    userId: string;
    internalToken: string;
    internalTokenExpiry?: number;
    error?: string;
    githubToken: string;
    githubLogin?: string;
  }
}
