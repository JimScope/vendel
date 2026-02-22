import { queryOptions, useQuery, useSuspenseQuery } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"

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
  return useQuery(apiKeyListQueryOptions)
}

export function useApiKeyListSuspense() {
  return useSuspenseQuery(apiKeyListQueryOptions)
}
