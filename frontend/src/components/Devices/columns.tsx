import type { ColumnDef } from "@tanstack/react-table"
import { Smartphone, Usb } from "lucide-react"

import { DeviceActionsMenu } from "@/components/Devices/DeviceActionsMenu"
import { formatDate } from "@/lib/utils"

export const columns: ColumnDef<Record<string, any>>[] = [
  {
    accessorKey: "name",
    header: "Name",
    cell: ({ row }) => <span className="font-medium">{row.original.name}</span>,
  },
  {
    accessorKey: "device_type",
    header: "Type",
    cell: ({ row }) => {
      const type = row.original.device_type || "android"
      return (
        <span className="inline-flex items-center gap-1.5 text-muted-foreground">
          {type === "modem" ? (
            <Usb className="h-4 w-4" />
          ) : (
            <Smartphone className="h-4 w-4" />
          )}
          {type === "modem" ? "USB Modem" : "Android"}
        </span>
      )
    },
  },
  {
    accessorKey: "phone_number",
    header: "Phone Number",
    cell: ({ row }) => (
      <span className="text-muted-foreground">{row.original.phone_number}</span>
    ),
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
        <DeviceActionsMenu device={row.original} />
      </div>
    ),
  },
]
