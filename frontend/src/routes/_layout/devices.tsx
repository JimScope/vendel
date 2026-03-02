import { createFileRoute } from "@tanstack/react-router"
import { Smartphone } from "lucide-react"
import { Suspense, useMemo } from "react"

import { DataTable } from "@/components/Common/DataTable"
import AddDevice from "@/components/Devices/AddDevice"
import { getColumns } from "@/components/Devices/columns"
import ModemAgentDownload from "@/components/Devices/ModemAgentDownload"
import PendingDevices from "@/components/Pending/PendingDevices"
import useAppConfig from "@/hooks/useAppConfig"
import { useDeviceListSuspense } from "@/hooks/useDeviceList"
import { useModemStatus } from "@/hooks/useModemStatus"

export const Route = createFileRoute("/_layout/devices")({
  component: Devices,
})

function DevicesTableContent() {
  const { data: devices } = useDeviceListSuspense()
  const { data: modemStatus } = useModemStatus()
  const columns = useMemo(() => getColumns(modemStatus), [modemStatus])

  if (!devices?.data || devices.data.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center text-center py-12">
        <div className="rounded-full bg-muted p-4 mb-4">
          <Smartphone className="h-8 w-8 text-muted-foreground" />
        </div>
        <h2 className="text-lg font-semibold">No devices registered</h2>
        <p className="text-muted-foreground">
          Add a device to start sending SMS messages
        </p>
        <AddDevice />
      </div>
    )
  }

  return (
    <DataTable
      columns={columns}
      data={devices?.data ?? []}
      caption="Registered devices"
    />
  )
}

function DevicesTable() {
  return (
    <Suspense fallback={<PendingDevices />}>
      <DevicesTableContent />
    </Suspense>
  )
}

function Devices() {
  const { config } = useAppConfig()

  return (
    <div className="flex flex-col gap-6">
      <title>{`Devices - ${config.appName}`}</title>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl">Devices</h1>
          <p className="text-muted-foreground">
            Manage your registered SMS devices
          </p>
        </div>
        <AddDevice />
      </div>
      <ModemAgentDownload />
      <DevicesTable />
    </div>
  )
}
