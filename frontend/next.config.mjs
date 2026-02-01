/** @type {import("next").NextConfig} */
const apiBase = process.env.NEXT_PUBLIC_API_BASE || "http://127.0.0.1:8080";

const nextConfig = {
  reactStrictMode: true,
  async rewrites() {
    return [
      {
        source: "/api/v1/:path*",
        destination: `${apiBase.replace(/\/$/, "")}/api/v1/:path*`,
      },
    ];
  },
};

export default nextConfig;
