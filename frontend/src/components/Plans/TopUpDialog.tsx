import { useQueryClient } from "@tanstack/react-query"
import { Cuer } from "cuer"
import { ArrowLeft, Check, Copy, ExternalLink } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { LoadingButton } from "@/components/ui/loading-button"
import useAppConfig from "@/hooks/useAppConfig"
import { useCopyToClipboard } from "@/hooks/useCopyToClipboard"
import { type TopUpResponse, useTopUp } from "@/hooks/useTopUp"

type Step = "provider" | "amount" | "result"

const PRESET_AMOUNTS = [5, 10, 25, 50]

interface TopUpDialogProps {
  trigger: React.ReactNode
}

function TopUpDialog({ trigger }: TopUpDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { config } = useAppConfig()
  const mutation = useTopUp()
  const [copiedText, copyToClipboard] = useCopyToClipboard()

  const [isOpen, setIsOpen] = useState(false)
  const [step, setStep] = useState<Step>("provider")
  const [selectedProvider, setSelectedProvider] = useState<string | null>(null)
  const [amount, setAmount] = useState<number | null>(null)
  const [customAmount, setCustomAmount] = useState("")
  const [topUpResult, setTopUpResult] = useState<TopUpResponse | null>(null)

  const providers = config.paymentProviders

  const resetState = () => {
    setStep("provider")
    setSelectedProvider(null)
    setAmount(null)
    setCustomAmount("")
    setTopUpResult(null)
    mutation.reset()
  }

  const handleOpenChange = (open: boolean) => {
    setIsOpen(open)
    if (!open) {
      resetState()
    }
  }

  const isTronDealer = (provider: string) => provider === "trondealer"

  const handleProviderSelect = (providerName: string) => {
    setSelectedProvider(providerName)
    if (isTronDealer(providerName)) {
      // TronDealer: skip amount step, submit immediately
      submitTopUp(providerName, 0)
    } else {
      setStep("amount")
    }
  }

  const handleAmountSubmit = () => {
    const finalAmount = amount ?? Number.parseFloat(customAmount)
    if (selectedProvider && finalAmount > 0) {
      submitTopUp(selectedProvider, finalAmount)
    }
  }

  const submitTopUp = (provider: string, topUpAmount: number) => {
    mutation.mutate(
      { amount: topUpAmount, provider },
      {
        onSuccess: (data: TopUpResponse) => {
          setTopUpResult(data)
          setStep("result")
          // Invalidate balance so it refreshes when the user closes the dialog
          queryClient.invalidateQueries({ queryKey: ["balance"] })
        },
      },
    )
  }

  const handleBack = () => {
    if (step === "amount") {
      setStep("provider")
      setSelectedProvider(null)
      setAmount(null)
      setCustomAmount("")
    } else if (step === "result") {
      if (selectedProvider && isTronDealer(selectedProvider)) {
        setStep("provider")
        setSelectedProvider(null)
      } else {
        setStep("amount")
        setAmount(null)
        setCustomAmount("")
      }
      setTopUpResult(null)
      mutation.reset()
    }
  }

  const dialogTitle = () => {
    switch (step) {
      case "provider":
        return t("plans.addFunds")
      case "amount":
        return t("plans.selectAmount")
      case "result":
        return t("plans.addFunds")
    }
  }

  const dialogDescription = () => {
    switch (step) {
      case "provider":
        return t("plans.selectProvider")
      case "amount":
        return t("plans.selectAmount")
      case "result":
        return topUpResult?.wallet_address
          ? t("plans.sendStablecoins")
          : undefined
    }
  }

  // Auto-select if only one provider
  const shouldAutoSelect = providers.length === 1

  const handleDialogOpen = () => {
    if (shouldAutoSelect && providers.length === 1) {
      const only = providers[0]
      if (isTronDealer(only.name)) {
        setSelectedProvider(only.name)
        submitTopUp(only.name, 0)
      } else {
        setSelectedProvider(only.name)
        setStep("amount")
      }
    }
  }

  const canSubmitAmount =
    (amount !== null && amount > 0) ||
    (customAmount !== "" && Number.parseFloat(customAmount) > 0)

  const showBackButton =
    (step === "amount" && providers.length > 1) ||
    (step === "result" && !shouldAutoSelect)

  return (
    <Dialog
      open={isOpen}
      onOpenChange={(open) => {
        handleOpenChange(open)
        if (open) handleDialogOpen()
      }}
    >
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent
        className="sm:max-w-sm"
        onInteractOutside={(e) => e.preventDefault()}
      >
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {showBackButton && (
              <Button
                variant="ghost"
                size="icon"
                className="size-7 shrink-0"
                onClick={handleBack}
              >
                <ArrowLeft className="size-4" />
              </Button>
            )}
            {dialogTitle()}
          </DialogTitle>
          {(() => {
            const desc = dialogDescription()
            return desc ? <DialogDescription>{desc}</DialogDescription> : null
          })()}
        </DialogHeader>

        {/* Step 1: Provider selection */}
        {step === "provider" && !shouldAutoSelect && (
          <div className="flex flex-col gap-2 py-2">
            {providers.map((provider) => (
              <Button
                key={provider.name}
                variant="outline"
                className="justify-start h-12"
                onClick={() => handleProviderSelect(provider.name)}
                disabled={mutation.isPending}
              >
                {provider.display_name}
              </Button>
            ))}
          </div>
        )}

        {/* Loading state when auto-selecting provider */}
        {step === "provider" && shouldAutoSelect && mutation.isPending && (
          <div className="flex items-center justify-center py-8">
            <p className="text-sm text-muted-foreground">
              {t("common.loading")}
            </p>
          </div>
        )}

        {/* Step 2: Amount selection (QvaPay/Stripe only) */}
        {step === "amount" && (
          <div className="space-y-4 py-2">
            <div className="grid grid-cols-2 gap-2">
              {PRESET_AMOUNTS.map((preset) => (
                <Button
                  key={preset}
                  variant={amount === preset ? "default" : "outline"}
                  className="h-12 text-lg"
                  onClick={() => {
                    setAmount(preset)
                    setCustomAmount("")
                  }}
                >
                  ${preset}
                </Button>
              ))}
            </div>

            <div className="flex items-center gap-2">
              <span className="text-sm text-muted-foreground shrink-0">
                {t("plans.customAmount")}:
              </span>
              <div className="relative flex-1">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground">
                  $
                </span>
                <Input
                  type="number"
                  min="1"
                  step="0.01"
                  placeholder="0.00"
                  className="pl-7"
                  value={customAmount}
                  onChange={(e) => {
                    setCustomAmount(e.target.value)
                    setAmount(null)
                  }}
                />
              </div>
            </div>
          </div>
        )}

        {/* Step 3: Result */}
        {step === "result" && topUpResult && (
          <div className="flex flex-col items-center gap-3 py-2">
            {topUpResult.wallet_address ? (
              <>
                <div
                  className="rounded-lg border bg-white p-3"
                  role="img"
                  aria-label="QR code containing wallet address"
                >
                  <Cuer.Root value={topUpResult.wallet_address} size={160}>
                    <Cuer.Finder fill="black" />
                    <Cuer.Cells fill="black" />
                  </Cuer.Root>
                </div>

                <div className="flex w-full items-center gap-2 rounded-lg border bg-muted/50 p-2">
                  <code className="flex-1 break-all font-mono text-xs">
                    {topUpResult.wallet_address}
                  </code>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="shrink-0"
                    onClick={() => copyToClipboard(topUpResult.wallet_address!)}
                    aria-label={
                      copiedText
                        ? t("common.copied")
                        : t("plans.copyAddress")
                    }
                  >
                    {copiedText ? (
                      <Check className="h-4 w-4 text-brand" />
                    ) : (
                      <Copy className="h-4 w-4" />
                    )}
                  </Button>
                </div>

                <p className="text-xs text-muted-foreground text-center">
                  {t("plans.sendStablecoins")} &middot;{" "}
                  {t("plans.confirmationDelay")}
                </p>
              </>
            ) : topUpResult.payment_url ? (
              <div className="w-full text-center space-y-4 py-4">
                <div className="mx-auto w-14 h-14 bg-primary/10 rounded-full flex items-center justify-center">
                  <ExternalLink className="h-7 w-7 text-primary" />
                </div>
                <Button
                  onClick={() =>
                    window.open(topUpResult.payment_url, "_blank")
                  }
                  size="lg"
                  className="gap-2"
                >
                  <ExternalLink className="h-4 w-4" />
                  {t("plans.payWith", {
                    provider:
                      providers.find((p) => p.name === selectedProvider)
                        ?.display_name ?? selectedProvider,
                  })}
                </Button>
                <p className="text-xs text-muted-foreground">
                  {t("plans.secureRedirectMsg", {
                    provider:
                      providers.find((p) => p.name === selectedProvider)
                        ?.display_name ?? selectedProvider,
                  })}
                </p>
              </div>
            ) : null}
          </div>
        )}

        {/* Footer */}
        {step === "amount" && (
          <DialogFooter>
            <LoadingButton
              onClick={handleAmountSubmit}
              loading={mutation.isPending}
              disabled={!canSubmitAmount}
            >
              {t("common.continue")}
            </LoadingButton>
          </DialogFooter>
        )}

        {step === "result" && (
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => handleOpenChange(false)}
            >
              {t("common.done")}
            </Button>
          </DialogFooter>
        )}
      </DialogContent>
    </Dialog>
  )
}

export default TopUpDialog
