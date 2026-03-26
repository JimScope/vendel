import { useMutation } from "@tanstack/react-query"

import pb from "@/lib/pocketbase"
import useCustomToast from "./useCustomToast"

export interface UpgradeResponse {
  subscription_id: string
  status: "active" | "pending"
  message: string
  payment_url?: string
  wallet_address?: string
}

interface UpgradePlanInput {
  plan_id: string
  billing_cycle: "monthly" | "yearly"
  provider: string | null
}

export function useUpgradePlan() {
  const { showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (data: UpgradePlanInput) => {
      return (await pb.send("/api/plans/upgrade", {
        method: "PUT",
        body: {
          plan_id: data.plan_id,
          billing_cycle: data.billing_cycle,
          provider: data.provider,
        },
      })) as UpgradeResponse
    },
    onError: () => {
      showErrorToast("Failed to upgrade plan. Please try again.")
    },
  })
}
