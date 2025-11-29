import { test, expect } from '@playwright/test';

test('navigation and basic flow', async ({ page }) => {
  await page.goto('/');
  await expect(page).toHaveTitle(/Noted/);

  // Check if dashboard elements are present
  await expect(page.getByText('Total Notes')).toBeVisible();
  await expect(page.getByText('Total Accounts')).toBeVisible();

  // Navigate to Notes
  await page.click('a[href="/notes"]');
  await expect(page).toHaveURL(/.*\/notes/);
  await expect(page.getByRole('heading', { name: 'Notes' })).toBeVisible();

  // We can't easily test creation without a running backend in this setup unless we mock it
  // or run the full stack. The config uses `npm run preview`, which serves the frontend.
  // If the backend isn't running, API calls will fail.
  // For a robust E2E, we'd need to spin up the backend too.
  // Assuming for now this is a smoke test for frontend routing.
});
