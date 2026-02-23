import { queryOptions, useQuery } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"

export const modemStatusQueryOptions = queryOptions({
  queryKey: ["modem-status"],
  queryFn: async () => {
    const response = await pb.send<{ online: Record<string, boolean> }>(
      "/api/sms/devices/status",
      { method: "GET" },
    )
    return response.online
  },
  staleTime: 30_000,
  refetchInterval: 30_000,
})

export function useModemStatus() {
  return useQuery(modemStatusQueryOptions)
}
