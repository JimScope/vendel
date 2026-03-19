import type { ColumnDef } from "@tanstack/react-table"
import type { TFunction } from "i18next"

import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import type { User } from "@/types/collections"
import { UserActionsMenu } from "./UserActionsMenu"

export interface UserTableData extends User {
  isCurrentUser: boolean
}

export function getColumns(t: TFunction): ColumnDef<UserTableData>[] {
  return [
    {
      accessorKey: "full_name",
      header: t("admin.fullName"),
      cell: ({ row }) => {
        const fullName = row.original.full_name
        return (
          <div className="flex items-center gap-2">
            <span
              className={cn(
                "font-medium",
                !fullName && "text-muted-foreground",
              )}
            >
              {fullName || t("common.na")}
            </span>
            {row.original.isCurrentUser && (
              <Badge variant="outline" className="text-xs">
                {t("admin.you")}
              </Badge>
            )}
          </div>
        )
      },
    },
    {
      accessorKey: "email",
      header: t("common.email"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">{row.original.email}</span>
      ),
    },
    {
      accessorKey: "is_superuser",
      header: t("admin.role"),
      cell: ({ row }) => (
        <Badge variant={row.original.is_superuser ? "default" : "secondary"}>
          {row.original.is_superuser
            ? t("admin.roleSuperuser")
            : t("admin.roleUser")}
        </Badge>
      ),
    },
    {
      accessorKey: "verified",
      header: t("common.status"),
      cell: ({ row }) => (
        <div className="flex items-center gap-2">
          <span
            className={cn(
              "size-2 rounded-full",
              row.original.verified ? "bg-green-500" : "bg-gray-400",
            )}
          />
          <span
            className={row.original.verified ? "" : "text-muted-foreground"}
          >
            {row.original.verified
              ? t("admin.statusVerified")
              : t("admin.statusUnverified")}
          </span>
        </div>
      ),
    },
    {
      id: "actions",
      header: () => <span className="sr-only">{t("common.actions")}</span>,
      cell: ({ row }) => (
        <div className="flex justify-end">
          <UserActionsMenu user={row.original} />
        </div>
      ),
    },
  ]
}
