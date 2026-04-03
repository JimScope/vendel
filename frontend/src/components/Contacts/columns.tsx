import type { ColumnDef } from "@tanstack/react-table"
import type { TFunction } from "i18next"

import { ContactActionsMenu } from "@/components/Contacts/ContactActionsMenu"
import { Badge } from "@/components/ui/badge"
import { formatDate } from "@/lib/utils"
import type { Contact, ContactGroup } from "@/types/collections"

export function getColumns(
  t: TFunction,
  groups: ContactGroup[],
): ColumnDef<Contact>[] {
  const groupMap = new Map(groups.map((g) => [g.id, g.name]))

  return [
    {
      accessorKey: "name",
      header: t("contacts.name"),
      cell: ({ row }) => (
        <span className="font-medium">{row.original.name}</span>
      ),
    },
    {
      accessorKey: "phone_number",
      header: t("contacts.phoneNumber"),
      cell: ({ row }) => (
        <span className="font-mono text-muted-foreground">
          {row.original.phone_number}
        </span>
      ),
    },
    {
      accessorKey: "groups",
      header: t("contacts.groups"),
      cell: ({ row }) => {
        const contactGroups = row.original.groups || []
        if (contactGroups.length === 0) return null
        return (
          <div className="flex flex-wrap gap-1">
            {contactGroups.map((groupId) => (
              <Badge key={groupId} variant="secondary" className="text-xs">
                {groupMap.get(groupId) || groupId}
              </Badge>
            ))}
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
          <ContactActionsMenu contact={row.original} />
        </div>
      ),
    },
  ]
}
