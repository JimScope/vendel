import { createFileRoute } from "@tanstack/react-router"
import { FileText } from "lucide-react"
import { Suspense } from "react"

import { DataTable } from "@/components/Common/DataTable"
import PendingTemplates from "@/components/Pending/PendingTemplates"
import AddTemplate from "@/components/Templates/AddTemplate"
import { columns } from "@/components/Templates/columns"
import useAppConfig from "@/hooks/useAppConfig"
import { useTemplateListSuspense } from "@/hooks/useTemplateList"

export const Route = createFileRoute("/_layout/templates")({
  component: Templates,
})

function TemplatesTableContent() {
  const { data: templates } = useTemplateListSuspense()

  if (!templates?.data || templates.data.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center text-center py-12">
        <div className="rounded-full bg-muted p-4 mb-4">
          <FileText className="h-8 w-8 text-muted-foreground" />
        </div>
        <h2 className="text-lg font-semibold">No templates yet</h2>
        <p className="text-muted-foreground">
          Create a template to save reusable SMS messages
        </p>
        <AddTemplate />
      </div>
    )
  }

  return (
    <DataTable
      columns={columns}
      data={templates?.data ?? []}
      caption="SMS templates"
    />
  )
}

function TemplatesTable() {
  return (
    <Suspense fallback={<PendingTemplates />}>
      <TemplatesTableContent />
    </Suspense>
  )
}

function Templates() {
  const { config } = useAppConfig()

  return (
    <div className="flex flex-col gap-6">
      <title>{`Templates - ${config.appName}`}</title>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl">Templates</h1>
          <p className="text-muted-foreground">
            Manage reusable SMS message templates
          </p>
        </div>
        <AddTemplate />
      </div>
      <TemplatesTable />
    </div>
  )
}
