import { test, expect } from '@playwright/test'

test.describe('Chat', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
    await page.fill('[name="username"]', 'alice')
    await page.fill('[name="password"]', 'password123')
    await page.click('button[type="submit"]')
    await expect(page).toHaveURL('/')
  })

  test('user can view chat list', async ({ page }) => {
    await expect(page.locator('[data-testid="chat-list"]')).toBeVisible()
    await expect(page.locator('[data-testid="chat-item"]').first()).toBeVisible()
  })

  test('user can select a chat', async ({ page }) => {
    await page.locator('[data-testid="chat-item"]').first().click()

    await expect(page.locator('[data-testid="chat-header"]')).toBeVisible()
    await expect(page.locator('[data-testid="message-list"]')).toBeVisible()
  })

  test('user can send a message', async ({ page }) => {
    await page.locator('[data-testid="chat-item"]').first().click()

    await page.locator('[data-testid="message-input"]').fill('Hello from E2E test!')
    await page.locator('[data-testid="send-button"]').click()

    await expect(page.locator('[data-testid="message-bubble"]').last()).toContainText(
      'Hello from E2E test!',
    )
  })

  test('message input clears after sending', async ({ page }) => {
    await page.locator('[data-testid="chat-item"]').first().click()

    const input = page.locator('[data-testid="message-input"]')
    await input.fill('Test message')
    await page.locator('[data-testid="send-button"]').click()

    await expect(input).toHaveValue('')
  })

  test('Enter key sends message', async ({ page }) => {
    await page.locator('[data-testid="chat-item"]').first().click()

    await page.locator('[data-testid="message-input"]').fill('Sent via Enter')
    await page.locator('[data-testid="message-input"]').press('Enter')

    await expect(page.locator('[data-testid="message-bubble"]').last()).toContainText(
      'Sent via Enter',
    )
  })

  test('user receives message in real-time', async ({ page, browser }) => {
    // Alice opens chat
    await page.locator('[data-testid="chat-item"]').first().click()

    // Bob opens same chat in new context
    const bobContext = await browser.newContext()
    const bobPage = await bobContext.newPage()
    await bobPage.goto('/login')
    await bobPage.fill('[name="username"]', 'bob')
    await bobPage.fill('[name="password"]', 'password123')
    await bobPage.click('button[type="submit"]')
    await bobPage.locator('[data-testid="chat-item"]').first().click()

    // Alice sends a message
    const uniqueMsg = `Hello Bob ${Date.now()}!`
    await page.locator('[data-testid="message-input"]').fill(uniqueMsg)
    await page.locator('[data-testid="send-button"]').click()

    // Bob should see it
    await expect(bobPage.locator('[data-testid="message-bubble"]').last()).toContainText(
      uniqueMsg,
      { timeout: 5000 },
    )

    await bobContext.close()
  })

  test('typing indicator visible to other user', async ({ page, browser }) => {
    await page.locator('[data-testid="chat-item"]').first().click()

    const bobContext = await browser.newContext()
    const bobPage = await bobContext.newPage()
    await bobPage.goto('/login')
    await bobPage.fill('[name="username"]', 'bob')
    await bobPage.fill('[name="password"]', 'password123')
    await bobPage.click('button[type="submit"]')
    await bobPage.locator('[data-testid="chat-item"]').first().click()

    // Bob types
    await bobPage.locator('[data-testid="message-input"]').fill('I am typing...')

    // Alice should see typing indicator
    await expect(page.locator('[data-testid="typing-indicator"]')).toContainText(/typing/i, {
      timeout: 5000,
    })

    await bobContext.close()
  })
})
