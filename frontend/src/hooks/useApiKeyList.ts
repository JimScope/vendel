import { queryOptions, useQuery, useSuspenseQuery } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"
import { useRealtimeQuery } from "./useRealtimeQuery"

export const apiKeyListQueryOptions = queryOptions({
  queryKey: ["api-keys"],
  queryFn: async () => {
    const result = await pb.collection("api_keys").getList(1, 100, {
      sort: "-created",
    })
    return { data: result.items, count: result.totalItems }
  },
  staleTime: 60_000,
})

export function useApiKeyList() {
  useRealtimeQuery("api_keys", [["api-keys"]])
  return useQuery(apiKeyListQueryOptions)
}

export function useApiKeyListSuspense() {
  useRealtimeQuery("api_keys", [["api-keys"]])
  return useSuspenseQuery(apiKeyListQueryOptions)
}
