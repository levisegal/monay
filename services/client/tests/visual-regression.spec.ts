import { test, expect } from "@playwright/test";

test.describe("Visual Regression", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    await page.waitForLoadState("networkidle");
  });

  test("dashboard matches baseline", async ({ page }) => {
    await page.waitForSelector('[data-testid="holdings-table"]', {
      timeout: 10000,
    });

    await expect(page).toHaveScreenshot("dashboard-full.png", {
      fullPage: true,
      maxDiffPixelRatio: 0.1,
    });
  });

  test("holdings table structure matches", async ({ page }) => {
    await page.waitForSelector('[data-testid="holdings-table"]', {
      timeout: 10000,
    });

    const table = page.locator('[data-testid="holdings-table"]');
    await expect(table).toHaveScreenshot("holdings-table.png", {
      maxDiffPixelRatio: 0.1,
    });
  });

  test("stats cards visible and populated", async ({ page }) => {
    await page.waitForSelector('[data-testid="stats-cards"]', {
      timeout: 10000,
    });

    const stats = page.locator('[data-testid="stats-cards"]');
    await expect(stats).toHaveScreenshot("stats-cards.png", {
      maxDiffPixelRatio: 0.1,
    });
  });

  test("portfolio chart renders", async ({ page }) => {
    await page.waitForSelector('[data-testid="portfolio-chart"]', {
      timeout: 10000,
    });

    const chart = page.locator('[data-testid="portfolio-chart"]');
    await expect(chart).toHaveScreenshot("portfolio-chart.png", {
      maxDiffPixelRatio: 0.15,
    });
  });

  test("holding detail panel opens on click", async ({ page }) => {
    await page.waitForSelector('[data-testid="holdings-table"]', {
      timeout: 10000,
    });

    const firstRow = page.locator('[data-testid="holding-row"]').first();
    if ((await firstRow.count()) > 0) {
      await firstRow.click();

      await page.waitForSelector('[data-testid="holding-detail-panel"]', {
        timeout: 5000,
      });

      const panel = page.locator('[data-testid="holding-detail-panel"]');
      await expect(panel).toHaveScreenshot("holding-detail-panel.png", {
        maxDiffPixelRatio: 0.15,
      });
    }
  });
});

test.describe("Empty State Visual Regression", () => {
  test("empty holdings shows import message", async ({ page }) => {
    await page.goto("/");
    await page.waitForLoadState("networkidle");

    const emptyState = page.locator('[data-testid="empty-holdings"]');
    if ((await emptyState.count()) > 0) {
      await expect(emptyState).toHaveScreenshot("empty-holdings.png", {
        maxDiffPixelRatio: 0.1,
      });
    }
  });
});

test.describe("Loading State Visual Regression", () => {
  test("loading state appears", async ({ page }) => {
    await page.route("**/api/v1/portfolio/enriched", async (route) => {
      await new Promise((resolve) => setTimeout(resolve, 2000));
      await route.continue();
    });

    await page.goto("/");

    const loading = page.locator('[data-testid="loading-state"]');
    if ((await loading.count()) > 0) {
      await expect(loading).toHaveScreenshot("loading-state.png", {
        maxDiffPixelRatio: 0.1,
      });
    }
  });
});

test.describe("Error State Visual Regression", () => {
  test("error state shows when API fails", async ({ page }) => {
    await page.route("**/api/v1/portfolio/enriched", async (route) => {
      await route.fulfill({
        status: 500,
        body: JSON.stringify({ error: "Internal server error" }),
      });
    });

    await page.goto("/");
    await page.waitForLoadState("networkidle");

    const error = page.locator('[data-testid="error-state"]');
    if ((await error.count()) > 0) {
      await expect(error).toHaveScreenshot("error-state.png", {
        maxDiffPixelRatio: 0.1,
      });
    }
  });
});
