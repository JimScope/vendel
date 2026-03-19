import type { ColumnDef } from "@tanstack/react-table"
import { useMemo } from "react"
import { useTranslation } from "react-i18next"
import { DataTable } from "@/components/Common/DataTable"
import { Badge } from "@/components/ui/badge"
import { formatDate } from "@/lib/utils"
import type { SMSMessage } from "@/types/collections"
import { SMSActionsMenu } from "./SMSActionsMenu"

interface SMSTableProps {
  data: SMSMessage[]
}

function getStatusBadgeVariant(
  status: string,
): "default" | "secondary" | "destructive" | "outline" {
  switch (status) {
    case "delivered":
      return "default"
    case "sent":
      return "secondary"
    case "failed":
      return "destructive"
    default:
      return "outline"
  }
}

function truncateText(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text
  return `${text.slice(0, maxLength)}...`
}

export function SMSTable({ data }: SMSTableProps) {
  const { t } = useTranslation()

  const columns: ColumnDef<SMSMessage>[] = useMemo(
    () => [
      {
        accessorKey: "to",
        header: t("sms.to"),
        cell: ({ row }) => {
          const msg = row.original
          const display =
            msg.message_type === "incoming" ? msg.from_number || msg.to : msg.to
          return <span className="font-medium">{display}</span>
        },
      },
      {
        accessorKey: "body",
        header: t("sms.body"),
        cell: ({ row }) => (
          <span className="max-w-xs truncate" title={row.original.body}>
            {truncateText(row.original.body, 50)}
          </span>
        ),
      },
      {
        accessorKey: "status",
        header: t("common.status"),
        cell: ({ row }) => (
          <Badge
            variant={getStatusBadgeVariant(row.original.status || "pending")}
          >
            {row.original.status || "pending"}
          </Badge>
        ),
      },
      {
        accessorKey: "message_type",
        header: t("sms.type"),
        cell: ({ row }) => (
          <Badge variant="outline">
            {row.original.message_type || "outgoing"}
          </Badge>
        ),
      },
      {
        accessorKey: "created",
        header: t("sms.date"),
        cell: ({ row }) => (
          <span className="text-muted-foreground text-sm">
            {formatDate(row.original.created)}
          </span>
        ),
      },
      {
        id: "actions",
        header: () => null,
        cell: ({ row }) => <SMSActionsMenu sms={row.original} />,
      },
    ],
    [t],
  )

  // Sort by created_at descending
  const sortedData = [...data].sort(
    (a, b) => new Date(b.created).getTime() - new Date(a.created).getTime(),
  )

  return (
    <DataTable
      columns={columns}
      data={sortedData}
      caption={t("sms.smsMessages")}
    />
  )
}
