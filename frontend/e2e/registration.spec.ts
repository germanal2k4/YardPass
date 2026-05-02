import { test, expect } from './fixtures';

test.describe('Subscription purchase (/register)', () => {
  test('shows subscription purchase form', async ({ page }) => {
    await page.goto('/register');
    await expect(page.getByRole('heading', { name: /Оплата подписки/i })).toBeVisible();
    await expect(page.getByLabel(/Здание/i)).toBeVisible();
    await expect(page.getByLabel(/Email для получения доступов/i)).toBeVisible();
    await expect(page.getByTestId('card-number-input')).toBeVisible();
  });

  test('shows validation error when building name is only whitespace', async ({ page }) => {
    await page.goto('/register');

    await page.getByLabel(/Здание/i).fill('   ');
    await page.getByLabel(/Email для получения доступов/i).fill('owner@example.com');
    await page.getByLabel(/Количество апартаментов/i).fill('10');
    await page.getByTestId('card-number-input').fill('4111111111111111');
    await page.getByLabel(/Владелец карты/i).fill('IVAN PETROV');
    await page.getByLabel(/Срок/i).fill('12/30');
    await page.getByLabel(/CVV/i).fill('123');
    await page.getByRole('button', { name: /Оплатить/i }).click();

    await expect(page.getByText(/Укажите название здания/i)).toBeVisible();
  });

  test('shows validation error for invalid apartment count', async ({ page }) => {
    await page.goto('/register');

    await page.getByLabel(/Здание/i).fill('ЖК Тестовый');
    await page.getByLabel(/Email для получения доступов/i).fill('owner@example.com');
    await page.getByLabel(/Количество апартаментов/i).fill('0');
    await page.getByTestId('card-number-input').fill('4111111111111111');
    await page.getByLabel(/Владелец карты/i).fill('IVAN PETROV');
    await page.getByLabel(/Срок/i).fill('12/30');
    await page.getByLabel(/CVV/i).fill('123');
    // Native min=1 blocks submit before React runs; we only assert app-level validation.
    await page.locator('form').evaluate((f: HTMLFormElement) => {
      f.noValidate = true;
    });
    await page.getByRole('button', { name: /Оплатить/i }).click();

    await expect(page.getByText(/Укажите корректное количество апартаментов/i)).toBeVisible();
  });

  test('navigates to login from registration page', async ({ page }) => {
    await page.goto('/register');

    const loginCta = page.getByRole('button', { name: /^Войти$/i });
    await expect(loginCta).toBeVisible();
    await loginCta.click();
    await expect(page).toHaveURL(/\/login/);
  });
});
