import type { ColumnDef } from "@tanstack/react-table"
import { Smartphone, Usb } from "lucide-react"

import { DeviceActionsMenu } from "@/components/Devices/DeviceActionsMenu"
import { formatDate } from "@/lib/utils"

export function getColumns(
  modemStatus?: Record<string, boolean>,
): ColumnDef<Record<string, any>>[] {
  return [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => (
        <span className="font-medium">{row.original.name}</span>
      ),
    },
    {
      accessorKey: "device_type",
      header: "Type",
      cell: ({ row }) => {
        const type = row.original.device_type || "android"
        const isModem = type === "modem"
        const isOnline = isModem && modemStatus?.[row.original.id]

        return (
          <span className="inline-flex items-center gap-1.5 text-muted-foreground">
            {isModem ? (
              <Usb className="h-4 w-4" />
            ) : (
              <Smartphone className="h-4 w-4" />
            )}
            {isModem ? "USB Modem" : "Android"}
            {isModem && (
              <span
                className={`ml-1 inline-block h-2 w-2 rounded-full ${isOnline ? "bg-emerald-500" : "bg-neutral-300"}`}
                title={isOnline ? "Online" : "Offline"}
              />
            )}
          </span>
        )
      },
    },
    {
      accessorKey: "phone_number",
      header: "Phone Number",
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {row.original.phone_number}
        </span>
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
}
