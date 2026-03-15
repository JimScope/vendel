import ConfirmDeleteDialog from "@/components/Common/ConfirmDeleteDialog"
import { useDeleteScheduledSMS } from "@/hooks/useScheduledSMSMutations"

interface DeleteScheduledSMSProps {
  id: string
  onSuccess: () => void
}

const DeleteScheduledSMS = ({ id, onSuccess }: DeleteScheduledSMSProps) => {
  const mutation = useDeleteScheduledSMS()

  return (
    <ConfirmDeleteDialog
      triggerLabel="Delete Schedule"
      title="Delete Scheduled SMS"
      description="Are you sure you want to delete this scheduled SMS? This action cannot be undone."
      isPending={mutation.isPending}
      onConfirm={(close) => {
        mutation.mutate(id, { onSuccess: () => { close(); onSuccess() } })
      }}
    />
  )
}

export default DeleteScheduledSMS
