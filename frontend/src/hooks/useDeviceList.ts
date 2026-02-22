import { queryOptions, useQuery, useSuspenseQuery } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"

export const deviceListQueryOptions = queryOptions({
  queryKey: ["devices"],
  queryFn: async () => {
    const result = await pb.collection("sms_devices").getList(1, 100, {
      sort: "-created",
    })
    return { data: result.items, count: result.totalItems }
  },
  staleTime: 60_000,
})

export function useDeviceList() {
  return useQuery(deviceListQueryOptions)
}

export function useDeviceListSuspense() {
  return useSuspenseQuery(deviceListQueryOptions)
}
