import { test, expect } from "@playwright/test";

test("loads main page", async ({ page }) => {
  const response = await page.goto("/");
  await expect(response?.status()).toBe(200);
});

test("loading a private page without logging in leads to an HTTP 401 Unauthorized error", async ({
  page,
}) => {
  const response = await page.goto("/private");
  await expect(response?.status()).toBe(401);
});
