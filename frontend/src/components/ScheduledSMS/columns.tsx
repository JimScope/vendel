import type { ColumnDef } from "@tanstack/react-table"

import { Badge } from "@/components/ui/badge"
import { cn, formatDate } from "@/lib/utils"
import { ScheduledSMSActionsMenu } from "./ScheduledSMSActionsMenu"

export const columns: ColumnDef<Record<string, any>>[] = [
  {
    accessorKey: "name",
    header: "Name",
    cell: ({ row }) => <span className="font-medium">{row.original.name}</span>,
  },
  {
    accessorKey: "schedule_type",
    header: "Type",
    cell: ({ row }) => {
      const type = row.original.schedule_type
      return (
        <Badge variant="secondary">
          {type === "one_time" ? "One-time" : "Recurring"}
        </Badge>
      )
    },
  },
  {
    accessorKey: "recipients",
    header: "Recipients",
    cell: ({ row }) => {
      const recipients = row.original.recipients
      const count = Array.isArray(recipients) ? recipients.length : 0
      return (
        <span className="text-muted-foreground">
          {count} {count === 1 ? "recipient" : "recipients"}
        </span>
      )
    },
  },
  {
    accessorKey: "next_run_at",
    header: "Next Run",
    cell: ({ row }) => (
      <span className="text-muted-foreground">
        {formatDate(row.original.next_run_at)}
      </span>
    ),
  },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => {
      const status = row.original.status
      return (
        <div className="flex items-center gap-2">
          <span
            className={cn(
              "size-2 rounded-full",
              status === "active"
                ? "bg-green-500"
                : status === "paused"
                  ? "bg-yellow-500"
                  : "bg-gray-400",
            )}
          />
          <span
            className={
              status === "active" ? "" : "text-muted-foreground capitalize"
            }
          >
            {status === "active"
              ? "Active"
              : status === "paused"
                ? "Paused"
                : "Completed"}
          </span>
        </div>
      )
    },
  },
  {
    id: "actions",
    header: () => <span className="sr-only">Actions</span>,
    cell: ({ row }) => (
      <div className="flex justify-end">
        <ScheduledSMSActionsMenu schedule={row.original} />
      </div>
    ),
  },
]
