import { queryOptions, useQuery, useSuspenseQuery } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"

export const planListQueryOptions = queryOptions({
  queryKey: ["plans"],
  queryFn: async () => {
    const result = await pb.collection("user_plans").getList(1, 100, {
      filter: "is_public = true",
      sort: "price",
    })
    return { data: result.items, count: result.totalItems }
  },
  staleTime: 300_000, // 5 minutes - plans don't change often
})

export function usePlanList() {
  return useQuery(planListQueryOptions)
}

export function usePlanListSuspense() {
  return useSuspenseQuery(planListQueryOptions)
}
