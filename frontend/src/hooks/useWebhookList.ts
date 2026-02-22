import { queryOptions, useQuery, useSuspenseQuery } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"
import { useRealtimeQuery } from "./useRealtimeQuery"

export const webhookListQueryOptions = queryOptions({
  queryKey: ["webhooks"],
  queryFn: async () => {
    const result = await pb.collection("webhook_configs").getList(1, 100, {
      sort: "-created",
    })
    return { data: result.items, count: result.totalItems }
  },
  staleTime: 60_000,
})

export function useWebhookList() {
  useRealtimeQuery("webhook_configs", [["webhooks"]])
  return useQuery(webhookListQueryOptions)
}

export function useWebhookListSuspense() {
  useRealtimeQuery("webhook_configs", [["webhooks"]])
  return useSuspenseQuery(webhookListQueryOptions)
}
