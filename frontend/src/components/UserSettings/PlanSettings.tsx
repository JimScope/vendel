import { useTranslation } from "react-i18next"
import QuotaCard from "@/components/Plans/QuotaCard"

const PlanSettings = () => {
  const { t } = useTranslation()
  return (
    <div className="max-w-md">
      <h3 className="text-lg font-semibold py-4">{t("settings.planUsage")}</h3>
      <QuotaCard />
    </div>
  )
}

export default PlanSettings
