import { queryOptions, useQuery, useSuspenseQuery } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"

export const quotaQueryOptions = queryOptions({
  queryKey: ["quota"],
  queryFn: async () => {
    return await pb.send("/api/plans/quota", {})
  },
  staleTime: 60_000,
})

export function useQuota() {
  return useQuery(quotaQueryOptions)
}

export function useQuotaSuspense() {
  return useSuspenseQuery(quotaQueryOptions)
}
