import { expect, test } from "@playwright/test"
import {
  cleanupCollection,
  createTestDevice,
  getCurrentUserId,
  resetUserQuota,
} from "./utils/privateApi.ts"

test.describe("SMS page", () => {
  test.describe.configure({ mode: "serial" })

  test.beforeEach(async ({ request }) => {
    await cleanupCollection(request, "sms_messages")
    await cleanupCollection(request, "sms_devices")
    await resetUserQuota(request)
  })

  test.afterEach(async ({ request }) => {
    await cleanupCollection(request, "sms_messages")
    await cleanupCollection(request, "sms_devices")
    await resetUserQuota(request)
  })

  test("Navigate to SMS page", async ({ page }) => {
    await page.goto("/sms")
    await expect(page.getByRole("heading", { name: "SMS" })).toBeVisible()
    await expect(page.getByText("Create and manage your SMS")).toBeVisible()
  })

  test("Tabs work correctly", async ({ page }) => {
    await page.goto("/sms")

    // All tab should be visible by default
    await expect(page.getByRole("tab", { name: "All" })).toBeVisible()
    await expect(page.getByRole("tab", { name: "Sent" })).toBeVisible()
    await expect(page.getByRole("tab", { name: "Received" })).toBeVisible()

    // Switch tabs and check empty states
    await page.getByRole("tab", { name: "Sent" }).click()
    await expect(page.getByText("You haven't sent any SMS yet")).toBeVisible()

    await page.getByRole("tab", { name: "Received" }).click()
    await expect(
      page.getByText("You haven't received any SMS yet"),
    ).toBeVisible()

    await page.getByRole("tab", { name: "All" }).click()
    await expect(page.getByText("You don't have any SMS yet")).toBeVisible()
  })

  test("Empty state is shown when no messages exist", async ({ page }) => {
    await page.goto("/sms")
    await expect(page.getByText("No messages found")).toBeVisible()
    await expect(page.getByText("You don't have any SMS yet")).toBeVisible()
  })

  test("Send SMS dialog opens with correct fields", async ({ page }) => {
    await page.goto("/sms")

    await page.getByRole("button", { name: "Send SMS" }).click()

    await expect(page.getByRole("heading", { name: "Send SMS" })).toBeVisible()
    await expect(
      page.getByText("Fill in the form below to add a sms to be sent."),
    ).toBeVisible()

    // Check form fields are present via their placeholders/labels
    const dialog = page.getByRole("dialog")
    await expect(
      dialog.getByPlaceholder("Phone numbers (space separated)"),
    ).toBeVisible()
    await expect(dialog.getByText("Select options")).toBeVisible()
    await expect(dialog.getByPlaceholder("Message Body")).toBeVisible()

    // Cancel closes the dialog
    await page.getByRole("button", { name: "Cancel" }).click()
    await expect(
      page.getByRole("heading", { name: "Send SMS" }),
    ).not.toBeVisible()
  })

  test("Send SMS with a device", async ({ page, request }) => {
    // Create a device via API for the test
    const userId = await getCurrentUserId(request)
    await createTestDevice(request, {
      name: "SMS Test Device",
      phone_number: "+15559999999",
      user_id: userId,
    })

    // Navigate after device creation to ensure fresh data
    await page.goto("/sms")

    await page.getByRole("button", { name: "Send SMS" }).click()
    await expect(page.getByRole("heading", { name: "Send SMS" })).toBeVisible()

    // Fill in recipients using TagInput (type + Enter to add tag)
    await page.getByPlaceholder("Phone numbers (space separated)").click()
    await page
      .getByPlaceholder("Phone numbers (space separated)")
      .fill("+15551112222")
    await page
      .getByPlaceholder("Phone numbers (space separated)")
      .press("Enter")

    // Select device from MultiSelect — click the trigger to open the popover
    await page.locator("button:has-text('Select options')").click()
    // Wait for the options to load and click the device
    await page.getByRole("option", { name: "SMS Test Device" }).click()
    // Close popover by pressing Escape
    await page.keyboard.press("Escape")

    // Fill in message body
    await page.getByPlaceholder("Message Body").fill("Hello from E2E test")

    // Submit
    await page.getByRole("button", { name: "Save" }).click()

    // The dialog should close and SMS should appear in the table
    await expect(page.getByText("Hello from E2E test")).toBeVisible({
      timeout: 10000,
    })
  })
})
