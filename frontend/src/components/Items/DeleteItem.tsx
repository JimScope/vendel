import ConfirmDeleteDialog from "@/components/Common/ConfirmDeleteDialog"
import { useDeleteItem } from "@/hooks/useItemMutations"

interface DeleteItemProps {
  id: string
  onSuccess: () => void
}

const DeleteItem = ({ id, onSuccess }: DeleteItemProps) => {
  const mutation = useDeleteItem()

  return (
    <ConfirmDeleteDialog
      triggerLabel="Delete Item"
      title="Delete Item"
      description="This item will be permanently deleted. Are you sure? You will not be able to undo this action."
      isPending={mutation.isPending}
      onConfirm={(close) => {
        mutation.mutate(id, { onSuccess: () => { close(); onSuccess() } })
      }}
    />
  )
}

export default DeleteItem
