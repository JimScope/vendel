import { Wallet } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import useAppConfig from "@/hooks/useAppConfig"
import { useBalance } from "@/hooks/useBalance"
import TopUpDialog from "./TopUpDialog"

function BalanceCard() {
  const { t } = useTranslation()
  const { data: balance } = useBalance()
  const { config } = useAppConfig()

  const hasProviders = config.paymentProviders.length > 0
  const hasBalance = (balance?.balance ?? 0) > 0

  if (!hasProviders && !hasBalance) return null

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Wallet className="h-5 w-5 text-brand" />
            {t("plans.balance")}
          </div>
          <span className="text-2xl">
            ${(balance?.balance ?? 0).toFixed(2)}{" "}
            <span className="text-sm font-normal text-muted-foreground">
              {balance?.currency ?? "USDT"}
            </span>
          </span>
        </CardTitle>
      </CardHeader>
      {hasProviders && (
        <CardContent>
          <TopUpDialog
            trigger={
              <button
                type="button"
                className="w-full rounded-lg border border-dashed border-brand/40 py-2.5 text-sm font-medium text-brand transition-colors hover:bg-brand/5"
              >
                {t("plans.addFunds")}
              </button>
            }
          />
        </CardContent>
      )}
    </Card>
  )
}

export default BalanceCard
