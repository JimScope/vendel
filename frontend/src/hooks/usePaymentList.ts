import { queryOptions, useQuery, useSuspenseQuery } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"
import { useRealtimeQuery } from "./useRealtimeQuery"

export const paymentListQueryOptions = queryOptions({
  queryKey: ["payments"],
  queryFn: async () => {
    const result = await pb.collection("payments").getList(1, 200, {
      sort: "-created",
      expand: "subscription",
    })
    return { data: result.items, count: result.totalItems }
  },
  staleTime: 60_000,
})

export function usePaymentList() {
  useRealtimeQuery("payments", [["payments"]])
  return useQuery(paymentListQueryOptions)
}

export function usePaymentListSuspense() {
  useRealtimeQuery("payments", [["payments"]])
  return useSuspenseQuery(paymentListQueryOptions)
}
