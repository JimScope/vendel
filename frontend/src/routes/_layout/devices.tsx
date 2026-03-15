import { createFileRoute } from "@tanstack/react-router"
import { Plus, Smartphone } from "lucide-react"
import { Suspense, useMemo, useState } from "react"
import { DataTable } from "@/components/Common/DataTable"
import AddDevice from "@/components/Devices/AddDevice"
import { getColumns } from "@/components/Devices/columns"
import ModemAgentDownload from "@/components/Devices/ModemAgentDownload"
import PendingDevices from "@/components/Pending/PendingDevices"
import { Button } from "@/components/ui/button"
import useAppConfig from "@/hooks/useAppConfig"
import { useDeviceListSuspense } from "@/hooks/useDeviceList"
import type { Device } from "@/types/collections"
import { useModemStatus } from "@/hooks/useModemStatus"

export const Route = createFileRoute("/_layout/devices")({
  component: Devices,
})

function DevicesEmptyState({ onAddDevice }: { onAddDevice: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center text-center py-12">
      <div className="rounded-full bg-muted p-4 mb-4">
        <Smartphone className="h-8 w-8 text-muted-foreground" />
      </div>
      <h2 className="text-lg font-semibold">No devices registered</h2>
      <p className="text-muted-foreground">
        Add a device to start sending SMS messages
      </p>
      <Button className="my-4" onClick={onAddDevice}>
        <Plus />
        Add Device
      </Button>
    </div>
  )
}

function DevicesTableContent({ onAddDevice }: { onAddDevice: () => void }) {
  const { data: devices } = useDeviceListSuspense()
  const { data: modemStatus } = useModemStatus()
  const columns = useMemo(() => getColumns(modemStatus), [modemStatus])

  if (!devices?.data || devices.data.length === 0) {
    return <DevicesEmptyState onAddDevice={onAddDevice} />
  }

  return (
    <DataTable
      columns={columns}
      data={(devices?.data ?? []) as unknown as Device[]}
      caption="Registered devices"
    />
  )
}

function DevicesTable({ onAddDevice }: { onAddDevice: () => void }) {
  return (
    <Suspense fallback={<PendingDevices />}>
      <DevicesTableContent onAddDevice={onAddDevice} />
    </Suspense>
  )
}

function Devices() {
  const { config } = useAppConfig()
  const [addDeviceOpen, setAddDeviceOpen] = useState(false)

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
        <AddDevice open={addDeviceOpen} onOpenChange={setAddDeviceOpen} />
      </div>
      <ModemAgentDownload />
      <DevicesTable onAddDevice={() => setAddDeviceOpen(true)} />
    </div>
  )
}
