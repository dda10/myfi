import { defineConfig, devices } from "@playwright/test";

/**
 * Playwright E2E test configuration for EziStock frontend.
 * Runs against Docker Compose with all services.
 * Requirements: 46.1, 46.2
 */
export default defineConfig({
  testDir: "./e2e",
  fullyParallel: false, // Sequential for E2E flows
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: 1,
  reporter: "html",
  timeout: 30_000,

  use: {
    baseURL: process.env.E2E_BASE_URL ?? "http://localhost:3000",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
  },

  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],

  /* Start frontend dev server if not running against Docker Compose */
  webServer: process.env.E2E_BASE_URL
    ? undefined
    : {
        command: "npm run dev",
        url: "http://localhost:3000",
        reuseExistingServer: !process.env.CI,
        timeout: 60_000,
      },
});
