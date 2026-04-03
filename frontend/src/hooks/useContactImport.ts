import { useMutation, useQueryClient } from "@tanstack/react-query"

import pb from "@/lib/pocketbase"
import useCustomToast from "./useCustomToast"

interface ImportResult {
  imported: number
  skipped: number
}

export function useContactImport() {
  const queryClient = useQueryClient()
  const { showErrorToast } = useCustomToast()

  return useMutation({
    mutationFn: async (data: {
      file: File
      groupId?: string
    }): Promise<ImportResult> => {
      const formData = new FormData()
      formData.append("file", data.file)
      if (data.groupId) formData.append("group_id", data.groupId)
      return await pb.send("/api/contacts/import", {
        method: "POST",
        body: formData,
      })
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Failed to import contacts")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["contacts"] })
    },
  })
}
