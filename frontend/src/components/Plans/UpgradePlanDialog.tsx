import { useQueryClient } from "@tanstack/react-query"
import { Check, ExternalLink } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { LoadingButton } from "@/components/ui/loading-button"
import { Skeleton } from "@/components/ui/skeleton"
import useAppConfig from "@/hooks/useAppConfig"
import useCustomToast from "@/hooks/useCustomToast"
import { usePlanList } from "@/hooks/usePlanList"
import { type UpgradeResponse, useUpgradePlan } from "@/hooks/usePlanMutations"
import type { Plan } from "@/types/collections"

interface UpgradePlanDialogProps {
  currentPlan?: string
}

function PlanCard({
  plan,
  isCurrentPlan,
  isSelected,
  onSelect,
}: {
  plan: Plan
  isCurrentPlan: boolean
  isSelected: boolean
  onSelect: () => void
}) {
  const { t } = useTranslation()
  return (
    <Card
      className={`relative cursor-pointer transition-colors hover:bg-accent/50 ${
        isSelected ? "border-primary ring-1 ring-primary" : ""
      } ${isCurrentPlan ? "bg-accent/30" : ""}`}
      onClick={onSelect}
    >
      {isCurrentPlan && (
        <div className="absolute -top-3 left-4 bg-primary text-primary-foreground text-xs px-2 py-1 rounded">
          {t("plans.currentPlan")}
        </div>
      )}
      <CardHeader className="pb-2">
        <CardTitle className="capitalize">{plan.name}</CardTitle>
        <CardDescription>
          <span className="text-2xl font-bold text-foreground">
            ${plan.price?.toFixed(2) ?? "0.00"}
          </span>
          <span className="text-muted-foreground">{t("plans.perMonth")}</span>
        </CardDescription>
      </CardHeader>
      <CardContent>
        <ul className="space-y-2 text-sm">
          <li className="flex items-center gap-2">
            <Check className="h-4 w-4 text-primary" />
            {t("plans.smsPerMonth", { count: plan.max_sms_per_month ?? 0 })}
          </li>
          <li className="flex items-center gap-2">
            <Check className="h-4 w-4 text-primary" />
            {t("plans.device", { count: plan.max_devices ?? 0 })}
          </li>
        </ul>
      </CardContent>
    </Card>
  )
}

function UpgradePlanDialog({ currentPlan }: UpgradePlanDialogProps) {
  const { t } = useTranslation()
  const [isOpen, setIsOpen] = useState(false)
  const [selectedPlanId, setSelectedPlanId] = useState<string | null>(null)
  const [selectedProvider, setSelectedProvider] = useState<string | null>(null)
  const [pendingPayment, setPendingPayment] = useState<{
    url: string
    plan: string
    providerDisplayName: string
  } | null>(null)
  const { data: plansData, isLoading } = usePlanList()
  const { config } = useAppConfig()
  const queryClient = useQueryClient()
  const { showSuccessToast, showInfoToast } = useCustomToast()
  const mutation = useUpgradePlan()

  const providers = config.paymentProviders
  const selectedPlan = plansData?.data?.find((p) => p.id === selectedPlanId)
  const isPaidPlan = selectedPlan && (selectedPlan.price ?? 0) > 0
  const needsProviderSelection = isPaidPlan && providers.length > 1

  // Auto-select provider if only one available
  const effectiveProvider =
    providers.length === 1 ? providers[0].name : selectedProvider

  const handleSubmit = () => {
    if (selectedPlanId) {
      mutation.mutate(
        { plan_id: selectedPlanId, provider: effectiveProvider },
        {
          onSuccess: (data: UpgradeResponse) => {
            if (data.status === "activated") {
              queryClient.invalidateQueries({ queryKey: ["quota"] })
              showSuccessToast(t("plans.planActivated"))
              setIsOpen(false)
              setSelectedPlanId(null)
              setSelectedProvider(null)
            } else if (data.status === "pending_payment" && data.payment_url) {
              const providerInfo = providers.find(
                (p) => p.name === effectiveProvider,
              )
              setPendingPayment({
                url: data.payment_url,
                plan: data.plan,
                providerDisplayName:
                  providerInfo?.display_name ?? "Payment Provider",
              })
              showInfoToast(t("plans.completePaymentInfo"))
            } else if (
              data.status === "pending_authorization" &&
              data.authorization_url
            ) {
              window.location.href = data.authorization_url
            }
          },
        },
      )
    }
  }

  const handleOpenChange = (open: boolean) => {
    setIsOpen(open)
    if (!open) {
      setSelectedPlanId(null)
      setSelectedProvider(null)
      setPendingPayment(null)
    }
  }

  const handlePaymentRedirect = () => {
    if (pendingPayment?.url) {
      window.open(pendingPayment.url, "_blank")
    }
  }

  const isCurrentPlanSelected =
    selectedPlan?.name.toLowerCase() === currentPlan?.toLowerCase()

  const canSubmit =
    selectedPlanId &&
    !isCurrentPlanSelected &&
    (!isPaidPlan || effectiveProvider)

  return (
    <Dialog open={isOpen} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button size="sm">{t("plans.upgradePlan")}</Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>
            {pendingPayment
              ? t("plans.completePayment")
              : t("plans.changePlan")}
          </DialogTitle>
          <DialogDescription>
            {pendingPayment
              ? t("plans.completePaymentDesc", { plan: pendingPayment.plan })
              : t("plans.choosePlanDesc")}
          </DialogDescription>
        </DialogHeader>

        {pendingPayment ? (
          // Payment pending view
          <div className="py-8 text-center space-y-4">
            <div className="mx-auto w-16 h-16 bg-primary/10 rounded-full flex items-center justify-center">
              <ExternalLink className="h-8 w-8 text-primary" />
            </div>
            <p className="text-muted-foreground">
              {t("plans.paymentRedirectMsg", {
                provider: pendingPayment.providerDisplayName,
              })}
            </p>
            <Button onClick={handlePaymentRedirect} size="lg" className="gap-2">
              <ExternalLink className="h-4 w-4" />
              {t("plans.payWith", {
                provider: pendingPayment.providerDisplayName,
              })}
            </Button>
            <p className="text-xs text-muted-foreground">
              {t("plans.secureRedirectMsg", {
                provider: pendingPayment.providerDisplayName,
              })}
            </p>
          </div>
        ) : isLoading ? (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 py-4">
            {["a", "b", "c", "d"].map((id) => (
              <Skeleton key={id} className="h-48 w-full" />
            ))}
          </div>
        ) : (
          <div className="space-y-6">
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 py-4">
              {(plansData?.data as unknown as Plan[])?.map((plan) => (
                <PlanCard
                  key={plan.id}
                  plan={plan}
                  isCurrentPlan={
                    plan.name.toLowerCase() === currentPlan?.toLowerCase()
                  }
                  isSelected={selectedPlanId === plan.id}
                  onSelect={() => {
                    setSelectedPlanId(plan.id)
                    setSelectedProvider(null)
                  }}
                />
              ))}
            </div>

            {needsProviderSelection && (
              <div className="space-y-2">
                <p className="text-sm font-medium">
                  {t("plans.paymentMethod")}
                </p>
                <div className="flex gap-2">
                  {providers.map((provider) => (
                    <Button
                      key={provider.name}
                      variant={
                        selectedProvider === provider.name
                          ? "default"
                          : "outline"
                      }
                      size="sm"
                      onClick={() => setSelectedProvider(provider.name)}
                    >
                      {provider.display_name}
                    </Button>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}

        <DialogFooter>
          {pendingPayment ? (
            <Button variant="outline" onClick={() => setPendingPayment(null)}>
              {t("plans.chooseDifferent")}
            </Button>
          ) : (
            <>
              <DialogClose asChild>
                <Button variant="outline" disabled={mutation.isPending}>
                  {t("common.cancel")}
                </Button>
              </DialogClose>
              <LoadingButton
                onClick={handleSubmit}
                loading={mutation.isPending}
                disabled={!canSubmit}
              >
                {isCurrentPlanSelected
                  ? t("plans.currentPlan")
                  : t("plans.confirmUpgrade")}
              </LoadingButton>
            </>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

export default UpgradePlanDialog
