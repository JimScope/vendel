import { useMutation } from "@tanstack/react-query"

import pb from "@/lib/pocketbase"

export interface UpgradeResponse {
  subscription_id: string
  status: string
  message: string
}

interface UpgradePlanInput {
  plan_id: string
  billing_cycle: "monthly" | "yearly"
}

export function useUpgradePlan() {
  return useMutation({
    mutationFn: async (data: UpgradePlanInput) => {
      return (await pb.send("/api/plans/upgrade", {
        method: "PUT",
        body: {
          plan_id: data.plan_id,
          billing_cycle: data.billing_cycle,
        },
      })) as UpgradeResponse
    },
  })
}
