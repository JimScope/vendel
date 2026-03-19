import { useTranslation } from "react-i18next"

import ConfirmDeleteDialog from "@/components/Common/ConfirmDeleteDialog"
import { useDeleteUser } from "@/hooks/useUserMutations"

interface DeleteUserProps {
  id: string
  onSuccess: () => void
}

const DeleteUser = ({ id, onSuccess }: DeleteUserProps) => {
  const { t } = useTranslation()
  const mutation = useDeleteUser()

  return (
    <ConfirmDeleteDialog
      triggerLabel={t("admin.deleteUser")}
      title={t("admin.deleteUser")}
      description={t("admin.deleteUserMsg")}
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

export default DeleteUser
