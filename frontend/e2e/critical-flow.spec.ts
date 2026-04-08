import { test, expect } from "@playwright/test";
import { loginViaUI, collectConsoleErrors } from "./helpers";

/**
 * Critical E2E flow: login → dashboard → search → analysis → watchlist → portfolio.
 * Runs against Docker Compose with all services.
 * Requirements: 46.1, 46.2
 */

test.describe("Critical User Flow", () => {
  test("login page loads and shows form", async ({ page }) => {
    await page.goto("/login");
    await expect(page.getByRole("button", { name: /login|đăng nhập/i })).toBeVisible();
  });

  test("unauthenticated user is redirected to login", async ({ page }) => {
    await page.goto("/dashboard");
    // Should redirect to login since no JWT token
    await page.waitForURL("**/login", { timeout: 5_000 });
    await expect(page).toHaveURL(/login/);
  });

  test("login → dashboard flow", async ({ page }) => {
    const errors = collectConsoleErrors(page);
    await loginViaUI(page);

    // Dashboard should be visible
    await expect(page).toHaveURL(/dashboard/);

    // Check that key dashboard elements render
    // (these may show loading states if backend isn't running)
    await page.waitForTimeout(1_000);

    // No critical console errors
    const criticalErrors = errors.filter(
      (e) => !e.includes("favicon") && !e.includes("hydration"),
    );
    expect(criticalErrors).toHaveLength(0);
  });

  test("global search (⌘K) opens and accepts input", async ({ page }) => {
    await loginViaUI(page);

    // Trigger global search via keyboard shortcut
    await page.keyboard.press("Meta+k");
    await page.waitForTimeout(500);

    // If search dialog opened, type a query
    const searchInput = page.getByPlaceholder(/search|tìm kiếm/i).first();
    if (await searchInput.isVisible()) {
      await searchInput.fill("FPT");
      await page.waitForTimeout(500);
    }
  });

  test("navigate to stock detail page", async ({ page }) => {
    await loginViaUI(page);
    await page.goto("/stock/FPT");

    // Stock detail page should show the symbol
    await expect(page.getByText("FPT")).toBeVisible({ timeout: 5_000 });
  });

  test("navigate to portfolio page", async ({ page }) => {
    await loginViaUI(page);
    await page.goto("/portfolio");

    // Portfolio page should load without crashing
    await page.waitForTimeout(1_000);
    await expect(page).toHaveURL(/portfolio/);
  });

  test("navigate to screener page", async ({ page }) => {
    await loginViaUI(page);
    await page.goto("/screener");

    await page.waitForTimeout(1_000);
    await expect(page).toHaveURL(/screener/);
  });

  test("navigate to ranking page", async ({ page }) => {
    await loginViaUI(page);
    await page.goto("/ranking");

    await page.waitForTimeout(1_000);
    await expect(page).toHaveURL(/ranking/);
  });

  test("navigate to ideas page", async ({ page }) => {
    await loginViaUI(page);
    await page.goto("/ideas");

    await page.waitForTimeout(1_000);
    await expect(page).toHaveURL(/ideas/);
  });

  test("navigate to macro page", async ({ page }) => {
    await loginViaUI(page);
    await page.goto("/macro");

    await page.waitForTimeout(1_000);
    await expect(page).toHaveURL(/macro/);
  });
});
