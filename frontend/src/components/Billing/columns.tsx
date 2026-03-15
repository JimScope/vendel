import type { ColumnDef } from "@tanstack/react-table"
import { ExternalLink } from "lucide-react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import type { Payment } from "@/types/collections"

function formatDate(dateString: string): string {
  if (!dateString) return "-"
  return new Date(dateString).toLocaleDateString()
}

function formatCurrency(amount: number, currency: string): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: currency || "USD",
  }).format(amount)
}

const statusStyles: Record<string, string> = {
  completed:
    "bg-green-500/10 text-green-700 border-green-500/20 dark:text-green-400",
  pending:
    "bg-yellow-500/10 text-yellow-700 border-yellow-500/20 dark:text-yellow-400",
  failed: "bg-red-500/10 text-red-700 border-red-500/20 dark:text-red-400",
  refunded:
    "bg-blue-500/10 text-blue-700 border-blue-500/20 dark:text-blue-400",
}

export const columns: ColumnDef<Payment>[] = [
  {
    accessorKey: "paid_at",
    header: "Date",
    cell: ({ row }) => (
      <span className="text-sm">
        {formatDate(row.original.paid_at || row.original.created)}
      </span>
    ),
  },
  {
    id: "period",
    header: "Period",
    cell: ({ row }) => (
      <span className="text-sm text-muted-foreground">
        {formatDate(row.original.period_start)} &mdash;{" "}
        {formatDate(row.original.period_end)}
      </span>
    ),
  },
  {
    accessorKey: "amount",
    header: "Amount",
    cell: ({ row }) => (
      <span className="font-medium">
        {formatCurrency(
          row.original.amount ?? 0,
          row.original.currency ?? "USD",
        )}
      </span>
    ),
  },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => {
      const status = row.original.status ?? "pending"
      return (
        <Badge variant="outline" className={cn(statusStyles[status])}>
          {status}
        </Badge>
      )
    },
  },
  {
    accessorKey: "provider",
    header: "Provider",
    cell: ({ row }) => (
      <span className="text-sm capitalize">{row.original.provider || "-"}</span>
    ),
  },
  {
    id: "actions",
    header: () => <span className="sr-only">Actions</span>,
    cell: ({ row }) => {
      const invoiceUrl = row.original.provider_invoice_url
      if (!invoiceUrl) return null
      return (
        <div className="flex justify-end">
          <Button variant="ghost" size="icon" asChild>
            <a href={invoiceUrl} target="_blank" rel="noopener noreferrer">
              <ExternalLink className="h-4 w-4" />
              <span className="sr-only">View invoice</span>
            </a>
          </Button>
        </div>
      )
    },
  },
]
