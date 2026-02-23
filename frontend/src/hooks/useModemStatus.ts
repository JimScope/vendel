import { useEffect, useRef } from "react"
import { queryOptions, useQuery, useQueryClient } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"

export const modemStatusQueryOptions = queryOptions({
  queryKey: ["modem-status"],
  queryFn: async () => {
    const response = await pb.send<{ online: Record<string, boolean> }>(
      "/api/sms/devices/status",
      { method: "GET" },
    )
    return response.online
  },
  staleTime: Infinity,
})

export function useModemStatus() {
  const queryClient = useQueryClient()
  const unsubRef = useRef<(() => Promise<void>) | null>(null)

  useEffect(() => {
    if (!pb.authStore.isValid) return

    pb.realtime
      .subscribe("modem-status", (data) => {
        queryClient.setQueryData(["modem-status"], data)
      })
      .then((unsub) => {
        unsubRef.current = unsub
      })

    return () => {
      unsubRef.current?.()
    }
  }, [queryClient])

  return useQuery(modemStatusQueryOptions)
}
