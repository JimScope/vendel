import { useTranslation } from "react-i18next"

import ConfirmDeleteDialog from "@/components/Common/ConfirmDeleteDialog"
import { useDeleteDevice } from "@/hooks/useDeviceMutations"

interface DeleteDeviceProps {
  id: string
  onSuccess: () => void
}

const DeleteDevice = ({ id, onSuccess }: DeleteDeviceProps) => {
  const { t } = useTranslation()
  const mutation = useDeleteDevice()

  return (
    <ConfirmDeleteDialog
      triggerLabel={t("devices.deleteDevice")}
      title={t("devices.deleteDevice")}
      description={t("devices.deleteDeviceMsg")}
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

export default DeleteDevice
