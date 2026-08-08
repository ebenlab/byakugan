import { test, expect } from '@playwright/test';

// The demo fixture is byakugan's own dossier: one project ("byakugan") with
// two pages — architecture.html and internals.html. Fixture-sensitive terms:
// "internals" ranks its page first by title match; "dispatch" appears only in
// internals.html (in a heading, so it survives the 2,000-char text cap) and
// the single-hit filter test relies on that.

test.describe('landing page', () => {
  test('lists the project with page counts', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('.bk-card')).toHaveCount(1);
    await expect(page.getByRole('link', { name: 'byakugan', exact: true })).toBeVisible();
    await expect(page.locator('#bk-meta')).toContainText('2 pages');
  });

  test('search returns ranked hits with highlights and hides the sections', async ({ page }) => {
    await page.goto('/');
    await page.locator('#bk-search').fill('internals');
    const hit = page.locator('.bk-hit').first();
    await expect(hit).toContainText('Byakugan Internals');
    await expect(hit.locator('mark').first()).toBeVisible();
    await expect(page.locator('#bk-projects')).toBeHidden();
  });

  test('slash focuses search, enter opens the first hit', async ({ page }) => {
    await page.goto('/');
    await page.keyboard.press('/');
    await expect(page.locator('#bk-search')).toBeFocused();
    await page.keyboard.type('internals');
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/byakugan\/internals\.html$/);
  });

  test('project URL scopes the landing page to that project', async ({ page }) => {
    await page.goto('/byakugan/');
    await expect(page.locator('.bk-card')).toHaveCount(1);
    await expect(page.getByRole('link', { name: 'byakugan', exact: true })).toBeVisible();
  });
});

test.describe('document pages', () => {
  test('serves the original document with the overlay injected', async ({ page }) => {
    await page.goto('/byakugan/architecture.html');
    await expect(page.getByRole('heading', { name: /Byakugan Architecture/ })).toBeVisible();
    await expect(page.locator('#bk-fab')).toBeVisible();
  });

  test('drawer shows the tree, current page, and prev/next', async ({ page }) => {
    await page.goto('/byakugan/architecture.html');
    await page.locator('#bk-fab').click();
    const drawer = page.locator('#bk-drawer');
    await expect(drawer).toHaveClass(/bk-open/);
    await expect(drawer.locator('.bk-tree a.bk-current')).toContainText('Byakugan Architecture');
    // First page of the dossier: nothing before it, internals after it.
    await expect(drawer.locator('#bk-prev')).toHaveAttribute('aria-disabled', 'true');
    await expect(drawer.locator('#bk-next')).toContainText('Byakugan Internals');
    await drawer.locator('#bk-next').click();
    await expect(page).toHaveURL(/internals\.html$/);
  });

  test('keyboard: b toggles the drawer, escape closes it', async ({ page }) => {
    await page.goto('/byakugan/architecture.html');
    await page.keyboard.press('b');
    await expect(page.locator('#bk-drawer')).toHaveClass(/bk-open/);
    await page.keyboard.press('Escape');
    await expect(page.locator('#bk-drawer')).not.toHaveClass(/bk-open/);
  });

  test('drawer search filters the tree', async ({ page }) => {
    await page.goto('/byakugan/architecture.html');
    await page.locator('#bk-fab').click();
    // "dispatch" appears only in the internals page (heading + body).
    await page.locator('#bk-drawer input').fill('dispatch');
    const links = page.locator('#bk-drawer .bk-tree a');
    await expect(links).toHaveCount(1);
    await expect(links.first()).toContainText('Byakugan Internals');
  });
});

test.describe('mobile (375px)', () => {
  test.use({ viewport: { width: 375, height: 812 } });

  test('landing reflows: full-width search, stacked sections, no horizontal scroll', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('.bk-card')).toHaveCount(1);
    // Search sits on its own row and spans (nearly) the full viewport.
    const search = await page.locator('#bk-search').boundingBox();
    expect(search!.width).toBeGreaterThan(300);
    // Nothing forces horizontal scrolling.
    const scrollWidth = await page.evaluate(() => document.documentElement.scrollWidth);
    expect(scrollWidth).toBeLessThanOrEqual(375);
  });

  test('drawer opens near full-width with touch-sized rows', async ({ page }) => {
    await page.goto('/byakugan/architecture.html');
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

  test('drawer toggle themes the overlay and doc.css pages in sync', async ({ page }) => {
    await page.goto('/byakugan/architecture.html');
    const hostBg = () => page.evaluate(() => getComputedStyle(document.body).backgroundColor);
    await page.locator('#bk-fab').click();
    const drawerBg = () => page.locator('#bk-drawer')
      .evaluate(el => getComputedStyle(el).backgroundColor);
    const btn = page.locator('#bk-drawer .bk-theme-btn');
    await btn.click(); // auto → light
    const lightBg = await drawerBg();
    const lightHostBg = await hostBg();
    await btn.click(); // light → dark
    await expect(page.locator('html')).toHaveAttribute('data-bk-theme', 'dark');
    // The drawer switches to the dark card color (palette-agnostic: it only
    // has to differ from the light value)…
    const darkBg = await drawerBg();
    expect(darkBg).not.toBe(lightBg);
    // …and a document styled with the shared doc.css opts into the same
    // explicit choice, so the page flips with the chrome.
    const darkHostBg = await hostBg();
    expect(darkHostBg).not.toBe(lightHostBg);
    // The choice persists and applies immediately on the next load.
    await page.reload();
    await expect(page.locator('html')).toHaveAttribute('data-bk-theme', 'dark');
    expect(await drawerBg()).toBe(darkBg);
    expect(await hostBg()).toBe(darkHostBg);
  });
});

test.describe('navigation affordances', () => {
  test('landing sections show per-project updated recency', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('.bk-card .bk-card-updated').first()).toContainText(/^Updated /);
    await expect(page.locator('.bk-card-updated')).toHaveCount(1);
  });

  test('drawer groups projects with per-project counts', async ({ page }) => {
    await page.goto('/byakugan/architecture.html');
    await page.locator('#bk-fab').click();
    const counts = page.locator('#bk-drawer .bk-tree-count');
    await expect(counts).toHaveCount(1);
    const byakugan = page.locator('#bk-drawer .bk-tree-project', { hasText: 'byakugan' });
    await expect(byakugan.locator('.bk-tree-count')).toHaveText('2');
  });
});

test.describe('back button', () => {
  test('search state lives in the URL and survives going back', async ({ page }) => {
    await page.goto('/');
    await page.locator('#bk-search').fill('internals');
    await expect(page).toHaveURL(/\?q=internals/);
    await page.locator('.bk-hit').first().click();
    await expect(page).toHaveURL(/byakugan\/internals\.html$/);
    await page.goBack();
    await expect(page.locator('#bk-search')).toHaveValue('internals');
    await expect(page.locator('.bk-hit').first()).toContainText('Byakugan Internals');
    await expect(page.locator('#bk-projects')).toBeHidden();
  });

  test('landing opened with ?q= starts searched', async ({ page }) => {
    await page.goto('/?q=debounce');
    await expect(page.locator('#bk-search')).toHaveValue('debounce');
    await expect(page.locator('.bk-hit').first()).toBeVisible();
  });

  test('back closes the drawer instead of leaving the page', async ({ page }) => {
    await page.goto('/byakugan/architecture.html');
    await page.locator('#bk-fab').click();
    await expect(page.locator('#bk-drawer')).toHaveClass(/bk-open/);
    await page.goBack();
    await expect(page.locator('#bk-drawer')).not.toHaveClass(/bk-open/);
    await expect(page).toHaveURL(/architecture\.html$/);
  });

  test('closing the drawer from the UI leaves no stray history entry', async ({ page }) => {
    await page.goto('/byakugan/architecture.html');
    await page.locator('#bk-fab').click();
    await page.locator('.bk-drawer-close').click();
    await expect(page.locator('#bk-drawer')).not.toHaveClass(/bk-open/);
    await expect(page).toHaveURL(/architecture\.html$/);
    // The drawer's pushed entry was popped again: no bkDrawer state remains.
    expect(await page.evaluate(() => (history.state as any)?.bkDrawer ?? null)).toBeNull();
  });
});

test.describe('API', () => {
  test('index.json exposes projects, titles, and text', async ({ request }) => {
    const res = await request.get('/api/index.json');
    expect(res.ok()).toBeTruthy();
    const idx = await res.json();
    expect(idx.pageCount).toBe(2);
    const byakugan = idx.projects.find((p: any) => p.name === 'byakugan');
    expect(byakugan.pages.map((p: any) => p.title))
      .toEqual(['Byakugan Architecture', 'Byakugan Internals']);
  });

  test('path traversal is rejected', async ({ request }) => {
    const res = await request.get('/../go.mod');
    expect(res.status()).not.toBe(200);
  });
});
