import { test, expect } from '@playwright/test';

test.describe('Access control', () => {
  test('unauthenticated user is redirected to login from /security', async ({ page }) => {
    await page.goto('/security');
    await expect(page).toHaveURL(/\/login/);
  });

  test('unauthenticated user is redirected to login from /admin', async ({ page }) => {
    await page.goto('/admin');
    await expect(page).toHaveURL(/\/login/);
  });

  test('/forbidden page is accessible', async ({ page }) => {
    await page.goto('/forbidden');
    await expect(page.getByText(/запрещен|Forbidden|Доступ/i)).toBeVisible();
  });
});
