import { queryOptions, useQuery } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"
import type { UserBalance } from "@/types/collections"

export const balanceQueryOptions = queryOptions({
  queryKey: ["balance"],
  queryFn: async () => {
    return (await pb.send("/api/plans/balance", {})) as UserBalance
  },
  staleTime: 60_000,
})

export function useBalance() {
  return useQuery(balanceQueryOptions)
}
