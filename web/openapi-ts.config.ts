import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
  input: "../api/docs/swagger.json",
  output: {
    format: "prettier",
    lint: "eslint",
    path: "./lib/api",
  },
  plugins: [
    {
      name: "@hey-api/client-axios",
      runtimeConfigPath: "../api-config",
    },
    {
      dates: true,
      name: "@hey-api/transformers",
    },
    {
      enums: "javascript",
      name: "@hey-api/typescript",
    },
    {
      name: "@hey-api/sdk",
      transformer: true,
    },
    "@hey-api/schemas",
    "@tanstack/react-query",
  ],
});
