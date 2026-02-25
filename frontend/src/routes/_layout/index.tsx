import { createFileRoute, Link } from "@tanstack/react-router"
import {
  CheckCircle2,
  ChevronRight,
  MessageSquare,
  MessageSquareText,
  Smartphone,
  Webhook,
} from "lucide-react"

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

function GettingStartedCard() {
  return (
    <Link
      to="/devices"
      className="rounded-lg focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
    >
      <Card className="cursor-pointer group">
        <CardHeader className="flex-row items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="rounded-full bg-brand/10 p-2">
              <CheckCircle2 className="h-5 w-5 text-brand" />
            </div>
            <div>
              <CardTitle>Getting Started</CardTitle>
              <CardDescription>
                Set up your first device and start sending SMS
              </CardDescription>
            </div>
          </div>
          <ChevronRight className="h-5 w-5 text-muted-foreground group-hover:translate-x-1 transition-transform" />
        </CardHeader>
      </Card>
    </Link>
  )
}

function Dashboard() {
  const { user: currentUser } = useAuth()
  const { config } = useAppConfig()
  const { data: smsData, isLoading: smsLoading } = useSMSList("outgoing")
  const { data: incomingSmsData, isLoading: incomingSmsLoading } =
    useSMSList("incoming")
  const { data: devicesData, isLoading: devicesLoading } = useDeviceList()
  const { data: webhooksData, isLoading: webhooksLoading } = useWebhookList()

  return (
    <div className="flex flex-col gap-8">
      <title>{`Dashboard - ${config.appName}`}</title>
      <div>
        <h1 className="text-2xl truncate max-w-md lg:max-w-lg">
          Hi, {currentUser?.full_name || currentUser?.email}
        </h1>
        <p className="text-muted-foreground">
          Welcome back, nice to see you again!
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="SMS Sent"
          description="Total messages sent"
          value={smsData?.count}
          icon={MessageSquare}
          href="/sms"
          isLoading={smsLoading}
        />
        <StatCard
          title="SMS Received"
          description="Total messages received"
          value={incomingSmsData?.count}
          icon={MessageSquareText}
          href="/sms"
          isLoading={incomingSmsLoading}
        />
        <StatCard
          title="Devices"
          description="Connected devices"
          value={devicesData?.count}
          icon={Smartphone}
          href="/devices"
          isLoading={devicesLoading}
        />
        <StatCard
          title="Webhooks"
          description="Active webhooks"
          value={webhooksData?.count}
          icon={Webhook}
          href="/webhooks"
          isLoading={webhooksLoading}
        />
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <QuotaCard />
        <GettingStartedCard />
      </div>
    </div>
  )
}
