import { useMutation, useQueryClient } from "@tanstack/react-query"

import pb from "@/lib/pocketbase"
import useCustomToast from "./useCustomToast"

interface ProfileUpdate {
  full_name?: string
  email?: string
}

interface PasswordChange {
  current_password: string
  new_password: string
  confirm_password: string
}

export function useUpdateProfile() {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (data: ProfileUpdate) => {
      const userId = pb.authStore.record?.id
      if (!userId) throw new Error("Not authenticated")
      return await pb.collection("users").update(userId, data)
    },
    onSuccess: () => {
      showSuccessToast("User updated successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to update profile")
    },
    onSettled: () => {
      queryClient.invalidateQueries()
    },
  })
}

export function useChangePassword() {
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (data: PasswordChange) => {
      const userId = pb.authStore.record?.id
      if (!userId) throw new Error("Not authenticated")
      return await pb.collection("users").update(userId, {
        oldPassword: data.current_password,
        password: data.new_password,
        passwordConfirm: data.confirm_password,
      })
    },
    onSuccess: () => {
      showSuccessToast("Password updated successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to change password")
    },
  })
}

export function useDeleteAccount() {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async () => {
      const userId = pb.authStore.record?.id
      if (!userId) throw new Error("Not authenticated")
      return await pb.collection("users").delete(userId)
    },
    onSuccess: () => {
      showSuccessToast("Your account has been successfully deleted")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to delete account")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["currentUser"] })
    },
  })
}

export function useExportData() {
  const { showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async () => {
      return await pb.send("/api/user/export", { method: "GET" })
    },
    onError: () => {
      showErrorToast("Failed to export data")
    },
  })
}
