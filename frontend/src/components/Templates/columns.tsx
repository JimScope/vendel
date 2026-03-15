import type { ColumnDef } from "@tanstack/react-table"

import { formatDate } from "@/lib/utils"
import type { SMSTemplate } from "@/types/collections"
import { TemplateActionsMenu } from "./TemplateActionsMenu"

export const columns: ColumnDef<SMSTemplate>[] = [
  {
    accessorKey: "name",
    header: "Name",
    cell: ({ row }) => <span className="font-medium">{row.original.name}</span>,
  },
  {
    accessorKey: "body",
    header: "Body",
    cell: ({ row }) => (
      <span className="text-muted-foreground text-sm truncate max-w-xs block">
        {row.original.body}
      </span>
    ),
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
        <TemplateActionsMenu template={row.original} />
      </div>
    ),
  },
]
