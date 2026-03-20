import { Link } from "@tanstack/react-router"
import { BookOpen, Check, ChevronRight, Code } from "lucide-react"
import { useTranslation } from "react-i18next"

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { useApiKeyList } from "@/hooks/useApiKeyList"
import { useDeviceList } from "@/hooks/useDeviceList"
import { useSMSList } from "@/hooks/useSMSList"
import { cn } from "@/lib/utils"

interface Step {
  title: string
  description: string
  completed: boolean
  href: string
}

export default function OnboardingChecklist() {
  const { t } = useTranslation()
  const { data: devicesData } = useDeviceList()
  const { data: apiKeysData } = useApiKeyList()
  const { data: smsData } = useSMSList("outgoing")

  const steps: Step[] = [
    {
      title: t("onboarding.registerDevice"),
      description: t("onboarding.registerDeviceDesc"),
      completed: (devicesData?.count ?? 0) > 0,
      href: "/devices",
    },
    {
      title: t("onboarding.createApiKey"),
      description: t("onboarding.createApiKeyDesc"),
      completed: (apiKeysData?.count ?? 0) > 0,
      href: "/integrations",
    },
    {
      title: t("onboarding.sendFirstSms"),
      description: t("onboarding.sendFirstSmsDesc"),
      completed: (smsData?.count ?? 0) > 0,
      href: "/sms",
    },
  ]

  const completedCount = steps.filter((s) => s.completed).length
  const allDone = completedCount === steps.length

  if (allDone) {
    const docLinks = [
      {
        title: t("onboarding.viewDocs"),
        description: t("onboarding.viewDocsDesc"),
        href: "https://vendel.cc/docs/",
        icon: BookOpen,
      },
      {
        title: t("onboarding.viewApiRef"),
        description: t("onboarding.viewApiRefDesc"),
        href: "https://vendel.cc/docs/api/send-sms/",
        icon: Code,
      },
    ]

    return (
      <Card>
        <CardHeader>
          <CardTitle>{t("onboarding.allDoneTitle")}</CardTitle>
          <CardDescription>{t("onboarding.allDoneSubtitle")}</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-2">
          {docLinks.map((link) => (
            <a
              key={link.href}
              href={link.href}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-3 rounded-lg p-3 transition-colors hover:bg-muted/50"
            >
              <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-brand/20">
                <link.icon className="h-4 w-4 text-brand" />
              </div>
              <div className="min-w-0 flex-1">
                <p className="text-sm font-medium">{link.title}</p>
                <p className="text-xs text-muted-foreground">
                  {link.description}
                </p>
              </div>
              <ChevronRight className="h-4 w-4 shrink-0 text-muted-foreground" />
            </a>
          ))}
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle>{t("onboarding.title")}</CardTitle>
            <CardDescription>{t("onboarding.subtitle")}</CardDescription>
          </div>
          <span className="text-sm font-medium text-muted-foreground">
            {completedCount}/{steps.length}
          </span>
        </div>
        <div className="mt-3 h-2 rounded-full bg-muted">
          <div
            className="h-full rounded-full bg-brand transition-all"
            style={{ width: `${(completedCount / steps.length) * 100}%` }}
          />
        </div>
      </CardHeader>
      <CardContent className="grid gap-2">
        {steps.map((step, index) => (
          <Link
            key={step.href}
            to={step.href}
            className={cn(
              "flex items-center gap-3 rounded-lg p-3 transition-colors hover:bg-muted/50",
              step.completed && "opacity-60",
            )}
          >
            {step.completed ? (
              <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-brand/20">
                <Check className="h-4 w-4 text-brand" />
              </div>
            ) : (
              <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full border-2 border-muted-foreground/30">
                <span className="text-xs font-semibold text-muted-foreground">
                  {index + 1}
                </span>
              </div>
            )}
            <div className="min-w-0 flex-1">
              <p
                className={cn(
                  "text-sm font-medium",
                  step.completed && "line-through",
                )}
              >
                {step.title}
              </p>
              <p className="text-xs text-muted-foreground">
                {step.description}
              </p>
            </div>
            <ChevronRight className="h-4 w-4 shrink-0 text-muted-foreground" />
          </Link>
        ))}
      </CardContent>
    </Card>
  )
}
