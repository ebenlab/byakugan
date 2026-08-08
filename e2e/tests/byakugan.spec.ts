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

test.describe('mobile (375px)', () => {
  test.use({ viewport: { width: 375, height: 812 } });

  test('landing reflows: full-width search, single-column cards, no horizontal scroll', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('.bk-card')).toHaveCount(3);
    // Search sits on its own row and spans (nearly) the full viewport.
    const search = await page.locator('#bk-search').boundingBox();
    expect(search!.width).toBeGreaterThan(300);
    // Cards stack in a single column: every card shares the same left edge.
    const lefts = await page.locator('.bk-card').evaluateAll(
      els => els.map(el => Math.round(el.getBoundingClientRect().left)));
    expect(new Set(lefts).size).toBe(1);
    // Nothing forces horizontal scrolling.
    const scrollWidth = await page.evaluate(() => document.documentElement.scrollWidth);
    expect(scrollWidth).toBeLessThanOrEqual(375);
  });

  test('drawer opens near full-width with touch-sized rows', async ({ page }) => {
    await page.goto('/payments/adr-001-postgres.html');
    await page.locator('#bk-fab').click();
    const drawer = page.locator('#bk-drawer');
    await expect(drawer).toHaveClass(/bk-open/);
    const box = await drawer.boundingBox();
    expect(box!.width).toBeGreaterThanOrEqual(330); // min(360px, 92vw) at 375px
    expect(box!.width).toBeLessThanOrEqual(375);
    const row = await page.locator('#bk-drawer .bk-tree a').first().boundingBox();
    expect(row!.height).toBeGreaterThanOrEqual(44);
  });
});

test.describe('theme toggle', () => {
  test('landing toggle cycles auto → light → dark and persists across reload', async ({ page }) => {
    await page.goto('/');
    const html = page.locator('html');
    const btn = page.locator('#bk-theme');
    await expect(btn).toBeVisible();
    await expect(html).not.toHaveAttribute('data-bk-theme', /.*/);
    await btn.click();
    await expect(html).toHaveAttribute('data-bk-theme', 'light');
    await btn.click();
    await expect(html).toHaveAttribute('data-bk-theme', 'dark');
    await page.reload();
    await expect(html).toHaveAttribute('data-bk-theme', 'dark');
    expect(await page.evaluate(() => localStorage.getItem('bk-theme'))).toBe('dark');
    await page.locator('#bk-theme').click(); // dark → back to auto
    await expect(html).not.toHaveAttribute('data-bk-theme', /.*/);
  });

  test('drawer toggle themes the overlay without touching the host page', async ({ page }) => {
    await page.goto('/overview.html');
    const hostBg = await page.evaluate(() => getComputedStyle(document.body).backgroundColor);
    await page.locator('#bk-fab').click();
    const drawerBg = () => page.locator('#bk-drawer')
      .evaluate(el => getComputedStyle(el).backgroundColor);
    const btn = page.locator('#bk-drawer .bk-theme-btn');
    await btn.click(); // auto → light
    const lightBg = await drawerBg();
    await btn.click(); // light → dark
    await expect(page.locator('html')).toHaveAttribute('data-bk-theme', 'dark');
    // The drawer switches to the dark card color (palette-agnostic: it only
    // has to differ from the light value)…
    const darkBg = await drawerBg();
    expect(darkBg).not.toBe(lightBg);
    // …while the host document's own styling is untouched.
    expect(await page.evaluate(() => getComputedStyle(document.body).backgroundColor)).toBe(hostBg);
    // The choice persists and applies immediately on the next load.
    await page.reload();
    await expect(page.locator('html')).toHaveAttribute('data-bk-theme', 'dark');
    expect(await drawerBg()).toBe(darkBg);
  });
});

test.describe('navigation affordances', () => {
  test('landing cards show per-project updated recency', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('.bk-card .bk-card-updated').first()).toContainText(/^Updated /);
    await expect(page.locator('.bk-card-updated')).toHaveCount(3);
  });

  test('drawer groups projects with per-project counts', async ({ page }) => {
    await page.goto('/overview.html');
    await page.locator('#bk-fab').click();
    const counts = page.locator('#bk-drawer .bk-tree-count');
    await expect(counts).toHaveCount(3);
    const payments = page.locator('#bk-drawer .bk-tree-project', { hasText: 'payments' });
    await expect(payments.locator('.bk-tree-count')).toHaveText('2');
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
