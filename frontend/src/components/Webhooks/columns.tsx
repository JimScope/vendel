import type { ColumnDef } from "@tanstack/react-table"
import type { TFunction } from "i18next"

import { Badge } from "@/components/ui/badge"
import { cn, formatDate } from "@/lib/utils"
import type { WebhookEvent } from "@/lib/webhook-events"
import type { WebhookConfig } from "@/types/collections"
import { WebhookActionsMenu } from "./WebhookActionsMenu"

const WEBHOOK_EVENT_KEYS = {
  sms_received: "webhookEvents.sms_received",
  sms_sent: "webhookEvents.sms_sent",
  sms_delivered: "webhookEvents.sms_delivered",
  sms_failed: "webhookEvents.sms_failed",
} as const

export function getColumns(t: TFunction): ColumnDef<WebhookConfig>[] {
  return [
    {
      accessorKey: "url",
      header: "URL",
      cell: ({ row }) => (
        <span className="font-medium font-mono text-sm truncate max-w-xs block">
          {row.original.url}
        </span>
      ),
    },
    {
      accessorKey: "events",
      header: t("webhooks.events"),
      cell: ({ row }) => {
        const events = row.original.events
        if (!events || events.length === 0) {
          return <Badge variant="secondary">{t("webhooks.none")}</Badge>
        }
        return (
          <div className="flex gap-1 flex-wrap">
            {events.map((event) => (
              <Badge key={event} variant="secondary">
                {t(
                  WEBHOOK_EVENT_KEYS[event as WebhookEvent] ??
                    ("webhookEvents.sms_received" as const),
                )}
              </Badge>
            ))}
          </div>
        )
      },
    },
    {
      accessorKey: "active",
      header: t("common.status"),
      cell: ({ row }) => {
        const isActive = row.original.active
        return (
          <div className="flex items-center gap-2">
            <span
              className={cn(
                "size-2 rounded-full",
                isActive ? "bg-green-500" : "bg-gray-400",
              )}
            />
            <span className={isActive ? "" : "text-muted-foreground"}>
              {isActive
                ? t("webhooks.statusActive")
                : t("webhooks.statusInactive")}
            </span>
          </div>
        )
      },
    },
    {
      accessorKey: "created",
      header: t("common.created"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {formatDate(row.original.created)}
        </span>
      ),
    },
    {
      id: "actions",
      header: () => <span className="sr-only">{t("common.actions")}</span>,
      cell: ({ row }) => (
        <div className="flex justify-end">
          <WebhookActionsMenu webhook={row.original} />
        </div>
      ),
    },
  ]
}
