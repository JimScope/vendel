import { useMutation } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"

export interface TopUpResponse {
  provider: string
  payment_url?: string
  wallet_address?: string
}

export function useTopUp() {
  return useMutation({
    mutationFn: async (data: { amount: number; provider: string }) => {
      return (await pb.send("/api/plans/topup", {
        method: "POST",
        body: data,
      })) as TopUpResponse
    },
  })
}
