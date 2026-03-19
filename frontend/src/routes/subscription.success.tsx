import { createFileRoute, Link as RouterLink } from "@tanstack/react-router"
import { CheckCircle } from "lucide-react"
import { useTranslation } from "react-i18next"
import { AuthLayout } from "@/components/Common/AuthLayout"
import { Button } from "@/components/ui/button"

export const Route = createFileRoute("/subscription/success")({
  component: SubscriptionSuccess,
  head: () => ({
    meta: [
      {
        title: "Subscription Activated",
      },
    ],
  }),
})

function SubscriptionSuccess() {
  const { t } = useTranslation()

  return (
    <AuthLayout>
      <div className="flex flex-col items-center gap-6 text-center">
        <div className="rounded-full bg-green-500/10 p-4">
          <CheckCircle className="h-12 w-12 text-green-500" />
        </div>

        <h1 className="text-2xl">{t("subscription.activated")}</h1>

        <p className="text-muted-foreground">
          {t("subscription.activatedDesc")}
        </p>

        <div className="text-sm text-muted-foreground">
          <p>{t("subscription.firstPayment")}</p>
          <p>{t("subscription.manageInSettings")}</p>
        </div>

        <div className="flex gap-4 mt-4">
          <Button asChild>
            <RouterLink to="/">{t("subscription.goToDashboard")}</RouterLink>
          </Button>
          <Button variant="outline" asChild>
            <RouterLink to="/settings">
              {t("subscription.viewSettings")}
            </RouterLink>
          </Button>
        </div>
      </div>
    </AuthLayout>
  )
}

export default SubscriptionSuccess
