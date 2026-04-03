import { queryOptions, useQuery, useSuspenseQuery } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"
import { useRealtimeQuery } from "./useRealtimeQuery"

export const contactGroupListQueryOptions = queryOptions({
  queryKey: ["contact-groups"],
  queryFn: async () => {
    const result = await pb.collection("contact_groups").getList(1, 100, {
      sort: "name",
    })
    return { data: result.items, count: result.totalItems }
  },
  staleTime: 60_000,
})

export function useContactGroupList() {
  useRealtimeQuery("contact_groups", [["contact-groups"]])
  return useQuery(contactGroupListQueryOptions)
}

export function useContactGroupListSuspense() {
  useRealtimeQuery("contact_groups", [["contact-groups"]])
  return useSuspenseQuery(contactGroupListQueryOptions)
}
