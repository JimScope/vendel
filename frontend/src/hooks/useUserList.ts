import { queryOptions, useQuery, useSuspenseQuery } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"

export const userListQueryOptions = queryOptions({
  queryKey: ["users"],
  queryFn: async () => {
    const result = await pb.collection("users").getList(1, 100, {
      sort: "-created",
    })
    return { data: result.items, count: result.totalItems }
  },
  staleTime: 60_000,
})

export function useUserList() {
  return useQuery(userListQueryOptions)
}

export function useUserListSuspense() {
  return useSuspenseQuery(userListQueryOptions)
}
