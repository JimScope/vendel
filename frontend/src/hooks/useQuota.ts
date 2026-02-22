import { queryOptions, useQuery, useSuspenseQuery } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"
import { useRealtimeQuery } from "./useRealtimeQuery"

export const quotaQueryOptions = queryOptions({
  queryKey: ["quota"],
  queryFn: async () => {
    return await pb.send("/api/plans/quota", {})
  },
  staleTime: 60_000,
})

export function useQuota() {
  useRealtimeQuery("user_quotas", [["quota"]])
  return useQuery(quotaQueryOptions)
}

export function useQuotaSuspense() {
  useRealtimeQuery("user_quotas", [["quota"]])
  return useSuspenseQuery(quotaQueryOptions)
}
