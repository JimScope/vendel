import { createFileRoute } from "@tanstack/react-router"
import { Clock } from "lucide-react"
import { Suspense } from "react"

import { DataTable } from "@/components/Common/DataTable"
import PendingScheduledSMS from "@/components/Pending/PendingScheduledSMS"
import AddScheduledSMS from "@/components/ScheduledSMS/AddScheduledSMS"
import { columns } from "@/components/ScheduledSMS/columns"
import useAppConfig from "@/hooks/useAppConfig"
import { useScheduledSMSListSuspense } from "@/hooks/useScheduledSMSList"

export const Route = createFileRoute("/_layout/scheduled")({
  component: Scheduled,
})

function ScheduledTableContent() {
  const { data: scheduled } = useScheduledSMSListSuspense()

  if (!scheduled?.data || scheduled.data.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center text-center py-12">
        <div className="rounded-full bg-muted p-4 mb-4">
          <Clock className="h-8 w-8 text-muted-foreground" />
        </div>
        <h3 className="text-lg font-semibold">No scheduled messages</h3>
        <p className="text-muted-foreground">
          Schedule an SMS to be sent at a specific time or on a recurring basis
        </p>
      </div>
    )
  }

  return <DataTable columns={columns} data={scheduled?.data ?? []} />
}

function ScheduledTable() {
  return (
    <Suspense fallback={<PendingScheduledSMS />}>
      <ScheduledTableContent />
    </Suspense>
  )
}

function Scheduled() {
  const { config } = useAppConfig()

  return (
    <div className="flex flex-col gap-6">
      <title>{`Scheduled - ${config.appName}`}</title>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl">Scheduled SMS</h1>
          <p className="text-muted-foreground">
            Schedule messages for one-time or recurring delivery
          </p>
        </div>
        <AddScheduledSMS />
      </div>
      <ScheduledTable />
    </div>
  )
}
