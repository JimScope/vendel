import { createFileRoute } from "@tanstack/react-router"
import { CreditCard } from "lucide-react"
import { Suspense } from "react"

import { columns } from "@/components/Billing/columns"
import { DataTable } from "@/components/Common/DataTable"
import PendingBilling from "@/components/Pending/PendingBilling"
import useAppConfig from "@/hooks/useAppConfig"
import { usePaymentListSuspense } from "@/hooks/usePaymentList"

export const Route = createFileRoute("/_layout/billing")({
  component: Billing,
})

function BillingTableContent() {
  const { data: payments } = usePaymentListSuspense()

  if (!payments?.data || payments.data.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center text-center py-12">
        <div className="rounded-full bg-muted p-4 mb-4">
          <CreditCard className="h-8 w-8 text-muted-foreground" />
        </div>
        <h2 className="text-lg font-semibold">No billing history</h2>
        <p className="text-muted-foreground">
          Your payment history will appear here after subscribing to a plan
        </p>
      </div>
    )
  }

  return (
    <DataTable
      columns={columns}
      data={payments?.data ?? []}
      caption="Payment history"
    />
  )
}

function BillingTable() {
  return (
    <Suspense fallback={<PendingBilling />}>
      <BillingTableContent />
    </Suspense>
  )
}

function Billing() {
  const { config } = useAppConfig()

  return (
    <div className="flex flex-col gap-6">
      <title>{`Billing - ${config.appName}`}</title>
      <div>
        <h1 className="text-2xl">Billing</h1>
        <p className="text-muted-foreground">
          View your payment history and invoices
        </p>
      </div>
      <BillingTable />
    </div>
  )
}
