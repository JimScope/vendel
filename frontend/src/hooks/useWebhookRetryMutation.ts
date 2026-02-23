import { useMutation, useQueryClient } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"
import useCustomToast from "./useCustomToast"

interface RetryWebhookResult {
  delivery_status: string
  response_status: number
  duration_ms: number
  error_message: string
  log_id?: string
}

export function useWebhookRetryMutation(webhookId: string) {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (logId: string): Promise<RetryWebhookResult> => {
      const response = await pb.send("/api/webhooks/retry", {
        method: "POST",
        body: { log_id: logId },
      })
      return response as RetryWebhookResult
    },
    onSuccess: (result) => {
      if (result.delivery_status === "success") {
        showSuccessToast(
          `Retry succeeded (HTTP ${result.response_status}, ${result.duration_ms}ms)`,
        )
      } else {
        showErrorToast(
          result.error_message ||
            `Retry failed (HTTP ${result.response_status})`,
        )
      }
      queryClient.invalidateQueries({ queryKey: ["webhook-logs", webhookId] })
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to retry webhook")
    },
  })
}
