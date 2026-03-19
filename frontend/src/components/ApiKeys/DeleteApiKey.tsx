import { useTranslation } from "react-i18next"

import ConfirmDeleteDialog from "@/components/Common/ConfirmDeleteDialog"
import { useDeleteApiKey } from "@/hooks/useApiKeyMutations"

interface DeleteApiKeyProps {
  id: string
  onSuccess: () => void
}

const DeleteApiKey = ({ id, onSuccess }: DeleteApiKeyProps) => {
  const { t } = useTranslation()
  const mutation = useDeleteApiKey()

  return (
    <ConfirmDeleteDialog
      triggerLabel={t("common.delete")}
      title={t("apiKeys.deleteKey")}
      description={t("apiKeys.deleteMsg")}
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

export default DeleteApiKey
