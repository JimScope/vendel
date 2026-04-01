import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import pb from "@/lib/pocketbase"

export function useSystemConfig() {
  return useQuery({
    queryKey: ["system-config"],
    queryFn: () => pb.send("/api/system-config", {}),
  })
}

export function useSaveSystemConfig() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (configs: Record<string, string>) => {
      return await pb.send("/api/system-config", {
        method: "PUT",
        body: configs,
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["system-config"] })
      queryClient.invalidateQueries({ queryKey: ["app-settings"] })
    },
  })
}
