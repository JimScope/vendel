import { createFileRoute, Link as RouterLink } from "@tanstack/react-router"
import { XCircle } from "lucide-react"
import { useTranslation } from "react-i18next"
import { AuthLayout } from "@/components/Common/AuthLayout"
import { Button } from "@/components/ui/button"

export const Route = createFileRoute("/subscription/error")({
  component: SubscriptionError,
  head: () => ({
    meta: [
      {
        title: "Subscription Error",
      },
    ],
  }),
})

function SubscriptionError() {
  const { t } = useTranslation()

  return (
    <AuthLayout>
      <div className="flex flex-col items-center gap-6 text-center">
        <div className="rounded-full bg-destructive/10 p-4">
          <XCircle className="h-12 w-12 text-destructive" />
        </div>

        <h1 className="text-2xl">{t("subscription.failed")}</h1>

        <p className="text-muted-foreground">{t("subscription.failedDesc")}</p>

        <div className="text-sm text-muted-foreground">
          <p>{t("subscription.noCharges")}</p>
          <p>{t("subscription.tryAgainOrContact")}</p>
        </div>

        <div className="flex gap-4 mt-4">
          <Button asChild>
            <RouterLink to="/settings">{t("subscription.tryAgain")}</RouterLink>
          </Button>
          <Button variant="outline" asChild>
            <RouterLink to="/">{t("subscription.goToDashboard")}</RouterLink>
          </Button>
        </div>
      </div>
    </AuthLayout>
  )
}

export default SubscriptionError
