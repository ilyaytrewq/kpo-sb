import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// For local dev without nginx: /api -> localhost:8080
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
});
