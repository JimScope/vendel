import { useMutation, useQueryClient } from "@tanstack/react-query"

import pb from "@/lib/pocketbase"
import useCustomToast from "./useCustomToast"

interface ScheduledSMSCreate {
  name: string
  recipients: string[]
  body: string
  device_id?: string
  schedule_type: "one_time" | "recurring"
  scheduled_at?: string
  cron_expression?: string
  timezone?: string
}

interface ScheduledSMSUpdate {
  name?: string
  recipients?: string[]
  body?: string
  device_id?: string
  schedule_type?: "one_time" | "recurring"
  scheduled_at?: string
  cron_expression?: string
  timezone?: string
  status?: "active" | "paused" | "completed"
}

export function useCreateScheduledSMS() {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (data: ScheduledSMSCreate) => {
      return await pb.collection("scheduled_sms").create({
        ...data,
        user: pb.authStore.record?.id,
        status: "active",
      })
    },
    onSuccess: () => {
      showSuccessToast("Scheduled SMS created successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to create scheduled SMS")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["scheduled-sms"] })
    },
  })
}

export function useUpdateScheduledSMS(scheduleId: string) {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (data: ScheduledSMSUpdate) => {
      return await pb.collection("scheduled_sms").update(scheduleId, data)
    },
    onSuccess: () => {
      showSuccessToast("Scheduled SMS updated successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to update scheduled SMS")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["scheduled-sms"] })
    },
  })
}

export function useDeleteScheduledSMS() {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (scheduleId: string) => {
      return await pb.collection("scheduled_sms").delete(scheduleId)
    },
    onSuccess: () => {
      showSuccessToast("The scheduled SMS was deleted successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to delete scheduled SMS")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["scheduled-sms"] })
    },
  })
}
