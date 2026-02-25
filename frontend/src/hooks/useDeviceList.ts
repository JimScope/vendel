import { queryOptions, useQuery, useSuspenseQuery } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"
import { useRealtimeQuery } from "./useRealtimeQuery"

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
  useRealtimeQuery("sms_devices", [["devices"], ["quota"]])
  return useQuery(deviceListQueryOptions)
}

export function useDeviceListSuspense() {
  useRealtimeQuery("sms_devices", [["devices"], ["quota"]])
  return useSuspenseQuery(deviceListQueryOptions)
}
