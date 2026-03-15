import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import pb from "@/lib/pocketbase"
import useCustomToast from "./useCustomToast"

export function useSystemConfig() {
  return useQuery({
    queryKey: ["system-config"],
    queryFn: () => pb.send("/api/system-config", {}),
  })
}

export function useUpdateSystemConfig() {
  const queryClient = useQueryClient()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async ({ key, value }: { key: string; value: string }) => {
      return await pb.send(`/api/system-config/${key}`, {
        method: "PATCH",
        body: { value },
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["system-config"] })
      showSuccessToast("Configuration updated successfully")
    },
    onError: () => {
      showErrorToast("Failed to update configuration")
    },
  })
}
