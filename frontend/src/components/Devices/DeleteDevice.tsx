import ConfirmDeleteDialog from "@/components/Common/ConfirmDeleteDialog"
import { useDeleteDevice } from "@/hooks/useDeviceMutations"

interface DeleteDeviceProps {
  id: string
  onSuccess: () => void
}

const DeleteDevice = ({ id, onSuccess }: DeleteDeviceProps) => {
  const mutation = useDeleteDevice()

  return (
    <ConfirmDeleteDialog
      triggerLabel="Delete Device"
      title="Delete Device"
      description="Are you sure you want to delete this device? This action cannot be undone."
      isPending={mutation.isPending}
      onConfirm={(close) => {
        mutation.mutate(id, { onSuccess: () => { close(); onSuccess() } })
      }}
    />
  )
}

export default DeleteDevice
