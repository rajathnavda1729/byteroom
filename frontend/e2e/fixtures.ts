import { test as base, Page } from '@playwright/test'

/**
 * Seed users for E2E tests.
 * Assumes the backend is running with test seed data accessible via API.
 */
export const SEED_USERS = {
  alice: { username: 'alice', password: 'password123' },
  bob: { username: 'bob', password: 'password123' },
}

export async function loginAs(page: Page, username: string, password: string) {
  await page.goto('/login')
  await page.fill('[name="username"]', username)
  await page.fill('[name="password"]', password)
  await page.click('button[type="submit"]')
  await page.waitForURL('/')
}

export const test = base.extend<{
  authenticatedPage: Page
}>({
  authenticatedPage: async ({ page }, use) => {
    await loginAs(page, SEED_USERS.alice.username, SEED_USERS.alice.password)
    await use(page)
  },
})

export { expect } from '@playwright/test'
