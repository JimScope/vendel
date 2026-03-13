import { createFileRoute } from "@tanstack/react-router"
import { Key, Plus } from "lucide-react"
import { Suspense, useState } from "react"

import AddApiKey from "@/components/ApiKeys/AddApiKey"
import { columns } from "@/components/ApiKeys/columns"
import { DataTable } from "@/components/Common/DataTable"
import PendingApiKeys from "@/components/Pending/PendingApiKeys"
import { Button } from "@/components/ui/button"
import { useApiKeyListSuspense } from "@/hooks/useApiKeyList"
import useAppConfig from "@/hooks/useAppConfig"

export const Route = createFileRoute("/_layout/integrations")({
  component: Integrations,
})

function ApiKeysEmptyState({ onAddApiKey }: { onAddApiKey: () => void }) {
  return (
    <div className="flex flex-col items-center justify-center text-center py-12">
      <div className="rounded-full bg-muted p-4 mb-4">
        <Key className="h-8 w-8 text-muted-foreground" />
      </div>
      <h2 className="text-lg font-semibold">No API keys</h2>
      <p className="text-muted-foreground">
        Create an API key to access the API programmatically
      </p>
      <Button className="my-4" onClick={onAddApiKey}>
        <Plus />
        Create API Key
      </Button>
    </div>
  )
}

function ApiKeysTableContent({ onAddApiKey }: { onAddApiKey: () => void }) {
  const { data: apiKeys } = useApiKeyListSuspense()

  if (!apiKeys?.data || apiKeys.data.length === 0) {
    return <ApiKeysEmptyState onAddApiKey={onAddApiKey} />
  }

  return (
    <DataTable
      columns={columns}
      data={apiKeys?.data ?? []}
      caption="API keys"
    />
  )
}

function ApiKeysTable({ onAddApiKey }: { onAddApiKey: () => void }) {
  return (
    <Suspense fallback={<PendingApiKeys />}>
      <ApiKeysTableContent onAddApiKey={onAddApiKey} />
    </Suspense>
  )
}

function Integrations() {
  const { config } = useAppConfig()
  const [addApiKeyOpen, setAddApiKeyOpen] = useState(false)

  return (
    <div className="flex flex-col gap-6">
      <title>{`Integrations - ${config.appName}`}</title>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl">Integrations</h1>
          <p className="text-muted-foreground">
            Manage API keys for programmatic access to the API
          </p>
        </div>
        <AddApiKey open={addApiKeyOpen} onOpenChange={setAddApiKeyOpen} />
      </div>
      <ApiKeysTable onAddApiKey={() => setAddApiKeyOpen(true)} />
    </div>
  )
}
