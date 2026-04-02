import { useMutation, useQueryClient } from "@tanstack/react-query"

import pb from "@/lib/pocketbase"
import useCustomToast from "./useCustomToast"

interface ContactGroupCreate {
  name: string
}

interface ContactGroupUpdate {
  name?: string
}

export function useCreateContactGroup() {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (data: ContactGroupCreate) => {
      return await pb.collection("contact_groups").create({
        ...data,
        user: pb.authStore.record?.id,
      })
    },
    onSuccess: () => {
      showSuccessToast("Group created successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to create group")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["contact-groups"] })
    },
  })
}

export function useUpdateContactGroup(groupId: string) {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (data: ContactGroupUpdate) => {
      return await pb.collection("contact_groups").update(groupId, data)
    },
    onSuccess: () => {
      showSuccessToast("Group updated successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to update group")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["contact-groups"] })
    },
  })
}

export function useDeleteContactGroup() {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (groupId: string) => {
      return await pb.collection("contact_groups").delete(groupId)
    },
    onSuccess: () => {
      showSuccessToast("Group deleted successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to delete group")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["contact-groups"] })
    },
  })
}
