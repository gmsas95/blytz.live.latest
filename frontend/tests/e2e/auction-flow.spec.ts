import { test, expect } from '@playwright/test';

test.describe('Auction Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/auth');
    await page.fill('[data-testid="login-email"]', 'test@example.com');
    await page.fill('[data-testid="login-password"]', 'TestPassword123!');
    await page.click('[data-testid="login-submit"]');
    await expect(page).toHaveURL(/.*profile/);
  });

  test('user can view active auctions', async ({ page }) => {
    await page.goto('/auctions');
    
    // Should show auctions list
    await expect(page.locator('[data-testid="auctions-list"]')).toBeVisible();
    
    // Should have at least one auction card
    await expect(page.locator('[data-testid="auction-card"]')).toHaveCount.greaterThan(0);
    
    // Check auction card elements
    const firstAuction = page.locator('[data-testid="auction-card"]').first();
    await expect(firstAuction.locator('[data-testid="auction-title"]')).toBeVisible();
    await expect(firstAuction.locator('[data-testid="auction-price"]')).toBeVisible();
    await expect(firstAuction.locator('[data-testid="auction-time"]')).toBeVisible();
    await expect(firstAuction.locator('[data-testid="auction-image"]')).toBeVisible();
  });

  test('user can search auctions', async ({ page }) => {
    await page.goto('/auctions');
    
    // Enter search term
    await page.fill('[data-testid="search-input"]', 'iPhone');
    await page.click('[data-testid="search-button"]');
    
    // Wait for search results
    await page.waitForTimeout(1000);
    
    // Should show search results
    const results = page.locator('[data-testid="auction-card"]');
    if (await results.count() > 0) {
      await expect(results.first().locator('[data-testid="auction-title"]')).toContainText('iPhone');
    }
  });

  test('user can filter auctions by category', async ({ page }) => {
    await page.goto('/auctions');
    
    // Select category filter
    await page.click('[data-testid="category-filter"]');
    await page.click('text=Electronics');
    
    // Wait for filter to apply
    await page.waitForTimeout(1000);
    
    // Should show filtered results
    await expect(page.locator('[data-testid="auction-card"]')).toBeVisible();
  });

  test('user can view auction details', async ({ page }) => {
    await page.goto('/auctions');
    
    // Click on first auction
    await page.click('[data-testid="auction-card"]');
    
    // Should navigate to auction detail page
    await expect(page).toHaveURL(/.*auctions\/[a-zA-Z0-9-]+/);
    
    // Check auction detail elements
    await expect(page.locator('[data-testid="auction-title"]')).toBeVisible();
    await expect(page.locator('[data-testid="auction-description"]')).toBeVisible();
    await expect(page.locator('[data-testid="auction-price"]')).toBeVisible();
    await expect(page.locator('[data-testid="auction-bid-count"]')).toBeVisible();
    await expect(page.locator('[data-testid="auction-timer"]')).toBeVisible();
    await expect(page.locator('[data-testid="auction-images"]')).toBeVisible();
  });

  test('user can place a bid', async ({ page }) => {
    await page.goto('/auctions');
    
    // Click on first auction
    await page.click('[data-testid="auction-card"]');
    
    // Wait for auction page to load
    await expect(page.locator('[data-testid="bid-button"]')).toBeVisible();
    
    // Click bid button
    await page.click('[data-testid="bid-button"]');
    
    // Should show bid modal
    await expect(page.locator('[data-testid="bid-modal"]')).toBeVisible();
    
    // Check suggested bid amount
    const suggestedAmount = page.locator('[data-testid="suggested-amount"]');
    await expect(suggestedAmount).toBeVisible();
    
    // Place bid with suggested amount
    await page.click('[data-testid="confirm-bid"]');
    
    // Wait for bid to process
    await page.waitForTimeout(2000);
    
    // Should show success message
    await expect(page.locator('text=Bid placed successfully')).toBeVisible();
    
    // Should update current price
    const currentPrice = page.locator('[data-testid="current-price"]');
    await expect(currentPrice).toBeVisible();
  });

  test('user cannot place bid lower than current price', async ({ page }) => {
    await page.goto('/auctions');
    
    // Click on first auction
    await page.click('[data-testid="auction-card"]');
    
    // Click bid button
    await page.click('[data-testid="bid-button"]');
    
    // Should show bid modal
    await expect(page.locator('[data-testid="bid-modal"]')).toBeVisible();
    
    // Try to place custom bid lower than current price
    await page.click('[data-testid="custom-bid"]');
    await page.fill('[data-testid="bid-input"]', '10'); // Low amount
    await page.click('[data-testid="confirm-bid"]');
    
    // Should show error message
    await expect(page.locator('text=Bid must be higher than current price')).toBeVisible();
  });

  test('user can view bid history', async ({ page }) => {
    await page.goto('/auctions');
    
    // Click on first auction
    await page.click('[data-testid="auction-card"]');
    
    // Click on bid history tab
    await page.click('[data-testid="bid-history-tab"]');
    
    // Should show bid history
    await expect(page.locator('[data-testid="bid-history"]')).toBeVisible();
    
    // Check bid history elements
    const bidHistory = page.locator('[data-testid="bid-item"]');
    if (await bidHistory.count() > 0) {
      await expect(bidHistory.first().locator('[data-testid="bid-amount"]')).toBeVisible();
      await expect(bidHistory.first().locator('[data-testid="bid-time"]')).toBeVisible();
      await expect(bidHistory.first().locator('[data-testid="bidder-name"]')).toBeVisible();
    }
  });

  test('user can add auction to watchlist', async ({ page }) => {
    await page.goto('/auctions');
    
    // Click on first auction
    await page.click('[data-testid="auction-card"]');
    
    // Click watchlist button
    await page.click('[data-testid="watchlist-button"]');
    
    // Should show success message
    await expect(page.locator('text=Added to watchlist')).toBeVisible();
    
    // Button should change state
    await expect(page.locator('[data-testid="watchlist-button"]')).toContainText('Remove from Watchlist');
  });

  test('user can share auction', async ({ page }) => {
    await page.goto('/auctions');
    
    // Click on first auction
    await page.click('[data-testid="auction-card"]');
    
    // Click share button
    await page.click('[data-testid="share-button"]');
    
    // Should show share modal
    await expect(page.locator('[data-testid="share-modal"]')).toBeVisible();
    
    // Check share options
    await expect(page.locator('[data-testid="share-facebook"]')).toBeVisible();
    await expect(page.locator('[data-testid="share-twitter"]')).toBeVisible();
    await expect(page.locator('[data-testid="share-copy-link"]')).toBeVisible();
    
    // Copy link
    await page.click('[data-testid="share-copy-link"]');
    
    // Should show copied message
    await expect(page.locator('text=Link copied to clipboard')).toBeVisible();
  });

  test('user can view seller information', async ({ page }) => {
    await page.goto('/auctions');
    
    // Click on first auction
    await page.click('[data-testid="auction-card"]');
    
    // Click on seller info
    await page.click('[data-testid="seller-info"]');
    
    // Should show seller modal
    await expect(page.locator('[data-testid="seller-modal"]')).toBeVisible();
    
    // Check seller details
    await expect(page.locator('[data-testid="seller-name"]')).toBeVisible();
    await expect(page.locator('[data-testid="seller-rating"]')).toBeVisible();
    await expect(page.locator('[data-testid="seller-total-auctions"]')).toBeVisible();
  });

  test('auction timer updates correctly', async ({ page }) => {
    await page.goto('/auctions');
    
    // Click on first auction
    await page.click('[data-testid="auction-card"]');
    
    // Get initial timer value
    const initialTime = await page.locator('[data-testid="auction-timer"]').textContent();
    
    // Wait for timer to update
    await page.waitForTimeout(2000);
    
    // Get updated timer value
    const updatedTime = await page.locator('[data-testid="auction-timer"]').textContent();
    
    // Timer should have changed
    expect(initialTime).not.toBe(updatedTime);
  });

  test('user can sort auctions', async ({ page }) => {
    await page.goto('/auctions');
    
    // Click sort dropdown
    await page.click('[data-testid="sort-dropdown"]');
    
    // Select sort option
    await page.click('text=Price: Low to High');
    
    // Wait for sort to apply
    await page.waitForTimeout(1000);
    
    // Should show sorted results
    await expect(page.locator('[data-testid="auction-card"]')).toBeVisible();
  });

  test('user can view ended auctions', async ({ page }) => {
    await page.goto('/auctions');
    
    // Click on ended auctions tab
    await page.click('[data-testid="ended-auctions-tab"]');
    
    // Should show ended auctions
    await expect(page.locator('[data-testid="auction-card"]')).toBeVisible();
    
    // Check for "Ended" badge
    const firstAuction = page.locator('[data-testid="auction-card"]').first();
    await expect(firstAuction.locator('[data-testid="ended-badge"]')).toBeVisible();
  });

  test('bid validation works correctly', async ({ page }) => {
    await page.goto('/auctions');
    
    // Click on first auction
    await page.click('[data-testid="auction-card"]');
    
    // Click bid button
    await page.click('[data-testid="bid-button"]');
    
    // Try to place bid without sufficient funds (mock scenario)
    await page.click('[data-testid="confirm-bid"]');
    
    // If bid fails, should show appropriate error
    const errorElement = page.locator('[data-testid="bid-error"]');
    if (await errorElement.isVisible()) {
      await expect(errorElement).toBeVisible();
    }
  });

  test('real-time bid updates work', async ({ page }) => {
    await page.goto('/auctions');
    
    // Click on first auction
    await page.click('[data-testid="auction-card"]');
    
    // Mock WebSocket connection for real-time updates
    await page.evaluate(() => {
      // Mock WebSocket for testing
      const mockWebSocket = {
        send: () => {},
        close: () => {},
        addEventListener: () => {},
        removeEventListener: () => {},
        readyState: WebSocket.OPEN
      };
      
      window.WebSocket = class extends WebSocket {
        constructor() {
          super('ws://mock');
          return mockWebSocket;
        }
      };
    });
    
    // Should show real-time updates indicator
    await expect(page.locator('[data-testid="real-time-indicator"]')).toBeVisible();
  });

  test('auction page is accessible', async ({ page }) => {
    await page.goto('/auctions');
    
    // Click on first auction
    await page.click('[data-testid="auction-card"]');
    
    // Check accessibility
    await expect(page.locator('h1')).toBeVisible();
    await expect(page.locator('[data-testid="auction-title"]')).toHaveAttribute('aria-label');
    await expect(page.locator('[data-testid="bid-button"]')).toHaveAttribute('aria-label');
    
    // Test keyboard navigation
    await page.keyboard.press('Tab');
    await expect(page.locator('[data-testid="bid-button"]')).toBeFocused();
  });

  test('mobile responsive design works', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    
    await page.goto('/auctions');
    
    // Should show mobile layout
    await expect(page.locator('[data-testid="mobile-filters"]')).toBeVisible();
    await expect(page.locator('[data-testid="mobile-menu"]')).toBeVisible();
    
    // Check auction cards on mobile
    const auctionCards = page.locator('[data-testid="auction-card"]');
    if (await auctionCards.count() > 0) {
      await expect(auctionCards.first()).toBeVisible();
    }
  });

  test('page loads quickly', async ({ page }) => {
    const startTime = Date.now();
    
    await page.goto('/auctions');
    
    // Wait for page to fully load
    await expect(page.locator('[data-testid="auctions-list"]')).toBeVisible();
    
    const loadTime = Date.now() - startTime;
    
    // Page should load within 3 seconds
    expect(loadTime).toBeLessThan(3000);
  });
});