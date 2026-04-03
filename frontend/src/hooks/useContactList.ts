import { queryOptions, useQuery, useSuspenseQuery } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"
import { useRealtimeQuery } from "./useRealtimeQuery"

export const contactListQueryOptions = queryOptions({
  queryKey: ["contacts"],
  queryFn: async () => {
    const result = await pb.collection("contacts").getList(1, 100, {
      sort: "-created",
      expand: "groups",
    })
    return { data: result.items, count: result.totalItems }
  },
  staleTime: 60_000,
})

export function useContactList() {
  useRealtimeQuery("contacts", [["contacts"]])
  return useQuery(contactListQueryOptions)
}

export function useContactListSuspense() {
  useRealtimeQuery("contacts", [["contacts"]])
  return useSuspenseQuery(contactListQueryOptions)
}
