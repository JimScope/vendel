import { expect, test } from "@playwright/test"
import { cleanupCollection } from "./utils/privateApi.ts"

test.describe("Integrations page", () => {
  test.describe.configure({ mode: "serial" })

  test.beforeEach(async ({ request }) => {
    await cleanupCollection(request, "api_keys")
  })

  test.afterEach(async ({ request }) => {
    await cleanupCollection(request, "api_keys")
  })

  test("Navigate to integrations page", async ({ page }) => {
    await page.goto("/integrations")
    await expect(
      page.getByRole("heading", { name: "Integrations" }),
    ).toBeVisible()
    await expect(
      page.getByText("Manage API keys for programmatic access to the API"),
    ).toBeVisible()
  })

  test("Empty state is shown when no API keys exist", async ({ page }) => {
    await page.goto("/integrations")
    await expect(page.getByText("No API keys")).toBeVisible()
    await expect(
      page.getByText("Create an API key to access the API programmatically"),
    ).toBeVisible()
  })

  test("Create an API key", async ({ page }) => {
    await page.goto("/integrations")

    await page.getByRole("button", { name: "Create API Key" }).click()
    await expect(
      page.getByRole("heading", { name: "Create API Key" }),
    ).toBeVisible()

    await page.getByLabel("Name *").fill("Test Key")
    await page.getByRole("button", { name: "Create" }).click()

    // Success state with key shown
    await expect(
      page.getByRole("heading", { name: "API Key Created" }),
    ).toBeVisible({ timeout: 10000 })
    await expect(
      page.getByText("Make sure to copy your API key now"),
    ).toBeVisible()

    // Close dialog
    await page.getByRole("button", { name: "Done" }).click()

    // Key should appear in table
    await expect(page.getByRole("cell", { name: "Test Key" })).toBeVisible()
    await expect(page.getByText("Active")).toBeVisible()
  })

  test("Rotate an API key", async ({ page }) => {
    await page.goto("/integrations")

    // Create a key first
    await page.getByRole("button", { name: "Create API Key" }).click()
    await page.getByLabel("Name *").fill("Rotate Me")
    await page.getByRole("button", { name: "Create" }).click()
    await expect(
      page.getByRole("heading", { name: "API Key Created" }),
    ).toBeVisible({ timeout: 10000 })
    await page.getByRole("button", { name: "Done" }).click()

    await expect(page.getByRole("cell", { name: "Rotate Me" })).toBeVisible()

    // Open actions menu and rotate
    await page
      .getByRole("row", { name: /Rotate Me/ })
      .getByRole("button")
      .click()
    await page.getByRole("menuitem", { name: "Rotate" }).click()

    await expect(
      page.getByRole("heading", { name: "Rotate API Key" }),
    ).toBeVisible()
    await page.getByRole("button", { name: "Rotate Key" }).click()

    // Success state
    await expect(
      page.getByRole("heading", { name: "Key Rotated" }),
    ).toBeVisible({ timeout: 10000 })
    await expect(page.getByText("Your old key has been revoked")).toBeVisible()

    await page.getByRole("button", { name: "Done" }).click()
  })

  test("Revoke an API key", async ({ page }) => {
    await page.goto("/integrations")

    // Create a key first
    await page.getByRole("button", { name: "Create API Key" }).click()
    await page.getByLabel("Name *").fill("Revoke Me")
    await page.getByRole("button", { name: "Create" }).click()
    await expect(
      page.getByRole("heading", { name: "API Key Created" }),
    ).toBeVisible({ timeout: 10000 })
    await page.getByRole("button", { name: "Done" }).click()

    await expect(page.getByRole("cell", { name: "Revoke Me" })).toBeVisible()

    // Open actions menu and revoke
    await page
      .getByRole("row", { name: /Revoke Me/ })
      .getByRole("button")
      .click()
    await page.getByRole("menuitem", { name: "Revoke" }).click()

    await expect(
      page.getByRole("heading", { name: "Revoke API Key" }),
    ).toBeVisible()
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Revoke" })
      .click()

    // Status should change to Revoked in the table cell
    await expect(
      page.getByRole("row", { name: /Revoke Me/ }).getByText("Revoked"),
    ).toBeVisible({ timeout: 10000 })
  })

  test("Delete an API key", async ({ page }) => {
    await page.goto("/integrations")

    // Create a key first
    await page.getByRole("button", { name: "Create API Key" }).click()
    await page.getByLabel("Name *").fill("Delete Me")
    await page.getByRole("button", { name: "Create" }).click()
    await expect(
      page.getByRole("heading", { name: "API Key Created" }),
    ).toBeVisible({ timeout: 10000 })
    await page.getByRole("button", { name: "Done" }).click()

    await expect(page.getByRole("cell", { name: "Delete Me" })).toBeVisible()

    // Open actions menu and delete
    await page
      .getByRole("row", { name: /Delete Me/ })
      .getByRole("button")
      .click()
    await page.getByRole("menuitem", { name: "Delete" }).click()

    await expect(
      page.getByRole("heading", { name: "Delete API Key" }),
    ).toBeVisible()
    await page
      .getByRole("dialog")
      .getByRole("button", { name: "Delete" })
      .click()

    // Key should be removed
    await expect(page.getByText("No API keys")).toBeVisible({ timeout: 10000 })
  })
})
