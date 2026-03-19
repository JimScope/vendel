import { useTranslation } from "react-i18next"
import { Button } from "@/components/ui/button"
import { useExportData } from "@/hooks/useAccountMutations"
import DeleteConfirmation from "./DeleteConfirmation"

const ExportData = () => {
  const { t } = useTranslation()
  const mutation = useExportData()

  const handleExport = () => {
    mutation.mutate(undefined, {
      onSuccess: (response) => {
        const blob = new Blob([JSON.stringify(response, null, 2)], {
          type: "application/json",
        })
        const url = URL.createObjectURL(blob)
        const a = document.createElement("a")
        a.href = url
        a.download = "vendel-export.json"
        a.click()
        URL.revokeObjectURL(url)
      },
    })
  }

  return (
    <div className="max-w-md mt-4 rounded-lg border p-4">
      <h3 className="font-semibold">{t("settings.exportData")}</h3>
      <p className="mt-1 text-sm text-muted-foreground">
        {t("settings.exportDataDesc")}
      </p>
      <Button
        variant="outline"
        className="mt-3"
        onClick={handleExport}
        disabled={mutation.isPending}
      >
        {mutation.isPending
          ? t("settings.exporting")
          : t("settings.exportButton")}
      </Button>
    </div>
  )
}

const DeleteAccount = () => {
  const { t } = useTranslation()
  return (
    <div className="flex flex-col gap-4">
      <ExportData />
      <div className="max-w-md rounded-lg border border-destructive/50 p-4">
        <h3 className="font-semibold text-destructive">
          {t("settings.deleteAccount")}
        </h3>
        <p className="mt-1 text-sm text-muted-foreground">
          {t("settings.deleteAccountDesc")}
        </p>
        <DeleteConfirmation />
      </div>
    </div>
  )
}

export default DeleteAccount
