import { queryOptions, useQuery, useSuspenseQuery } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"
import { useRealtimeQuery } from "./useRealtimeQuery"

export const templateListQueryOptions = queryOptions({
  queryKey: ["templates"],
  queryFn: async () => {
    const result = await pb.collection("sms_templates").getList(1, 100, {
      sort: "-created",
    })
    return { data: result.items, count: result.totalItems }
  },
  staleTime: 60_000,
})

export function useTemplateList() {
  useRealtimeQuery("sms_templates", [["templates"]])
  return useQuery(templateListQueryOptions)
}

export function useTemplateListSuspense() {
  useRealtimeQuery("sms_templates", [["templates"]])
  return useSuspenseQuery(templateListQueryOptions)
}
