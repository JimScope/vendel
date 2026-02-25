import type { ColumnDef } from "@tanstack/react-table"

import { Badge } from "@/components/ui/badge"
import { cn, formatDate } from "@/lib/utils"
import { ApiKeyActionsMenu } from "./ApiKeyActionsMenu"

function formatRelativeDate(dateString: string): string {
  const date = new Date(dateString)
  const now = new Date()
  const diffMs = date.getTime() - now.getTime()
  const diffDays = Math.ceil(diffMs / (1000 * 60 * 60 * 24))

  if (diffDays < 0) return `${Math.abs(diffDays)}d ago`
  if (diffDays === 0) return "Today"
  if (diffDays === 1) return "Tomorrow"
  if (diffDays < 30) return `${diffDays}d`
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

export const columns: ColumnDef<Record<string, any>>[] = [
  {
    accessorKey: "name",
    header: "Name",
    cell: ({ row }) => <span className="font-medium">{row.original.name}</span>,
  },
  {
    accessorKey: "key_prefix",
    header: "Key",
    cell: ({ row }) => (
      <span className="font-mono text-sm text-muted-foreground">
        {row.original.key_prefix}
      </span>
    ),
  },
  {
    accessorKey: "is_active",
    header: "Status",
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
            {isActive ? "Active" : "Revoked"}
          </span>
        </div>
      )
    },
  },
  {
    accessorKey: "last_used_at",
    header: "Last Used",
    cell: ({ row }) => (
      <span className="text-muted-foreground">
        {row.original.last_used_at
          ? formatDate(row.original.last_used_at)
          : "Never"}
      </span>
    ),
  },
  {
    accessorKey: "expires_at",
    header: "Expires",
    cell: ({ row }) => {
      const expiresAt = row.original.expires_at
      if (!expiresAt) {
        return <span className="text-muted-foreground">Never</span>
      }
      if (isExpired(expiresAt)) {
        return (
          <Badge variant="destructive" className="text-[10px]">
            Expired {formatRelativeDate(expiresAt)}
          </Badge>
        )
      }
      if (isExpiringSoon(expiresAt)) {
        return (
          <Badge
            variant="outline"
            className="text-[10px] border-yellow-500 text-yellow-600"
          >
            {formatRelativeDate(expiresAt)}
          </Badge>
        )
      }
      return (
        <span className="text-muted-foreground">
          {formatRelativeDate(expiresAt)}
        </span>
      )
    },
  },
  {
    accessorKey: "created_at",
    header: "Created",
    cell: ({ row }) => (
      <span className="text-muted-foreground">
        {formatDate(row.original.created_at)}
      </span>
    ),
  },
  {
    id: "actions",
    header: () => <span className="sr-only">Actions</span>,
    cell: ({ row }) => (
      <div className="flex justify-end">
        <ApiKeyActionsMenu apiKey={row.original} />
      </div>
    ),
  },
]
