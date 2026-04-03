import { useMutation, useQueryClient } from "@tanstack/react-query"

import pb from "@/lib/pocketbase"
import useCustomToast from "./useCustomToast"

interface SMSMessageCreate {
  recipients: string[]
  body: string
  device_id?: string
}

interface SMSTemplateCreate {
  recipients: string[]
  template_id: string
  variables?: Record<string, string>
  device_id?: string
}

export function useSendSMS() {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (data: SMSMessageCreate) => {
      return await pb.send("/api/sms/send", {
        method: "POST",
        body: data,
      })
    },
    onSuccess: () => {
      showSuccessToast("SMS sent successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to send SMS")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["sms"] })
      queryClient.invalidateQueries({ queryKey: ["quota"] })
    },
  })
}

export function useSendSMSTemplate() {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (data: SMSTemplateCreate) => {
      return await pb.send("/api/sms/send-template", {
        method: "POST",
        body: data,
      })
    },
    onSuccess: () => {
      showSuccessToast("SMS sent successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to send SMS")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["sms"] })
      queryClient.invalidateQueries({ queryKey: ["quota"] })
    },
  })
}
