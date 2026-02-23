import { useEffect, useRef } from "react"
import { useQuery, useQueryClient } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"

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

  return useQuery<Record<string, boolean>>({
    queryKey: ["modem-status"],
    queryFn: () => ({}),
    staleTime: Infinity,
  })
}
