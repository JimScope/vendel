import { useTranslation } from "react-i18next"
import ConfirmDeleteDialog from "@/components/Common/ConfirmDeleteDialog"
import { useDeleteScheduledSMS } from "@/hooks/useScheduledSMSMutations"

interface DeleteScheduledSMSProps {
  id: string
  onSuccess: () => void
}

const DeleteScheduledSMS = ({ id, onSuccess }: DeleteScheduledSMSProps) => {
  const { t } = useTranslation()
  const mutation = useDeleteScheduledSMS()

  return (
    <ConfirmDeleteDialog
      triggerLabel={t("scheduled.deleteScheduled")}
      title={t("scheduled.deleteScheduled")}
      description={t("scheduled.deleteMsg")}
      isPending={mutation.isPending}
      onConfirm={(close) => {
        mutation.mutate(id, {
          onSuccess: () => {
            close()
            onSuccess()
          },
        })
      }}
    />
  )
}

export default DeleteScheduledSMS
