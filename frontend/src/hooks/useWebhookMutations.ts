import { useMutation, useQueryClient } from "@tanstack/react-query"

import pb from "@/lib/pocketbase"
import useCustomToast from "./useCustomToast"

interface WebhookConfigCreate {
  url: string
  secret_key?: string
  events?: string[]
  active?: boolean
}

interface WebhookConfigUpdate {
  url?: string
  secret_key?: string
  events?: string[]
  active?: boolean
}

export function useCreateWebhook() {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (data: WebhookConfigCreate) => {
      return await pb.collection("webhook_configs").create({
        ...data,
        user: pb.authStore.record?.id,
      })
    },
    onSuccess: () => {
      showSuccessToast("Webhook created successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to create webhook")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["webhooks"] })
    },
  })
}

export function useUpdateWebhook(webhookId: string) {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (data: WebhookConfigUpdate) => {
      return await pb.collection("webhook_configs").update(webhookId, data)
    },
    onSuccess: () => {
      showSuccessToast("Webhook updated successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to update webhook")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["webhooks"] })
    },
  })
}

export function useDeleteWebhook() {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (webhookId: string) => {
      return await pb.collection("webhook_configs").delete(webhookId)
    },
    onSuccess: () => {
      showSuccessToast("The webhook was deleted successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to delete webhook")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["webhooks"] })
    },
  })
}
