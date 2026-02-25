function extractErrorMessage(err: unknown): string {
  if (err instanceof Error) {
    // PocketBase ClientResponseError has response.data
    const pbError = err as {
      response?: { data?: Record<string, unknown>; message?: string }
    }
    if (pbError.response?.data) {
      const data = pbError.response.data
      // Handle object with message property (e.g., quota errors)
      if (data.message && typeof data.message === "string") {
        return data.message
      }
      // Handle detail field
      if (data.detail && typeof data.detail === "string") {
        return data.detail
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

export const getInitials = (name: string): string => {
  return name
    .split(" ")
    .slice(0, 2)
    .map((word) => word[0])
    .join("")
    .toUpperCase()
}
