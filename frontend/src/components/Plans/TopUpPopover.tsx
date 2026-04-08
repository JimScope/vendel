import { Gift, Wallet } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import useAppConfig from "@/hooks/useAppConfig"
import { useBalance } from "@/hooks/useBalance"
import RedeemPromoDialog from "./RedeemPromoDialog"
import TopUpDialog from "./TopUpDialog"

function TopUpPopover() {
  const { t } = useTranslation()
  const { data: balance } = useBalance()
  const { config } = useAppConfig()

  const hasProviders = config.paymentProviders.length > 0
  const hasBalance = (balance?.balance ?? 0) > 0

  if (!hasProviders && !hasBalance) return null

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="ghost" size="sm" className="gap-1.5">
          <Wallet className="size-4 text-brand" />
          <span className="text-sm font-medium">
            ${(balance?.balance ?? 0).toFixed(2)}
          </span>
        </Button>
      </PopoverTrigger>
      <PopoverContent align="end" side="bottom" className="w-56 p-3">
        <div className="flex flex-col items-center gap-3">
          <div className="flex items-center gap-2">
            <Wallet className="size-4 text-brand" />
            <span className="text-lg font-semibold">
              ${(balance?.balance ?? 0).toFixed(2)}{" "}
              <span className="text-xs font-normal text-muted-foreground">
                {balance?.currency ?? "USDT"}
              </span>
            </span>
          </div>

          {hasProviders && (
            <TopUpDialog
              trigger={
                <Button variant="outline" size="sm" className="w-full">
                  {t("plans.addFunds")}
                </Button>
              }
            />
          )}
          <RedeemPromoDialog
            trigger={
              <Button variant="ghost" size="sm" className="w-full gap-1.5">
                <Gift className="size-3.5" />
                {t("plans.redeemPromo")}
              </Button>
            }
          />
        </div>
      </PopoverContent>
    </Popover>
  )
}

export default TopUpPopover
