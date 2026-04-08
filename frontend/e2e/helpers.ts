import { Page, expect } from "@playwright/test";

const API_URL = process.env.E2E_API_URL ?? "http://localhost:8080";

/**
 * Test user credentials — must exist in the test database.
 * Create via POST /api/auth/register before running E2E tests.
 */
export const TEST_USER = {
  username: "e2e_test_user",
  password: "TestPass123!",
};

/** Login via the UI login page and wait for redirect to dashboard. */
export async function loginViaUI(page: Page): Promise<void> {
  await page.goto("/login");
  await page.getByPlaceholder(/username/i).fill(TEST_USER.username);
  await page.getByPlaceholder(/password/i).fill(TEST_USER.password);
  await page.getByRole("button", { name: /login|đăng nhập/i }).click();
  // Wait for redirect to dashboard
  await page.waitForURL("**/dashboard", { timeout: 10_000 });
}

/** Login via API and inject the JWT token into localStorage. */
export async function loginViaAPI(page: Page): Promise<string> {
  const resp = await page.request.post(`${API_URL}/api/auth/login`, {
    data: { username: TEST_USER.username, password: TEST_USER.password },
  });
  const body = await resp.json();
  const token = body.token as string;

  // Inject token into localStorage before navigating
  await page.goto("/login");
  await page.evaluate((t) => localStorage.setItem("ezistock-token", t), token);
  return token;
}

/** Assert the page has no console errors (warnings are OK). */
export function collectConsoleErrors(page: Page): string[] {
  const errors: string[] = [];
  page.on("console", (msg) => {
    if (msg.type() === "error") {
      errors.push(msg.text());
    }
  });
  return errors;
}
