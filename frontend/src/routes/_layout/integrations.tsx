import { createFileRoute } from "@tanstack/react-router"
import { Key, Plus } from "lucide-react"
import { Suspense, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

import AddApiKey from "@/components/ApiKeys/AddApiKey"
import { getColumns } from "@/components/ApiKeys/columns"
import { DataTable } from "@/components/Common/DataTable"
import PendingApiKeys from "@/components/Pending/PendingApiKeys"
import { Button } from "@/components/ui/button"
import { useApiKeyListSuspense } from "@/hooks/useApiKeyList"
import useAppConfig from "@/hooks/useAppConfig"
import type { ApiKey } from "@/types/collections"

export const Route = createFileRoute("/_layout/integrations")({
  component: Integrations,
})

function ApiKeysEmptyState({ onAddApiKey }: { onAddApiKey: () => void }) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-col items-center justify-center text-center py-12">
      <div className="rounded-full bg-muted p-4 mb-4">
        <Key className="h-8 w-8 text-muted-foreground" />
      </div>
      <h2 className="text-lg font-semibold">{t("apiKeys.noApiKeys")}</h2>
      <p className="text-muted-foreground">{t("apiKeys.noApiKeysDesc")}</p>
      <Button className="my-4" onClick={onAddApiKey}>
        <Plus />
        {t("apiKeys.createApiKey")}
      </Button>
    </div>
  )
}

function ApiKeysTableContent({ onAddApiKey }: { onAddApiKey: () => void }) {
  const { t } = useTranslation()
  const columns = useMemo(() => getColumns(t), [t])
  const { data: apiKeys } = useApiKeyListSuspense()

  if (!apiKeys?.data || apiKeys.data.length === 0) {
    return <ApiKeysEmptyState onAddApiKey={onAddApiKey} />
  }

  return (
    <DataTable
      columns={columns}
      data={(apiKeys?.data ?? []) as unknown as ApiKey[]}
      caption={t("apiKeys.title")}
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
  const { t } = useTranslation()
  const { config } = useAppConfig()
  const [addApiKeyOpen, setAddApiKeyOpen] = useState(false)

  return (
    <div className="flex flex-col gap-6">
      <title>{`${t("apiKeys.title")} - ${config.appName}`}</title>
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl">{t("apiKeys.title")}</h1>
          <p className="text-muted-foreground">{t("apiKeys.description")}</p>
        </div>
        <AddApiKey open={addApiKeyOpen} onOpenChange={setAddApiKeyOpen} />
      </div>
      <ApiKeysTable onAddApiKey={() => setAddApiKeyOpen(true)} />
    </div>
  )
}
