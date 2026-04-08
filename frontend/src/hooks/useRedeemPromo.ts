import { useMutation } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"

export interface RedeemPromoResponse {
  message: string
  amount: number
  new_balance: number
}

export function useRedeemPromo() {
  return useMutation({
    mutationFn: async (code: string) => {
      return (await pb.send("/api/promos/redeem", {
        method: "POST",
        body: { code },
      })) as RedeemPromoResponse
    },
  })
}
