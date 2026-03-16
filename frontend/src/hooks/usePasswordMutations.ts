import { useMutation } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"

export function useRecoverPassword() {
  return useMutation({
    mutationFn: (email: string) =>
      pb.collection("users").requestPasswordReset(email),
  })
}

export function useResetPassword() {
  return useMutation({
    mutationFn: (data: { token: string; newPassword: string }) =>
      pb
        .collection("users")
        .confirmPasswordReset(data.token, data.newPassword, data.newPassword),
  })
}
