import { test, expect } from '@playwright/test';

test.describe('Security Page (Guard)', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login?role=guard');
    await page.getByLabel(/Имя пользователя/i).fill('guard');
    await page.getByLabel(/Пароль/i).fill('password');
    await page.getByRole('button', { name: /Войти/i }).click();
    await expect(page).toHaveURL(/\/security/);
  });

  test('renders QR and car plate sections', async ({ page }) => {
    await expect(page.getByText('Проверка QR-кода')).toBeVisible();
    await expect(page.getByText('Проверка по номеру')).toBeVisible();
  });

  test('validates QR code with valid UUID', async ({ page }) => {
    const qrInput = page.getByPlaceholder(/Сканируйте QR-код/i);
    await qrInput.fill('valid-uuid');
    await qrInput.press('Enter');

    await expect(page.getByText('Пропуск действителен')).toBeVisible();
  });

  test('shows error for unknown QR code', async ({ page }) => {
    const qrInput = page.getByPlaceholder(/Сканируйте QR-код/i);
    await qrInput.fill('unknown-uuid');
    await qrInput.press('Enter');

    await expect(page.getByRole('alert')).toBeVisible();
  });

  test('shows expired pass result', async ({ page }) => {
    const qrInput = page.getByPlaceholder(/Сканируйте QR-код/i);
    await qrInput.fill('expired-uuid');
    await qrInput.press('Enter');

    await expect(page.getByText(/недействителен|истек/i)).toBeVisible();
  });

  test('displays car plate input section', async ({ page }) => {
    await expect(page.getByText('Проверка по номеру')).toBeVisible();
    await expect(page.getByRole('button', { name: /Проверить номер/i })).toBeVisible();
  });

  test('car plate submit button is disabled when empty', async ({ page }) => {
    const submitBtn = page.getByRole('button', { name: /Проверить номер/i });
    await expect(submitBtn).toBeDisabled();
  });

  test('displays events log section', async ({ page }) => {
    await expect(page.getByText('Журнал событий')).toBeVisible();
  });

  test('error can be dismissed', async ({ page }) => {
    const qrInput = page.getByPlaceholder(/Сканируйте QR-код/i);
    await qrInput.fill('unknown-uuid');
    await qrInput.press('Enter');

    const alert = page.getByRole('alert');
    await expect(alert).toBeVisible();

    const closeBtn = page.getByRole('button', { name: /close/i });
    if (await closeBtn.isVisible()) {
      await closeBtn.click();
      await expect(alert).not.toBeVisible();
    }
  });
});
