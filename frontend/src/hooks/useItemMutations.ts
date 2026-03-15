import { useMutation, useQueryClient } from "@tanstack/react-query"

import pb from "@/lib/pocketbase"
import useCustomToast from "./useCustomToast"

interface ItemCreate {
  title: string
  description?: string
}

interface ItemUpdate {
  title?: string
  description?: string
}

export function useCreateItem() {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (data: ItemCreate) => {
      return await pb.collection("items").create(data)
    },
    onSuccess: () => {
      showSuccessToast("Item created successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to create item")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["items"] })
    },
  })
}

export function useUpdateItem(itemId: string) {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (data: ItemUpdate) => {
      return await pb.collection("items").update(itemId, data)
    },
    onSuccess: () => {
      showSuccessToast("Item updated successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to update item")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["items"] })
    },
  })
}

export function useDeleteItem() {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (itemId: string) => {
      return await pb.collection("items").delete(itemId)
    },
    onSuccess: () => {
      showSuccessToast("The item was deleted successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to delete item")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["items"] })
    },
  })
}
