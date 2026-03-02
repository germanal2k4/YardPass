import { test, expect } from '@playwright/test';

test.describe('Admin Residents Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login?role=admin');
    await page.getByLabel(/Имя пользователя/i).fill('admin');
    await page.getByLabel(/Пароль/i).fill('password');
    await page.getByRole('button', { name: /Войти/i }).click();
    await expect(page).toHaveURL(/\/admin/);

    await page.getByText('Жители').click();
    await expect(page).toHaveURL(/\/admin\/residents/);
  });

  test('displays residents table', async ({ page }) => {
    await expect(page.getByText('Список жителей')).toBeVisible();
    await expect(page.getByText('Иван Петров')).toBeVisible();
  });

  test('shows create form', async ({ page }) => {
    await expect(page.getByText('Добавить нового жителя')).toBeVisible();
  });

  test('shows bulk import section', async ({ page }) => {
    await expect(page.getByText('Массовое создание и импорт')).toBeVisible();
  });
});
