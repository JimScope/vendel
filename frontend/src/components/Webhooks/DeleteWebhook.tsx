import ConfirmDeleteDialog from "@/components/Common/ConfirmDeleteDialog"
import { useDeleteWebhook } from "@/hooks/useWebhookMutations"

interface DeleteWebhookProps {
  id: string
  onSuccess: () => void
}

const DeleteWebhook = ({ id, onSuccess }: DeleteWebhookProps) => {
  const mutation = useDeleteWebhook()

  return (
    <ConfirmDeleteDialog
      triggerLabel="Delete Webhook"
      title="Delete Webhook"
      description="Are you sure you want to delete this webhook? You will no longer receive notifications at this endpoint."
      isPending={mutation.isPending}
      onConfirm={(close) => {
        mutation.mutate(id, { onSuccess: () => { close(); onSuccess() } })
      }}
    />
  )
}

export default DeleteWebhook
