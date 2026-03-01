import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import { svelteTesting } from "@testing-library/svelte/vite";

export default defineConfig({
  plugins: [
    svelte(),
    // Necessario para testes de componente em Svelte 5 com Vitest/Testing Library.
    svelteTesting()
  ],
  server: {
    port: 5173,
    strictPort: true,
    proxy: {
      // Encaminha chamadas da SPA para a API local durante desenvolvimento.
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true
      }
    }
  },
  test: {
    environment: "jsdom",
    setupFiles: ["./src/tests/setup.ts"],
    include: ["src/tests/**/*.test.ts"]
  }
});
