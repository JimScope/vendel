import { useState } from "react"
import { Button } from "@/components/ui/button"
import useCustomToast from "@/hooks/useCustomToast"
import pb from "@/lib/pocketbase"
import DeleteConfirmation from "./DeleteConfirmation"

const ExportData = () => {
  const [loading, setLoading] = useState(false)
  const { showErrorToast } = useCustomToast()

  const handleExport = async () => {
    setLoading(true)
    try {
      const response = await pb.send("/api/user/export", { method: "GET" })
      const blob = new Blob([JSON.stringify(response, null, 2)], {
        type: "application/json",
      })
      const url = URL.createObjectURL(blob)
      const a = document.createElement("a")
      a.href = url
      a.download = "vendel-export.json"
      a.click()
      URL.revokeObjectURL(url)
    } catch {
      showErrorToast("Failed to export data")
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-md mt-4 rounded-lg border p-4">
      <h3 className="font-semibold">Export Your Data</h3>
      <p className="mt-1 text-sm text-muted-foreground">
        Download a copy of all your data including messages, devices, webhooks,
        and account information.
      </p>
      <Button
        variant="outline"
        className="mt-3"
        onClick={handleExport}
        disabled={loading}
      >
        {loading ? "Exporting..." : "Export Data"}
      </Button>
    </div>
  )
}

const DeleteAccount = () => {
  return (
    <div className="flex flex-col gap-4">
      <ExportData />
      <div className="max-w-md rounded-lg border border-destructive/50 p-4">
        <h3 className="font-semibold text-destructive">Delete Account</h3>
        <p className="mt-1 text-sm text-muted-foreground">
          Permanently delete your account and all associated data.
        </p>
        <DeleteConfirmation />
      </div>
    </div>
  )
}

export default DeleteAccount
