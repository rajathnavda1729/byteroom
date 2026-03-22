import { test, expect } from '@playwright/test'

test.describe('Authentication', () => {
  test('user can register a new account', async ({ page }) => {
    await page.goto('/register')

    await page.fill('[name="username"]', 'newuser_e2e')
    await page.fill('[name="password"]', 'SecurePass123!')
    await page.fill('[name="confirmPassword"]', 'SecurePass123!')

    await page.click('button[type="submit"]')

    // After successful registration, redirected to chat
    await expect(page).toHaveURL('/')
  })

  test('user can login with valid credentials', async ({ page }) => {
    await page.goto('/login')

    await page.fill('[name="username"]', 'alice')
    await page.fill('[name="password"]', 'password123')

    await page.click('button[type="submit"]')

    await expect(page).toHaveURL('/')
    await expect(page.locator('[data-testid="chat-list"]')).toBeVisible()
  })

  test('user sees error on invalid login', async ({ page }) => {
    await page.goto('/login')

    await page.fill('[name="username"]', 'alice')
    await page.fill('[name="password"]', 'wrongpassword')

    await page.click('button[type="submit"]')

    await expect(page.locator('[role="alert"]')).toContainText(/invalid/i)
  })

  test('user can logout', async ({ page }) => {
    await page.goto('/login')
    await page.fill('[name="username"]', 'alice')
    await page.fill('[name="password"]', 'password123')
    await page.click('button[type="submit"]')

    await expect(page).toHaveURL('/')

    await page.click('button[aria-label="Logout"]')

    await expect(page).toHaveURL('/login')
  })

  test('protected route redirects unauthenticated users to login', async ({ page }) => {
    await page.goto('/')
    await expect(page).toHaveURL('/login')
  })
})
