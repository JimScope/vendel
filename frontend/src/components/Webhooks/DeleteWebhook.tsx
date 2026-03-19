import { useTranslation } from "react-i18next"

import ConfirmDeleteDialog from "@/components/Common/ConfirmDeleteDialog"
import { useDeleteWebhook } from "@/hooks/useWebhookMutations"

interface DeleteWebhookProps {
  id: string
  onSuccess: () => void
}

const DeleteWebhook = ({ id, onSuccess }: DeleteWebhookProps) => {
  const { t } = useTranslation()
  const mutation = useDeleteWebhook()

  return (
    <ConfirmDeleteDialog
      triggerLabel={t("webhooks.deleteWebhook")}
      title={t("webhooks.deleteWebhook")}
      description={t("webhooks.deleteMsg")}
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

export default DeleteWebhook
