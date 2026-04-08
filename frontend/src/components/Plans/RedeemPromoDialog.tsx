import { useQueryClient } from "@tanstack/react-query"
import { Gift } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

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
import useCustomToast from "@/hooks/useCustomToast"
import { useRedeemPromo } from "@/hooks/useRedeemPromo"

interface RedeemPromoDialogProps {
  trigger: React.ReactNode
}

function RedeemPromoDialog({ trigger }: RedeemPromoDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const mutation = useRedeemPromo()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  const [isOpen, setIsOpen] = useState(false)
  const [code, setCode] = useState("")

  const resetState = () => {
    setCode("")
    mutation.reset()
  }

  const handleOpenChange = (open: boolean) => {
    setIsOpen(open)
    if (!open) resetState()
  }

  const handleSubmit = () => {
    if (!code.trim()) return

    mutation.mutate(code.trim(), {
      onSuccess: (data) => {
        showSuccessToast(
          t("plans.promoApplied", { amount: data.amount.toFixed(2) }),
        )
        queryClient.invalidateQueries({ queryKey: ["balance"] })
        handleOpenChange(false)
      },
      onError: (error) => {
        const message =
          (error as { response?: { message?: string } }).response?.message ??
          t("plans.promoInvalid")
        showErrorToast(message)
      },
    })
  }

  return (
    <Dialog open={isOpen} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Gift className="h-5 w-5 text-brand" />
            {t("plans.redeemPromo")}
          </DialogTitle>
          <DialogDescription>
            {t("plans.redeemPromoDescription")}
          </DialogDescription>
        </DialogHeader>

        <div className="py-2">
          <Input
            placeholder={t("plans.promoPlaceholder")}
            value={code}
            onChange={(e) => setCode(e.target.value.toUpperCase())}
            onKeyDown={(e) => {
              if (e.key === "Enter" && code.trim()) handleSubmit()
            }}
            autoFocus
          />
        </div>

        <DialogFooter>
          <LoadingButton
            onClick={handleSubmit}
            loading={mutation.isPending}
            disabled={!code.trim()}
          >
            {t("plans.redeem")}
          </LoadingButton>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

export default RedeemPromoDialog
