import ConfirmDeleteDialog from "@/components/Common/ConfirmDeleteDialog"
import { useDeleteApiKey } from "@/hooks/useApiKeyMutations"

interface DeleteApiKeyProps {
  id: string
  onSuccess: () => void
}

const DeleteApiKey = ({ id, onSuccess }: DeleteApiKeyProps) => {
  const mutation = useDeleteApiKey()

  return (
    <ConfirmDeleteDialog
      triggerLabel="Delete"
      title="Delete API Key"
      description="Are you sure you want to permanently delete this API key? This action cannot be undone."
      isPending={mutation.isPending}
      onConfirm={(close) => {
        mutation.mutate(id, { onSuccess: () => { close(); onSuccess() } })
      }}
    />
  )
}

export default DeleteApiKey
