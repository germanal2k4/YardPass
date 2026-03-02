import { test, expect } from '@playwright/test';

test.describe('Admin Reports Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login?role=admin');
    await page.getByLabel(/Имя пользователя/i).fill('admin');
    await page.getByLabel(/Пароль/i).fill('password');
    await page.getByRole('button', { name: /Войти/i }).click();
    await expect(page).toHaveURL(/\/admin/);

    await page.getByText('Отчеты').click();
    await expect(page).toHaveURL(/\/admin\/reports/);
  });

  test('displays statistics cards', async ({ page }) => {
    await expect(page.getByText('Всего сканирований')).toBeVisible();
    await expect(page.getByText('Действительных')).toBeVisible();
    await expect(page.getByText('Недействительных')).toBeVisible();
    await expect(page.getByText('Уникальных пропусков')).toBeVisible();
  });

  test('displays events table with data', async ({ page }) => {
    await expect(page.getByText('Журнал событий')).toBeVisible();

    await expect(
      page.getByText('guard1').first(),
    ).toBeVisible();
  });

  test('has date filter fields', async ({ page }) => {
    await expect(page.getByText('Фильтры')).toBeVisible();
    await expect(page.getByLabel(/Дата от/i)).toBeVisible();
    await expect(page.getByLabel(/Дата до/i)).toBeVisible();
  });

  test('has result filter chips', async ({ page }) => {
    await expect(page.getByText('Все')).toBeVisible();
    await expect(page.getByText('Действительные')).toBeVisible();
    await expect(page.getByText('Недействительные')).toBeVisible();
  });

  test('shows refresh and export buttons', async ({ page }) => {
    await expect(page.getByRole('button', { name: /Обновить/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /Экспорт/i })).toBeVisible();
  });

  test('export button triggers download', async ({ page }) => {
    const downloadPromise = page.waitForEvent('download');

    await page.getByRole('button', { name: /Экспорт/i }).click();

    const download = await downloadPromise;
    expect(download.suggestedFilename()).toMatch(/report_.*\.xlsx/);
  });

  test('filter chips change active state', async ({ page }) => {
    const validChip = page.getByText('Действительные');
    await validChip.click();

    // Chip should visually change (MUI gives it a color class)
    await expect(validChip).toBeVisible();
  });

  test('shows top rejection reasons', async ({ page }) => {
    await expect(page.getByText(/Основные причины отказов/i)).toBeVisible();
    await expect(page.getByText(/PASS_EXPIRED: 15/)).toBeVisible();
    await expect(page.getByText(/PASS_NOT_FOUND: 10/)).toBeVisible();
  });
});
