import { useTranslation } from "react-i18next"

import ConfirmDeleteDialog from "@/components/Common/ConfirmDeleteDialog"
import { useDeleteContact } from "@/hooks/useContactMutations"

interface DeleteContactProps {
  id: string
  onSuccess: () => void
}

const DeleteContact = ({ id, onSuccess }: DeleteContactProps) => {
  const { t } = useTranslation()
  const mutation = useDeleteContact()

  return (
    <ConfirmDeleteDialog
      triggerLabel={t("contacts.deleteContact")}
      title={t("contacts.deleteContact")}
      description={t("contacts.deleteContactDesc")}
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

export default DeleteContact
