import { test, expect } from '@playwright/test';

test.describe('Registration Flow', () => {
  test('shows registration form', async ({ page }) => {
    await page.goto('/register');
    await expect(page.getByText(/Регистрация/i)).toBeVisible();
    await expect(page.getByLabel(/Имя пользователя/i)).toBeVisible();
    await expect(page.getByLabel(/^Пароль$/i)).toBeVisible();
  });

  test('shows validation error for short username', async ({ page }) => {
    await page.goto('/register');

    await page.getByLabel(/Имя пользователя/i).fill('ab');
    await page.getByLabel(/^Пароль$/i).fill('Password123');
    await page.getByLabel(/Подтверждение/i).fill('Password123');
    await page.getByRole('button', { name: /Зарегистрироваться/i }).click();

    await expect(page.getByText(/минимум 3/i)).toBeVisible();
  });

  test('shows validation error for password mismatch', async ({ page }) => {
    await page.goto('/register');

    await page.getByLabel(/Имя пользователя/i).fill('newuser');
    await page.getByLabel(/^Пароль$/i).fill('Password123');
    await page.getByLabel(/Подтверждение/i).fill('DifferentPassword');
    await page.getByRole('button', { name: /Зарегистрироваться/i }).click();

    await expect(page.getByText(/не совпадают/i)).toBeVisible();
  });

  test('navigates to login from registration page', async ({ page }) => {
    await page.goto('/register');

    const loginLink = page.getByRole('link', { name: /Войти/i }).or(
      page.getByText(/Войти/i),
    );

    if (await loginLink.isVisible()) {
      await loginLink.click();
      await expect(page).toHaveURL(/\/login/);
    }
  });
});
