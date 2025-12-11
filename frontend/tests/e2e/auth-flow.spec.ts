import { test, expect } from '@playwright/test';

test.describe('Authentication Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('user can register successfully', async ({ page }) => {
    // Navigate to auth page
    await page.click('text=Sign Up');
    await expect(page).toHaveURL(/.*auth/);

    // Fill registration form
    const timestamp = Date.now();
    const email = `testuser${timestamp}@example.com`;
    
    await page.fill('[data-testid="register-name"]', 'Test User');
    await page.fill('[data-testid="register-email"]', email);
    await page.fill('[data-testid="register-password"]', 'TestPassword123!');
    await page.fill('[data-testid="register-confirm-password"]', 'TestPassword123!');

    // Submit registration
    await page.click('[data-testid="register-submit"]');

    // Should redirect to profile or dashboard
    await expect(page).toHaveURL(/.*profile/);
    
    // Check for success message
    await expect(page.locator('text=Registration successful')).toBeVisible();
  });

  test('user can login successfully', async ({ page }) => {
    // Navigate to auth page
    await page.click('text=Sign In');
    await expect(page).toHaveURL(/.*auth/);

    // Fill login form
    await page.fill('[data-testid="login-email"]', 'test@example.com');
    await page.fill('[data-testid="login-password"]', 'TestPassword123!');

    // Submit login
    await page.click('[data-testid="login-submit"]');

    // Should redirect to profile or dashboard
    await expect(page).toHaveURL(/.*profile/);
    
    // Check for user menu or profile element
    await expect(page.locator('[data-testid="user-menu"]')).toBeVisible();
  });

  test('user login fails with invalid credentials', async ({ page }) => {
    // Navigate to auth page
    await page.click('text=Sign In');
    await expect(page).toHaveURL(/.*auth/);

    // Fill login form with invalid credentials
    await page.fill('[data-testid="login-email"]', 'invalid@example.com');
    await page.fill('[data-testid="login-password"]', 'WrongPassword123!');

    // Submit login
    await page.click('[data-testid="login-submit"]');

    // Should show error message
    await expect(page.locator('text=Invalid email or password')).toBeVisible();
    
    // Should stay on auth page
    await expect(page).toHaveURL(/.*auth/);
  });

  test('user registration fails with existing email', async ({ page }) => {
    // Navigate to auth page
    await page.click('text=Sign Up');
    await expect(page).toHaveURL(/.*auth/);

    // Fill registration form with existing email
    await page.fill('[data-testid="register-name"]', 'Test User');
    await page.fill('[data-testid="register-email"]', 'existing@example.com');
    await page.fill('[data-testid="register-password"]', 'TestPassword123!');
    await page.fill('[data-testid="register-confirm-password"]', 'TestPassword123!');

    // Submit registration
    await page.click('[data-testid="register-submit"]');

    // Should show error message
    await expect(page.locator('text=Email already exists')).toBeVisible();
  });

  test('user can logout successfully', async ({ page }) => {
    // First login
    await page.goto('/auth');
    await page.fill('[data-testid="login-email"]', 'test@example.com');
    await page.fill('[data-testid="login-password"]', 'TestPassword123!');
    await page.click('[data-testid="login-submit"]');

    // Wait for successful login
    await expect(page.locator('[data-testid="user-menu"]')).toBeVisible();

    // Click user menu and logout
    await page.click('[data-testid="user-menu"]');
    await page.click('text=Logout');

    // Should redirect to home page
    await expect(page).toHaveURL('/');
    
    // User menu should not be visible
    await expect(page.locator('[data-testid="user-menu"]')).not.toBeVisible();
    
    // Sign in button should be visible
    await expect(page.locator('text=Sign In')).toBeVisible();
  });

  test('password validation works correctly', async ({ page }) => {
    await page.goto('/auth');
    await page.click('text=Sign Up');

    // Try to register with weak password
    await page.fill('[data-testid="register-name"]', 'Test User');
    await page.fill('[data-testid="register-email"]', 'test@example.com');
    await page.fill('[data-testid="register-password"]', '123');
    await page.fill('[data-testid="register-confirm-password"]', '123');

    await page.click('[data-testid="register-submit"]');

    // Should show password validation error
    await expect(page.locator('text=Password must be at least 8 characters')).toBeVisible();
  });

  test('email validation works correctly', async ({ page }) => {
    await page.goto('/auth');
    await page.click('text=Sign Up');

    // Try to register with invalid email
    await page.fill('[data-testid="register-name"]', 'Test User');
    await page.fill('[data-testid="register-email"]', 'invalid-email');
    await page.fill('[data-testid="register-password"]', 'TestPassword123!');
    await page.fill('[data-testid="register-confirm-password"]', 'TestPassword123!');

    await page.click('[data-testid="register-submit"]');

    // Should show email validation error
    await expect(page.locator('text=Please enter a valid email address')).toBeVisible();
  });

  test('password confirmation works correctly', async ({ page }) => {
    await page.goto('/auth');
    await page.click('text=Sign Up');

    // Try to register with mismatched passwords
    await page.fill('[data-testid="register-name"]', 'Test User');
    await page.fill('[data-testid="register-email"]', 'test@example.com');
    await page.fill('[data-testid="register-password"]', 'TestPassword123!');
    await page.fill('[data-testid="register-confirm-password"]', 'DifferentPassword123!');

    await page.click('[data-testid="register-submit"]');

    // Should show password mismatch error
    await expect(page.locator('text=Passwords do not match')).toBeVisible();
  });

  test('protected routes redirect to auth when not logged in', async ({ page }) => {
    // Try to access protected routes without authentication
    const protectedRoutes = ['/profile', '/checkout', '/orders'];
    
    for (const route of protectedRoutes) {
      await page.goto(route);
      
      // Should redirect to auth page
      await expect(page).toHaveURL(/.*auth/);
      
      // Should show login form
      await expect(page.locator('[data-testid="login-form"]')).toBeVisible();
    }
  });

  test('user can reset password', async ({ page }) => {
    await page.goto('/auth');
    await page.click('text=Sign In');

    // Click forgot password link
    await page.click('text=Forgot Password?');

    // Should show password reset form
    await expect(page.locator('[data-testid="reset-password-form"]')).toBeVisible();

    // Fill reset form
    await page.fill('[data-testid="reset-email"]', 'test@example.com');
    await page.click('[data-testid="reset-submit"]');

    // Should show success message
    await expect(page.locator('text=Password reset email sent')).toBeVisible();
  });

  test('form fields have proper accessibility attributes', async ({ page }) => {
    await page.goto('/auth');
    
    // Check form fields have proper labels and ARIA attributes
    const emailInput = page.locator('[data-testid="login-email"]');
    await expect(emailInput).toHaveAttribute('aria-label');
    await expect(emailInput).toHaveAttribute('type', 'email');

    const passwordInput = page.locator('[data-testid="login-password"]');
    await expect(passwordInput).toHaveAttribute('aria-label');
    await expect(passwordInput).toHaveAttribute('type', 'password');

    // Check buttons have proper ARIA attributes
    const submitButton = page.locator('[data-testid="login-submit"]');
    await expect(submitButton).toHaveAttribute('aria-label');
    await expect(submitButton).toHaveAttribute('type', 'submit');
  });

  test('keyboard navigation works properly', async ({ page }) => {
    await page.goto('/auth');
    await page.click('text=Sign In');

    // Test tab navigation
    await page.keyboard.press('Tab');
    await expect(page.locator('[data-testid="login-email"]')).toBeFocused();

    await page.keyboard.press('Tab');
    await expect(page.locator('[data-testid="login-password"]')).toBeFocused();

    await page.keyboard.press('Tab');
    await expect(page.locator('[data-testid="login-submit"]')).toBeFocused();

    // Test Enter key submission
    await page.fill('[data-testid="login-email"]', 'test@example.com');
    await page.fill('[data-testid="login-password"]', 'TestPassword123!');
    await page.keyboard.press('Enter');

    // Should attempt to login
    await expect(page).toHaveURL(/.*profile/);
  });

  test('form shows loading state during submission', async ({ page }) => {
    await page.goto('/auth');
    await page.click('text=Sign In');

    // Fill form
    await page.fill('[data-testid="login-email"]', 'test@example.com');
    await page.fill('[data-testid="login-password"]', 'TestPassword123!');

    // Mock slow network response
    await page.route('**/api/v1/auth/login', async route => {
      await new Promise(resolve => setTimeout(resolve, 2000));
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ success: true, token: 'mock-token' })
      });
    });

    // Submit form
    await page.click('[data-testid="login-submit"]');

    // Should show loading state
    await expect(page.locator('[data-testid="login-submit"]')).toHaveAttribute('disabled');
    await expect(page.locator('[data-testid="loading-spinner"]')).toBeVisible();
  });
});