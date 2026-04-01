import { useTranslation } from "react-i18next"
import BalanceCard from "@/components/Plans/BalanceCard"
import QuotaCard from "@/components/Plans/QuotaCard"

const PlanSettings = () => {
  const { t } = useTranslation()
  return (
    <div className="max-w-md space-y-4">
      <h3 className="text-lg font-semibold py-4">{t("settings.planUsage")}</h3>
      <QuotaCard />
      <BalanceCard />
    </div>
  )
}

export default PlanSettings
