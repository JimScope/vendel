import { useTranslation } from "react-i18next"

import ConfirmDeleteDialog from "@/components/Common/ConfirmDeleteDialog"
import { useDeleteContactGroup } from "@/hooks/useContactGroupMutations"

interface DeleteGroupProps {
  id: string
  onSuccess: () => void
}

const DeleteGroup = ({ id, onSuccess }: DeleteGroupProps) => {
  const { t } = useTranslation()
  const mutation = useDeleteContactGroup()

  return (
    <ConfirmDeleteDialog
      triggerLabel={t("contacts.deleteGroup")}
      title={t("contacts.deleteGroup")}
      description={t("contacts.deleteGroupDesc")}
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

export default DeleteGroup
