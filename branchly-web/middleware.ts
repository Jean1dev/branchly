import nextAuthMiddleware from "next-auth/middleware";

export default nextAuthMiddleware;

export const config = {
  matcher: [
    "/dashboard/:path*",
    "/repositories/:path*",
    "/jobs/:path*",
    "/settings/:path*",
  ],
};
