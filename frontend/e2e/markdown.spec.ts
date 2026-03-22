import { test, expect } from '@playwright/test'

test.describe('Markdown & Code Rendering', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
    await page.fill('[name="username"]', 'alice')
    await page.fill('[name="password"]', 'password123')
    await page.click('button[type="submit"]')
    await page.locator('[data-testid="chat-item"]').first().click()
    await expect(page.locator('[data-testid="message-list"]')).toBeVisible()
  })

  test('bold and italic text renders correctly', async ({ page }) => {
    await page.locator('[data-testid="message-input"]').fill('**bold** and *italic*')
    await page.locator('[data-testid="send-button"]').click()

    const bubble = page.locator('[data-testid="message-bubble"]').last()
    await expect(bubble.locator('strong')).toContainText('bold')
    await expect(bubble.locator('em')).toContainText('italic')
  })

  test('code blocks render with syntax highlighting', async ({ page }) => {
    const codeMessage = '```javascript\nconst x = 42;\nconsole.log(x);\n```'

    await page.locator('[data-testid="message-input"]').fill(codeMessage)
    await page.locator('[data-testid="send-button"]').click()

    const bubble = page.locator('[data-testid="message-bubble"]').last()
    await expect(bubble.locator('pre')).toBeVisible()
    // Syntax highlighter renders tokens inside <span> elements
    await expect(bubble.locator('pre span').first()).toBeVisible()
  })

  test('links open in new tab', async ({ page }) => {
    await page.locator('[data-testid="message-input"]').fill('[Google](https://google.com)')
    await page.locator('[data-testid="send-button"]').click()

    const link = page.locator('[data-testid="message-bubble"]').last().locator('a[href="https://google.com"]')
    await expect(link).toHaveAttribute('target', '_blank')
    await expect(link).toHaveAttribute('rel', 'noopener noreferrer')
  })

  test('inline code renders as code element', async ({ page }) => {
    await page.locator('[data-testid="message-input"]').fill('Use `const` keyword')
    await page.locator('[data-testid="send-button"]').click()

    const bubble = page.locator('[data-testid="message-bubble"]').last()
    await expect(bubble.locator('code')).toContainText('const')
  })
})
