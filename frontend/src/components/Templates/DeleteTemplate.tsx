import ConfirmDeleteDialog from "@/components/Common/ConfirmDeleteDialog"
import { useDeleteTemplate } from "@/hooks/useTemplateMutations"

interface DeleteTemplateProps {
  id: string
  onSuccess: () => void
}

const DeleteTemplate = ({ id, onSuccess }: DeleteTemplateProps) => {
  const mutation = useDeleteTemplate()

  return (
    <ConfirmDeleteDialog
      triggerLabel="Delete Template"
      title="Delete Template"
      description="Are you sure you want to delete this template? This action cannot be undone."
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

export default DeleteTemplate
