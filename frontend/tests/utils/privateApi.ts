import type { APIRequestContext } from "@playwright/test"
import { firstSuperuser, firstSuperuserPassword } from "../config.ts"

const API_URL = process.env.VITE_API_URL || "http://localhost:8090"

async function getSuperuserToken(request: APIRequestContext): Promise<string> {
  const response = await request.post(
    `${API_URL}/api/collections/_superusers/auth-with-password`,
    {
      data: {
        identity: firstSuperuser,
        password: firstSuperuserPassword,
      },
    },
  )

  if (!response.ok()) {
    throw new Error(`Failed to authenticate superuser: ${response.status()}`)
  }

  const body = await response.json()
  return body.token as string
}

export async function createUser(
  request: APIRequestContext,
  data: { email: string; password: string; name?: string },
) {
  const token = await getSuperuserToken(request)

  const response = await request.post(
    `${API_URL}/api/collections/users/records`,
    {
      headers: { Authorization: `Bearer ${token}` },
      data: {
        email: data.email,
        password: data.password,
        passwordConfirm: data.password,
        full_name: data.name ?? "Test User",
        verified: true,
      },
    },
  )

  if (!response.ok()) {
    const body = await response.text()
    throw new Error(`Failed to create user: ${response.status()} ${body}`)
  }

  return await response.json()
}

export async function createTestDevice(
  request: APIRequestContext,
  data: { name: string; phone_number: string; user_id: string },
) {
  const token = await getSuperuserToken(request)

  const response = await request.post(
    `${API_URL}/api/collections/sms_devices/records`,
    {
      headers: { Authorization: `Bearer ${token}` },
      data: {
        name: data.name,
        phone_number: data.phone_number,
        user: data.user_id,
        device_type: "android",
      },
    },
  )

  if (!response.ok()) {
    const body = await response.text()
    throw new Error(`Failed to create device: ${response.status()} ${body}`)
  }

  return await response.json()
}

export async function deleteRecord(
  request: APIRequestContext,
  collection: string,
  id: string,
) {
  const token = await getSuperuserToken(request)

  const response = await request.delete(
    `${API_URL}/api/collections/${collection}/records/${id}`,
    {
      headers: { Authorization: `Bearer ${token}` },
    },
  )

  // Ignore 404 (already deleted) errors
  if (!response.ok() && response.status() !== 404) {
    const body = await response.text()
    throw new Error(
      `Failed to delete ${collection}/${id}: ${response.status()} ${body}`,
    )
  }
}

export async function listRecords(
  request: APIRequestContext,
  collection: string,
  filter?: string,
): Promise<{ items: Array<{ id: string; [key: string]: unknown }> }> {
  const token = await getSuperuserToken(request)

  const params = new URLSearchParams()
  if (filter) params.set("filter", filter)

  const response = await request.get(
    `${API_URL}/api/collections/${collection}/records?${params.toString()}`,
    {
      headers: { Authorization: `Bearer ${token}` },
    },
  )

  if (!response.ok()) {
    const body = await response.text()
    throw new Error(
      `Failed to list ${collection}: ${response.status()} ${body}`,
    )
  }

  return await response.json()
}

export async function getCurrentUserId(
  request: APIRequestContext,
): Promise<string> {
  const token = await getSuperuserToken(request)

  const response = await request.get(
    `${API_URL}/api/collections/users/records?filter=(email='${firstSuperuser}')`,
    {
      headers: { Authorization: `Bearer ${token}` },
    },
  )

  if (!response.ok()) {
    throw new Error(`Failed to get current user: ${response.status()}`)
  }

  const body = await response.json()
  return body.items[0].id as string
}

export async function cleanupCollection(
  request: APIRequestContext,
  collection: string,
) {
  const records = await listRecords(request, collection)
  for (const record of records.items) {
    await deleteRecord(request, collection, record.id)
  }
}

export async function resetUserQuota(
  request: APIRequestContext,
): Promise<void> {
  const token = await getSuperuserToken(request)
  const userId = await getCurrentUserId(request)

  const quotas = await listRecords(request, "user_quotas", `user='${userId}'`)
  for (const quota of quotas.items) {
    // Reset counters
    await request.patch(
      `${API_URL}/api/collections/user_quotas/records/${quota.id}`,
      {
        headers: { Authorization: `Bearer ${token}` },
        data: { devices_registered: 0, sms_sent_this_month: 0 },
      },
    )

    // Bump the plan's max_devices to avoid quota conflicts in parallel tests
    const planId = quota.plan as string
    if (planId) {
      await request.patch(
        `${API_URL}/api/collections/user_plans/records/${planId}`,
        {
          headers: { Authorization: `Bearer ${token}` },
          data: { max_devices: 10, max_sms_per_month: 100 },
        },
      )
    }
  }
}
