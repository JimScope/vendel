import { useMutation, useQueryClient } from "@tanstack/react-query"

import pb from "@/lib/pocketbase"
import useCustomToast from "./useCustomToast"

interface ContactCreate {
  name: string
  phone_number: string
  groups?: string[]
  notes?: string
}

interface ContactUpdate {
  name?: string
  phone_number?: string
  groups?: string[]
  notes?: string
}

export function useCreateContact() {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (data: ContactCreate) => {
      return await pb.collection("contacts").create({
        ...data,
        user: pb.authStore.record?.id,
      })
    },
    onSuccess: () => {
      showSuccessToast("Contact created successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to create contact")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["contacts"] })
    },
  })
}

export function useUpdateContact(contactId: string) {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (data: ContactUpdate) => {
      return await pb.collection("contacts").update(contactId, data)
    },
    onSuccess: () => {
      showSuccessToast("Contact updated successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to update contact")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["contacts"] })
    },
  })
}

export function useDeleteContact() {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (contactId: string) => {
      return await pb.collection("contacts").delete(contactId)
    },
    onSuccess: () => {
      showSuccessToast("Contact deleted successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to delete contact")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["contacts"] })
    },
  })
}
