import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    globals: true,
    environment: "node",
    testTimeout: 30000,
    hookTimeout: 30000,
    pool: "forks",
    poolOptions: {
      forks: {
        singleFork: true,
      },
    },
    server: {
      deps: {
        inline: ["stridejs"],
      },
    },
  },
  resolve: {
    alias: {
      stridejs: "stridejs/dist/esm", // Force ESM path
    },
  },
});
