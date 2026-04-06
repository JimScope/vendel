import { queryOptions, useQuery, useSuspenseQuery } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"
import { useRealtimeQuery } from "./useRealtimeQuery"

export const contactGroupListQueryOptions = queryOptions({
  queryKey: ["contact-groups"],
  queryFn: async () => {
    const items = await pb.collection("contact_groups").getFullList({
      sort: "name",
    })
    return { data: items, count: items.length }
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
