import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import * as path from "node:path";

const devProxyTarget = process.env.VITE_DEV_PROXY_TARGET ?? "http://localhost";
const proxiedPaths = [
  "/auth",
  "/users",
  "/chats",
  "/deals",
  "/offers",
  "/offer-groups",
  "/reviews",
  "/providers",
  "/authors",
  "/tags",
  "/me",
  "/s3",
  "/swagger-specs",
  "/docs",
  "/admin/tags",
  "/admin/offer-reports",
];

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    }
  },
  server: {
    proxy: Object.fromEntries(
      proxiedPaths.map((proxyPath) => [
        proxyPath,
        {
          target: devProxyTarget,
          changeOrigin: true,
        },
      ]),
    ),
  },
})
