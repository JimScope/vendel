import { expect, test } from "@playwright/test"
import { cleanupCollection, resetUserQuota } from "./utils/privateApi.ts"

test.describe("Devices page", () => {
  test.describe.configure({ mode: "serial" })

  test.beforeEach(async ({ request }) => {
    await cleanupCollection(request, "sms_devices")
    await resetUserQuota(request)
  })

  test.afterEach(async ({ request }) => {
    await cleanupCollection(request, "sms_devices")
    await resetUserQuota(request)
  })

  test("Navigate to devices page", async ({ page }) => {
    await page.goto("/devices")
    await expect(page.getByRole("heading", { name: "Devices" })).toBeVisible()
    await expect(
      page.getByText("Manage your registered SMS devices"),
    ).toBeVisible()
  })

  test("Empty state is shown when no devices exist", async ({ page }) => {
    await page.goto("/devices")
    await expect(page.getByText("No devices registered")).toBeVisible()
    await expect(
      page.getByText("Add a device to start sending SMS messages"),
    ).toBeVisible()
  })

  test("Add a new device", async ({ page }) => {
    await page.goto("/devices")

    await page.getByRole("button", { name: "Add Device" }).click()
    await expect(
      page.getByRole("heading", { name: "Add Device" }),
    ).toBeVisible()

    await page.getByLabel("Name *").fill("Test Phone")
    await page.getByLabel("Phone Number *").fill("+15551234567")
    await page.getByRole("button", { name: "Create Device" }).click()

    await expect(
      page.getByRole("heading", { name: "Device Created" }),
    ).toBeVisible({ timeout: 10000 })
    await expect(page.getByText("Save the API key below")).toBeVisible()

    // Close the dialog
    await page.getByRole("button", { name: "Done" }).click()

    // Device should appear in table
    await expect(page.getByRole("cell", { name: "Test Phone" })).toBeVisible()
    await expect(page.getByRole("cell", { name: "+15551234567" })).toBeVisible()
  })

  test("Edit a device", async ({ page }) => {
    await page.goto("/devices")

    // First create a device
    await page.getByRole("button", { name: "Add Device" }).click()
    await page.getByLabel("Name *").fill("Original Name")
    await page.getByLabel("Phone Number *").fill("+15559876543")
    await page.getByRole("button", { name: "Create Device" }).click()
    await expect(
      page.getByRole("heading", { name: "Device Created" }),
    ).toBeVisible({ timeout: 10000 })
    await page.getByRole("button", { name: "Done" }).click()

    // Wait for table to render with the new device
    await expect(
      page.getByRole("cell", { name: "Original Name" }),
    ).toBeVisible()

    // Open actions menu and edit
    await page
      .getByRole("row", { name: /Original Name/ })
      .getByRole("button")
      .click()
    await page.getByRole("menuitem", { name: "Edit Device" }).click()

    await expect(
      page.getByRole("heading", { name: "Edit Device" }),
    ).toBeVisible()

    await page.getByLabel("Name *").fill("Updated Name")
    await page.getByRole("button", { name: "Save" }).click()

    await expect(page.getByRole("cell", { name: "Updated Name" })).toBeVisible()
  })

  test("Delete a device", async ({ page }) => {
    await page.goto("/devices")

    // First create a device
    await page.getByRole("button", { name: "Add Device" }).click()
    await page.getByLabel("Name *").fill("To Delete")
    await page.getByLabel("Phone Number *").fill("+15550001111")
    await page.getByRole("button", { name: "Create Device" }).click()
    await expect(
      page.getByRole("heading", { name: "Device Created" }),
    ).toBeVisible({ timeout: 10000 })
    await page.getByRole("button", { name: "Done" }).click()

    await expect(page.getByRole("cell", { name: "To Delete" })).toBeVisible()

    // Open actions menu and delete
    await page
      .getByRole("row", { name: /To Delete/ })
      .getByRole("button")
      .click()
    await page.getByRole("menuitem", { name: "Delete Device" }).click()

    await expect(
      page.getByRole("heading", { name: "Delete Device" }),
    ).toBeVisible()
    await page.getByRole("button", { name: "Delete" }).click()

    // Device should be removed
    await expect(page.getByText("No devices registered")).toBeVisible()
  })
})
