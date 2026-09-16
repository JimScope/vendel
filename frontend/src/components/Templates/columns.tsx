import type { TFunction } from "i18next"
import type { AppColumnDef } from "@/lib/table"

import { formatDate } from "@/lib/utils"
import type { SMSTemplate } from "@/types/collections"
import { TemplateActionsMenu } from "./TemplateActionsMenu"

export const getColumns = (t: TFunction): AppColumnDef<SMSTemplate>[] => [
  {
    accessorKey: "name",
    header: t("common.name"),
    cell: ({ row }) => <span className="font-medium">{row.original.name}</span>,
  },
  {
    accessorKey: "body",
    header: t("templates.body"),
    cell: ({ row }) => (
      <span className="text-muted-foreground text-sm truncate max-w-xs block">
        {row.original.body}
      </span>
    ),
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
        <TemplateActionsMenu template={row.original} />
      </div>
    ),
  },
]
