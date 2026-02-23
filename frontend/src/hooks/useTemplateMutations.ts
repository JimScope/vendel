import { useMutation, useQueryClient } from "@tanstack/react-query"

import pb from "@/lib/pocketbase"
import useCustomToast from "./useCustomToast"

interface TemplateCreate {
  name: string
  body: string
}

interface TemplateUpdate {
  name?: string
  body?: string
}

export function useCreateTemplate() {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (data: TemplateCreate) => {
      return await pb.collection("sms_templates").create({
        ...data,
        user: pb.authStore.record?.id,
      })
    },
    onSuccess: () => {
      showSuccessToast("Template created successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to create template")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["templates"] })
    },
  })
}

export function useUpdateTemplate(templateId: string) {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (data: TemplateUpdate) => {
      return await pb.collection("sms_templates").update(templateId, data)
    },
    onSuccess: () => {
      showSuccessToast("Template updated successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to update template")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["templates"] })
    },
  })
}

export function useDeleteTemplate() {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (templateId: string) => {
      return await pb.collection("sms_templates").delete(templateId)
    },
    onSuccess: () => {
      showSuccessToast("The template was deleted successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to delete template")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["templates"] })
    },
  })
}
