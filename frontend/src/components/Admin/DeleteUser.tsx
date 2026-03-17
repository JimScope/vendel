import ConfirmDeleteDialog from "@/components/Common/ConfirmDeleteDialog"
import { useDeleteUser } from "@/hooks/useUserMutations"

interface DeleteUserProps {
  id: string
  onSuccess: () => void
}

const DeleteUser = ({ id, onSuccess }: DeleteUserProps) => {
  const mutation = useDeleteUser()

  return (
    <ConfirmDeleteDialog
      triggerLabel="Delete User"
      title="Delete User"
      description={
        <>
          All items associated with this user will also be{" "}
          <strong>permanently deleted.</strong> Are you sure? You will not be
          able to undo this action.
        </>
      }
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
