import { createFileRoute, Link } from "@tanstack/react-router"
import {
  MessageSquare,
  MessageSquareText,
  Smartphone,
  Webhook,
} from "lucide-react"
import { useTranslation } from "react-i18next"

import OnboardingChecklist from "@/components/Dashboard/OnboardingChecklist"
import QuotaCard from "@/components/Plans/QuotaCard"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import useAppConfig from "@/hooks/useAppConfig"
import useAuth from "@/hooks/useAuth"
import { useDeviceList } from "@/hooks/useDeviceList"
import { useSMSList } from "@/hooks/useSMSList"
import { useWebhookList } from "@/hooks/useWebhookList"

export const Route = createFileRoute("/_layout/")({
  component: Dashboard,
})

function StatCard({
  title,
  description,
  value,
  icon: Icon,
  href,
  isLoading,
}: {
  title: string
  description: string
  value: number | undefined
  icon: React.ElementType
  href: string
  isLoading: boolean
}) {
  return (
    <Link
      to={href}
      className="rounded-lg focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
    >
      <Card className="cursor-pointer">
        <CardHeader>
          <CardTitle className="flex items-center gap-3">
            <Icon className="h-5 w-5 text-brand" />
            {title}
          </CardTitle>
          <CardDescription>{description}</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <Skeleton className="h-8 w-16" />
          ) : (
            <p className="text-3xl font-bold">{value ?? 0}</p>
          )}
        </CardContent>
      </Card>
    </Link>
  )
}

function Dashboard() {
  const { t } = useTranslation()
  const { user: currentUser } = useAuth()
  const { config } = useAppConfig()
  const { data: smsData, isLoading: smsLoading } = useSMSList("outgoing")
  const { data: incomingSmsData, isLoading: incomingSmsLoading } =
    useSMSList("incoming")
  const { data: devicesData, isLoading: devicesLoading } = useDeviceList()
  const { data: webhooksData, isLoading: webhooksLoading } = useWebhookList()

  return (
    <div className="flex flex-col gap-8">
      <title>{`${t("sidebar.dashboard")} - ${config.appName}`}</title>
      <div>
        <h1 className="text-2xl truncate max-w-md lg:max-w-lg">
          {t("dashboard.greeting", {
            name: currentUser?.full_name || currentUser?.email,
          })}
        </h1>
        <p className="text-muted-foreground">{t("dashboard.welcomeBack")}</p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title={t("dashboard.smsSent")}
          description={t("dashboard.totalSent")}
          value={smsData?.count}
          icon={MessageSquare}
          href="/sms"
          isLoading={smsLoading}
        />
        <StatCard
          title={t("dashboard.smsReceived")}
          description={t("dashboard.totalReceived")}
          value={incomingSmsData?.count}
          icon={MessageSquareText}
          href="/sms"
          isLoading={incomingSmsLoading}
        />
        <StatCard
          title={t("dashboard.devices")}
          description={t("dashboard.connectedDevices")}
          value={devicesData?.count}
          icon={Smartphone}
          href="/devices"
          isLoading={devicesLoading}
        />
        <StatCard
          title={t("dashboard.webhooks")}
          description={t("dashboard.activeWebhooks")}
          value={webhooksData?.count}
          icon={Webhook}
          href="/webhooks"
          isLoading={webhooksLoading}
        />
      </div>

      <div className="grid gap-4 lg:grid-cols-3">
        <QuotaCard />
        <div className="lg:col-span-2">
          <OnboardingChecklist />
        </div>
      </div>
    </div>
  )
}
