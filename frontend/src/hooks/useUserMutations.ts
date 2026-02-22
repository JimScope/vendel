import { useMutation, useQueryClient } from "@tanstack/react-query"

import pb from "@/lib/pocketbase"
import useCustomToast from "./useCustomToast"

interface UserCreate {
  email: string
  password: string
  passwordConfirm: string
  full_name?: string
  is_superuser?: boolean
  is_active?: boolean
}

interface UserUpdate {
  email?: string
  full_name?: string
  is_superuser?: boolean
}

export function useCreateUser() {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (data: UserCreate) => {
      return await pb.collection("users").create(data)
    },
    onSuccess: () => {
      showSuccessToast("User created successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to create user")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] })
    },
  })
}

export function useUpdateUser(userId: string) {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (data: UserUpdate) => {
      return await pb.collection("users").update(userId, data)
    },
    onSuccess: () => {
      showSuccessToast("User updated successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to update user")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] })
    },
  })
}

export function useDeleteUser() {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (userId: string) => {
      return await pb.collection("users").delete(userId)
    },
    onSuccess: () => {
      showSuccessToast("The user was deleted successfully")
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to delete user")
    },
    onSettled: () => {
      queryClient.invalidateQueries()
    },
  })
}
