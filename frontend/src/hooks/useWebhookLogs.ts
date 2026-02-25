import { queryOptions, useQuery, useSuspenseQuery } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"

export const webhookLogsQueryOptions = (webhookId: string) =>
  queryOptions({
    queryKey: ["webhook-logs", webhookId],
    queryFn: async () => {
      const result = await pb
        .collection("webhook_delivery_logs")
        .getList(1, 50, {
          filter: `webhook = "${webhookId}"`,
          sort: "-created",
        })
      return { data: result.items, count: result.totalItems }
    },
    staleTime: 30_000,
    enabled: !!webhookId,
  })

export function useWebhookLogs(webhookId: string) {
  return useQuery(webhookLogsQueryOptions(webhookId))
}

export function useWebhookLogsSuspense(webhookId: string) {
  return useSuspenseQuery(webhookLogsQueryOptions(webhookId))
}
