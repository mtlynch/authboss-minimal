import { test, expect } from "@playwright/test";

test("loads main page", async ({ page }) => {
  const response = await page.goto("/");
  await expect(response?.status()).toBe(200);
});
