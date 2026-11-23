import { defineConfig, devices } from "@playwright/test";

const ci = Boolean(process.env.CI);

// Os testes mudam os dados do seed (votam, encerram, criam votações): rodam em série, um worker, sem nova tentativa.
export default defineConfig({
  testDir: "./tests",
  fullyParallel: false,
  workers: 1,
  retries: 0,
  forbidOnly: ci,
  timeout: 60_000,
  expect: { timeout: 10_000 },
  reporter: [["list"], ["html", { open: "never" }]],
  use: {
    baseURL: process.env.BASE_URL ?? "http://localhost:3000",
    locale: "pt-BR",
    timezoneId: "America/Sao_Paulo",
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
  },
  projects: [
    {
      name: "chromium",
      // Localmente usa o Google Chrome instalado; no CI, o Chromium baixado por `npx playwright install`.
      use: { ...devices["Desktop Chrome"], channel: ci ? undefined : "chrome" },
    },
  ],
});
