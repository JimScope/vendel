import type { ColumnDef } from "@tanstack/react-table"
import type { TFunction } from "i18next"

import { Badge } from "@/components/ui/badge"
import { cn, formatDate } from "@/lib/utils"
import type { ApiKey } from "@/types/collections"
import { ApiKeyActionsMenu } from "./ApiKeyActionsMenu"

function formatRelativeDate(dateString: string, t: TFunction): string {
  const date = new Date(dateString)
  const now = new Date()
  const diffMs = date.getTime() - now.getTime()
  const diffDays = Math.ceil(diffMs / (1000 * 60 * 60 * 24))

  if (diffDays < 0) return t("common.daysAgo", { count: Math.abs(diffDays) })
  if (diffDays === 0) return t("common.today")
  if (diffDays === 1) return t("common.inDays", { count: 1 })
  if (diffDays < 30) return t("common.inDays", { count: diffDays })
  return date.toLocaleDateString()
}

function isExpired(dateString: string): boolean {
  return new Date(dateString) < new Date()
}

function isExpiringSoon(dateString: string): boolean {
  const date = new Date(dateString)
  const now = new Date()
  const diffDays = (date.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)
  return diffDays > 0 && diffDays <= 7
}

export const getColumns = (t: TFunction): ColumnDef<ApiKey>[] => [
  {
    accessorKey: "name",
    header: t("common.name"),
    cell: ({ row }) => <span className="font-medium">{row.original.name}</span>,
  },
  {
    accessorKey: "is_active",
    header: t("common.status"),
    cell: ({ row }) => {
      const isActive = row.original.is_active
      return (
        <div className="flex items-center gap-2">
          <span
            className={cn(
              "size-2 rounded-full",
              isActive ? "bg-green-500" : "bg-gray-400",
            )}
          />
          <span className={isActive ? "" : "text-muted-foreground"}>
            {isActive ? t("apiKeys.statusActive") : t("apiKeys.statusRevoked")}
          </span>
        </div>
      )
    },
  },
  {
    accessorKey: "last_used_at",
    header: t("apiKeys.lastUsed"),
    cell: ({ row }) => (
      <span className="text-muted-foreground">
        {row.original.last_used_at
          ? formatDate(row.original.last_used_at)
          : t("apiKeys.never")}
      </span>
    ),
  },
  {
    accessorKey: "expires_at",
    header: t("apiKeys.expires"),
    cell: ({ row }) => {
      const expiresAt = row.original.expires_at
      if (!expiresAt) {
        return (
          <span className="text-muted-foreground">{t("apiKeys.never")}</span>
        )
      }
      if (isExpired(expiresAt)) {
        return (
          <Badge variant="destructive" className="text-[10px]">
            {t("apiKeys.expired")} {formatRelativeDate(expiresAt, t)}
          </Badge>
        )
      }
      if (isExpiringSoon(expiresAt)) {
        return (
          <Badge
            variant="outline"
            className="text-[10px] border-yellow-500 text-yellow-600"
          >
            {formatRelativeDate(expiresAt, t)}
          </Badge>
        )
      }
      return (
        <span className="text-muted-foreground">
          {formatRelativeDate(expiresAt, t)}
        </span>
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
        <ApiKeyActionsMenu apiKey={row.original} />
      </div>
    ),
  },
]
