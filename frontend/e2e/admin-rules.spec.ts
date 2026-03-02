import { test, expect } from '@playwright/test';

test.describe('Admin Rules Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login?role=admin');
    await page.getByLabel(/Имя пользователя/i).fill('admin');
    await page.getByLabel(/Пароль/i).fill('password');
    await page.getByRole('button', { name: /Войти/i }).click();
    await expect(page).toHaveURL(/\/admin/);

    await page.getByText('Правила и настройки').click();
    await expect(page).toHaveURL(/\/admin\/rules/);
  });

  test('displays rules form with loaded values', async ({ page }) => {
    await expect(page.getByText(/Правила/i)).toBeVisible();

    await expect(page.getByLabel(/Начало тихих часов/i).or(
      page.locator('input[type="time"]').first(),
    )).toBeVisible();
  });

  test('allows editing quiet hours fields', async ({ page }) => {
    const startInput = page.getByLabel(/Начало тихих часов/i).or(
      page.locator('input[type="time"]').first(),
    );

    await startInput.click();
    await startInput.fill('23:00');
    await expect(startInput).toHaveValue('23:00');
  });

  test('save button is present and clickable', async ({ page }) => {
    const saveBtn = page.getByRole('button', { name: /Сохранить/i });
    await expect(saveBtn).toBeVisible();
    await saveBtn.click();

    // Should see success state or remain on the same page without errors
    await expect(page).toHaveURL(/\/admin\/rules/);
  });
});
