import type { NextConfig } from "next";

const backend = process.env.BACKEND_URL || "http://127.0.0.1:8080";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  output: "standalone",
  async rewrites() {
    return [{ source: "/api/:path*", destination: `${backend}/api/:path*` }];
  },
};

export default nextConfig;
