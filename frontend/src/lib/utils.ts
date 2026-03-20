import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatDate(dateString: string | null): string {
  if (!dateString) return "Never"
  return new Date(dateString).toLocaleString()
}

// ── Error handling ──────────────────────────────────────────────────

function extractErrorMessage(err: unknown): string {
  if (err instanceof Error) {
    // PocketBase ClientResponseError has response.data
    const pbError = err as {
      response?: { data?: Record<string, unknown>; message?: string }
    }
    if (pbError.response?.data) {
      const data = pbError.response.data
      if (data.message && typeof data.message === "string") {
        return data.message
      }
      if (data.detail && typeof data.detail === "string") {
        return data.detail
      }
      // PocketBase field-level validation errors: { fieldName: { code, message } }
      const fieldError = Object.values(data).find(
        (v): v is { message: string } =>
          typeof v === "object" &&
          v !== null &&
          "message" in v &&
          typeof (v as { message: unknown }).message === "string",
      )
      if (fieldError) {
        return fieldError.message
      }
    }
    if (pbError.response?.message) {
      return pbError.response.message
    }
    return err.message
  }

  return "Something went wrong."
}

export const handleError = function (
  this: (msg: string) => void,
  err: unknown,
) {
  const errorMessage = extractErrorMessage(err)
  this(errorMessage)
}

// ── Misc helpers ────────────────────────────────────────────────────

export const getInitials = (name: string): string => {
  return name
    .split(" ")
    .slice(0, 2)
    .map((word) => word[0])
    .join("")
    .toUpperCase()
}
