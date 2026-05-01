import { test, expect } from './fixtures';

test.describe('Authentication', () => {
  test('shows login page', async ({ page }) => {
    await page.goto('/login');
    await expect(page.getByText('Вход в систему')).toBeVisible();
    await expect(page.getByText('YardPass')).toBeVisible();
  });

  test('shows admin chip with ?role=admin', async ({ page }) => {
    await page.goto('/login?role=admin');
    await expect(page.getByText('Администратор')).toBeVisible();
  });

  test('shows guard chip with ?role=guard', async ({ page }) => {
    await page.goto('/login?role=guard');
    await expect(page.getByText('Охранник')).toBeVisible();
  });

  test('admin login redirects to /admin', async ({ page }) => {
    await page.goto('/login?role=admin');

    await page.getByLabel(/Имя пользователя/i).fill('admin');
    await page.getByLabel(/Пароль/i).fill('password');
    await page.getByRole('button', { name: /Войти/i }).click();

    await expect(page).toHaveURL(/\/admin/);
    await expect(page.getByText('Панель администратора')).toBeVisible();
  });

  test('guard login redirects to /security', async ({ page }) => {
    await page.goto('/login?role=guard');

    await page.getByLabel(/Имя пользователя/i).fill('guard');
    await page.getByLabel(/Пароль/i).fill('password');
    await page.getByRole('button', { name: /Войти/i }).click();

    await expect(page).toHaveURL(/\/security/);
    await expect(page.getByText('Сканирование пропусков')).toBeVisible();
  });

  test('shows error on invalid credentials', async ({ page }) => {
    await page.goto('/login');

    await page.getByLabel(/Имя пользователя/i).fill('wrong');
    await page.getByLabel(/Пароль/i).fill('wrong');
    await page.getByRole('button', { name: /Войти/i }).click();

    await expect(page.getByRole('alert')).toBeVisible();
  });
});
