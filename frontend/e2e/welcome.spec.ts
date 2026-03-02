import { test, expect } from '@playwright/test';

test.describe('Welcome Page – Role Selection', () => {
  test('displays branding and role selection cards', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByText('YardPass')).toBeVisible();
    await expect(page.getByText('Система управления пропусками')).toBeVisible();
    await expect(page.getByText('Охрана')).toBeVisible();
    await expect(page.getByText('Администратор')).toBeVisible();
  });

  test('guard card navigates to login with role=guard', async ({ page }) => {
    await page.goto('/');
    await page.getByRole('button', { name: /Войти как охранник/i }).click();
    await expect(page).toHaveURL(/\/login\?role=guard/);
  });

  test('admin card navigates to login with role=admin', async ({ page }) => {
    await page.goto('/');
    await page.getByRole('button', { name: /Войти как администратор/i }).click();
    await expect(page).toHaveURL(/\/login\?role=admin/);
  });

  test('registration button navigates to /register', async ({ page }) => {
    await page.goto('/');
    await page.getByRole('button', { name: /Зарегистрироваться/i }).click();
    await expect(page).toHaveURL(/\/register/);
  });
});
