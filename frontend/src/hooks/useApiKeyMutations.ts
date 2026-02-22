import { useMutation, useQueryClient } from "@tanstack/react-query"

import pb from "@/lib/pocketbase"
import useCustomToast from "./useCustomToast"

interface ApiKeyCreate {
  name: string
}

export function useCreateApiKey() {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (data: ApiKeyCreate) => {
      return await pb.collection("api_keys").create({
        ...data,
        user: pb.authStore.record?.id,
        is_active: true,
      })
    },
    onSuccess: () => {
      showSuccessToast("API key created successfully")
      queryClient.invalidateQueries({ queryKey: ["api-keys"] })
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to create API key")
    },
  })
}

export function useDeleteApiKey() {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (apiKeyId: string) => {
      return await pb.collection("api_keys").delete(apiKeyId)
    },
    onSuccess: () => {
      showSuccessToast("API key deleted successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to delete API key")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["api-keys"] })
    },
  })
}

export function useRevokeApiKey() {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (apiKeyId: string) => {
      return await pb.collection("api_keys").update(apiKeyId, {
        is_active: false,
      })
    },
    onSuccess: () => {
      showSuccessToast("API key revoked successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to revoke API key")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["api-keys"] })
    },
  })
}
