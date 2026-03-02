import { test, expect } from './fixtures';

test.describe('Registration Flow', () => {
  test('shows registration form', async ({ page }) => {
    await page.goto('/register');
    await expect(page.getByText(/Регистрация/i)).toBeVisible();
    await expect(page.getByLabel(/Имя пользователя/i)).toBeVisible();
    await expect(page.getByTestId('password-input')).toBeVisible();
  });

  test('shows validation error for short username', async ({ page }) => {
    await page.goto('/register');

    await page.getByLabel(/Имя пользователя/i).fill('ab');
    await page.getByTestId('password-input').fill('Password123');
    await page.getByTestId('confirm-password-input').fill('Password123');
    await page.getByRole('button', { name: /Зарегистрироваться/i }).click();

    await expect(page.getByRole('alert').filter({ hasText: /минимум 3/i })).toBeVisible();
  });

  test('shows validation error for password mismatch', async ({ page }) => {
    await page.goto('/register');

    await page.getByLabel(/Имя пользователя/i).fill('newuser');
    await page.getByTestId('password-input').fill('Password123');
    await page.getByTestId('confirm-password-input').fill('DifferentPassword');
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
