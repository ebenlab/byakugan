import { test, expect } from '@playwright/test';

test.describe('landing page', () => {
  test('lists every project with page counts', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('.bk-card')).toHaveCount(3);
    await expect(page.getByText('Top-level documents')).toBeVisible();
    await expect(page.getByRole('link', { name: 'payments', exact: true })).toBeVisible();
    await expect(page.locator('#bk-meta')).toContainText('4 pages');
  });

  test('search returns ranked hits with highlights and hides the grid', async ({ page }) => {
    await page.goto('/');
    await page.locator('#bk-search').fill('serializable postgres');
    const hit = page.locator('.bk-hit').first();
    await expect(hit).toContainText('ADR-001');
    await expect(hit.locator('mark').first()).toBeVisible();
    await expect(page.locator('#bk-projects')).toBeHidden();
  });

  test('slash focuses search, enter opens the first hit', async ({ page }) => {
    await page.goto('/');
    await page.keyboard.press('/');
    await expect(page.locator('#bk-search')).toBeFocused();
    await page.keyboard.type('serializable postgres');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/adr-001-postgres\.html$/);
  });

  test('project URL scopes the landing page to that project', async ({ page }) => {
    await page.goto('/payments/');
    await expect(page.locator('.bk-card')).toHaveCount(1);
    await expect(page.getByRole('link', { name: 'payments', exact: true })).toBeVisible();
  });
});

test.describe('document pages', () => {
  test('serves the original document with the overlay injected', async ({ page }) => {
    await page.goto('/payments/adr-001-postgres.html');
    await expect(page.getByRole('heading', { name: /ADR-001/ })).toBeVisible();
    await expect(page.locator('#bk-fab')).toBeVisible();
  });

  test('drawer shows the tree, current page, and prev/next', async ({ page }) => {
    await page.goto('/payments/adr-001-postgres.html');
    await page.locator('#bk-fab').click();
    const drawer = page.locator('#bk-drawer');
    await expect(drawer).toHaveClass(/bk-open/);
    await expect(drawer.locator('.bk-tree a.bk-current')).toContainText('ADR-001');
    await expect(drawer.locator('#bk-prev')).toContainText('PRD: Signup Flow');
    await expect(drawer.locator('#bk-next')).toContainText('Payments Architecture');
    await drawer.locator('#bk-next').click();
    await expect(page).toHaveURL(/architecture\.html$/);
  });

  test('keyboard: b toggles the drawer, escape closes it', async ({ page }) => {
    await page.goto('/overview.html');
    await page.keyboard.press('b');
    await expect(page.locator('#bk-drawer')).toHaveClass(/bk-open/);
    await page.keyboard.press('Escape');
    await expect(page.locator('#bk-drawer')).not.toHaveClass(/bk-open/);
  });

  test('drawer search filters the tree', async ({ page }) => {
    await page.goto('/overview.html');
    await page.locator('#bk-fab').click();
    // "activation" appears only in the signup PRD's body text.
    await page.locator('#bk-drawer input').fill('activation');
    const links = page.locator('#bk-drawer .bk-tree a');
    await expect(links).toHaveCount(1);
    await expect(links.first()).toContainText('PRD: Signup Flow');
  });
});

test.describe('API', () => {
  test('index.json exposes projects, titles, and text', async ({ request }) => {
    const res = await request.get('/api/index.json');
    expect(res.ok()).toBeTruthy();
    const idx = await res.json();
    expect(idx.pageCount).toBe(4);
    const payments = idx.projects.find((p: any) => p.name === 'payments');
    expect(payments.pages.map((p: any) => p.title)).toContain('Payments Architecture');
  });

  test('path traversal is rejected', async ({ request }) => {
    const res = await request.get('/../go.mod');
    expect(res.status()).not.toBe(200);
  });
});
