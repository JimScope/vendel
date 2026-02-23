import { useMutation, useQueryClient } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"
import useCustomToast from "./useCustomToast"

interface TestWebhookResult {
  delivery_status: string
  response_status: number
  duration_ms: number
  error_message: string
  log_id?: string
}

export function useTestWebhook() {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (webhookId: string): Promise<TestWebhookResult> => {
      const response = await pb.send("/api/webhooks/test", {
        method: "POST",
        body: { webhook_id: webhookId },
      })
      return response as TestWebhookResult
    },
    onSuccess: (result, webhookId) => {
      if (result.delivery_status === "success") {
        showSuccessToast(
          `Webhook test succeeded (HTTP ${result.response_status}, ${result.duration_ms}ms)`,
        )
      } else {
        showErrorToast(
          result.error_message ||
            `Webhook test failed (HTTP ${result.response_status})`,
        )
      }
      queryClient.invalidateQueries({ queryKey: ["webhook-logs", webhookId] })
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to test webhook")
    },
  })
}
