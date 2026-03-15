import { useMutation } from "@tanstack/react-query"

import pb from "@/lib/pocketbase"
import useCustomToast from "./useCustomToast"

export interface UpgradeResponse {
  status: "activated" | "pending_payment" | "pending_authorization"
  plan: string
  message: string
  payment_url?: string
  authorization_url?: string
}

interface UpgradePlanInput {
  plan_id: string
  provider: string | null
}

export function useUpgradePlan() {
  const { showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (data: UpgradePlanInput) => {
      return (await pb.send("/api/plans/upgrade", {
        method: "POST",
        body: { plan_id: data.plan_id, provider: data.provider },
      })) as UpgradeResponse
    },
    onError: () => {
      showErrorToast("Failed to upgrade plan. Please try again.")
    },
  })
}
