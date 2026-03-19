import { useTranslation } from "react-i18next"

import ConfirmDeleteDialog from "@/components/Common/ConfirmDeleteDialog"
import { useDeleteTemplate } from "@/hooks/useTemplateMutations"

interface DeleteTemplateProps {
  id: string
  onSuccess: () => void
}

const DeleteTemplate = ({ id, onSuccess }: DeleteTemplateProps) => {
  const { t } = useTranslation()
  const mutation = useDeleteTemplate()

  return (
    <ConfirmDeleteDialog
      triggerLabel={t("templates.deleteTemplate")}
      title={t("templates.deleteTemplate")}
      description={t("templates.deleteMsg")}
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
