import type { ColumnDef } from "@tanstack/react-table"

import { Badge } from "@/components/ui/badge"
import { cn, formatDate } from "@/lib/utils"
import type { WebhookConfig } from "@/types/collections"
import { WebhookActionsMenu } from "./WebhookActionsMenu"

export const columns: ColumnDef<WebhookConfig>[] = [
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
    header: "Events",
    cell: ({ row }) => {
      const events = row.original.events || "all"
      return <Badge variant="secondary">{events}</Badge>
    },
  },
  {
    accessorKey: "active",
    header: "Status",
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
            {isActive ? "Active" : "Inactive"}
          </span>
        </div>
      )
    },
  },
  {
    accessorKey: "created",
    header: "Created",
    cell: ({ row }) => (
      <span className="text-muted-foreground">
        {formatDate(row.original.created)}
      </span>
    ),
  },
  {
    id: "actions",
    header: () => <span className="sr-only">Actions</span>,
    cell: ({ row }) => (
      <div className="flex justify-end">
        <WebhookActionsMenu webhook={row.original} />
      </div>
    ),
  },
]
