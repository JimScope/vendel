import { expect, test } from "@playwright/test"
import { cleanupCollection } from "./utils/privateApi.ts"

test.describe("Webhooks page", () => {
  test.describe.configure({ mode: "serial" })

  test.beforeEach(async ({ request }) => {
    await cleanupCollection(request, "webhook_delivery_logs")
    await cleanupCollection(request, "webhook_configs")
  })

  test.afterEach(async ({ request }) => {
    await cleanupCollection(request, "webhook_delivery_logs")
    await cleanupCollection(request, "webhook_configs")
  })

  test("Navigate to webhooks page", async ({ page }) => {
    await page.goto("/webhooks")
    await expect(page.getByRole("heading", { name: "Webhooks" })).toBeVisible()
    await expect(
      page.getByText(
        "Configure webhooks to receive notifications for incoming SMS",
      ),
    ).toBeVisible()
  })

  test("Empty state is shown when no webhooks exist", async ({ page }) => {
    await page.goto("/webhooks")
    await expect(page.getByText("No webhooks configured")).toBeVisible()
    await expect(
      page.getByText(
        "Add a webhook to receive notifications when SMS messages arrive",
      ),
    ).toBeVisible()
  })

  test("Add a new webhook", async ({ page }) => {
    await page.goto("/webhooks")

    await page.getByRole("button", { name: "Add Webhook" }).click()
    await expect(
      page.getByRole("heading", { name: "Add Webhook" }),
    ).toBeVisible()

    await page.getByLabel("URL *").fill("https://example.com/webhook-test")
    await page.getByLabel("Events").fill("incoming_sms")
    await page.getByRole("button", { name: "Create Webhook" }).click()

    // Webhook should appear in table
    await expect(
      page.getByRole("cell", { name: "https://example.com/webhook-test" }),
    ).toBeVisible({ timeout: 10000 })
    await expect(page.getByText("Active")).toBeVisible()
  })

  test("Test webhook and view logs", async ({ page }) => {
    await page.goto("/webhooks")

    // Create a webhook first
    await page.getByRole("button", { name: "Add Webhook" }).click()
    await page.getByLabel("URL *").fill("https://example.com/test-delivery")
    await page.getByLabel("Events").fill("incoming_sms")
    await page.getByRole("button", { name: "Create Webhook" }).click()

    await expect(
      page.getByRole("cell", { name: "https://example.com/test-delivery" }),
    ).toBeVisible({ timeout: 10000 })

    // Test the webhook
    const actionsButton = page
      .getByRole("row", { name: /example\.com\/test-delivery/ })
      .getByRole("button")
    await actionsButton.click()
    await page.getByRole("menuitem", { name: "Test Webhook" }).click()

    // Wait for the test delivery to complete, then close the dropdown
    await page.waitForTimeout(2000)
    await page.keyboard.press("Escape")

    // View logs
    await actionsButton.click()
    await page.getByRole("menuitem", { name: "View Logs" }).click()

    await expect(page.getByText("Delivery Logs")).toBeVisible()
    // The test delivery should appear as a failed log (badge element)
    await expect(
      page.locator('[data-slot="badge"]', { hasText: "FAIL" }),
    ).toBeVisible({ timeout: 10000 })
  })

  test("Edit a webhook", async ({ page }) => {
    await page.goto("/webhooks")

    // Create a webhook first
    await page.getByRole("button", { name: "Add Webhook" }).click()
    await page.getByLabel("URL *").fill("https://example.com/original")
    await page.getByRole("button", { name: "Create Webhook" }).click()

    await expect(
      page.getByRole("cell", { name: "https://example.com/original" }),
    ).toBeVisible({ timeout: 10000 })

    // Open actions menu and edit
    await page
      .getByRole("row", { name: /example\.com\/original/ })
      .getByRole("button")
      .click()
    await page.getByRole("menuitem", { name: "Edit Webhook" }).click()

    await expect(
      page.getByRole("heading", { name: "Edit Webhook" }),
    ).toBeVisible()

    await page.getByLabel("URL *").fill("https://example.com/updated")
    await page.getByRole("button", { name: "Save" }).click()

    await expect(
      page.getByRole("cell", { name: "https://example.com/updated" }),
    ).toBeVisible({ timeout: 10000 })
  })

  test("Delete a webhook", async ({ page }) => {
    await page.goto("/webhooks")

    // Create a webhook first
    await page.getByRole("button", { name: "Add Webhook" }).click()
    await page.getByLabel("URL *").fill("https://example.com/to-delete")
    await page.getByRole("button", { name: "Create Webhook" }).click()

    await expect(
      page.getByRole("cell", { name: "https://example.com/to-delete" }),
    ).toBeVisible({ timeout: 10000 })

    // Open actions menu and delete
    await page
      .getByRole("row", { name: /example\.com\/to-delete/ })
      .getByRole("button")
      .click()
    await page.getByRole("menuitem", { name: "Delete Webhook" }).click()

    await expect(
      page.getByRole("heading", { name: "Delete Webhook" }),
    ).toBeVisible()
    await page.getByRole("button", { name: "Delete" }).click()

    // Webhook should be removed
    await expect(page.getByText("No webhooks configured")).toBeVisible()
  })
})
