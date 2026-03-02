import { test, expect } from './fixtures';

test.describe('Admin Rules Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login?role=admin');
    await page.getByLabel(/Имя пользователя/i).fill('admin');
    await page.getByLabel(/Пароль/i).fill('password');
    await page.getByRole('button', { name: /Войти/i }).click();
    await expect(page).toHaveURL(/\/admin/);

    await page.getByTestId('nav-rules').click();
    await expect(page).toHaveURL(/\/admin\/rules/);
  });

  test('displays rules form with loaded values', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /Правила и настройки/ })).toBeVisible();

    await expect(page.getByText('Настройка правил контрольного пункта')).toBeVisible();
  });

  test('allows editing quiet hours fields', async ({ page }) => {
    const startInput = page.locator('input[name="quiet_hours_start"]');
    await expect(startInput).toBeVisible();

    await startInput.click();
    await startInput.fill('23:00');
    await expect(startInput).toHaveValue('23:00');
  });

  test('save button is present and clickable', async ({ page }) => {
    const startInput = page.locator('input[name="quiet_hours_start"]');
    await expect(startInput).toBeVisible();
    await startInput.click();
    await startInput.fill('23:00');

    const saveBtn = page.getByRole('button', { name: /Сохранить/i });
    await expect(saveBtn).toBeVisible();
    await saveBtn.click();

    await expect(page).toHaveURL(/\/admin\/rules/);
  });
});
