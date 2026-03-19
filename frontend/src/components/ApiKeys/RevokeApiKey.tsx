import { Ban } from "lucide-react"
import { useState } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { LoadingButton } from "@/components/ui/loading-button"
import { useRevokeApiKey } from "@/hooks/useApiKeyMutations"

interface RevokeApiKeyProps {
  id: string
  onSuccess: () => void
}

const RevokeApiKey = ({ id, onSuccess }: RevokeApiKeyProps) => {
  const { t } = useTranslation()
  const [isOpen, setIsOpen] = useState(false)
  const { handleSubmit } = useForm()

  const revokeApiKeyMutation = useRevokeApiKey()

  const onSubmit = async () => {
    revokeApiKeyMutation.mutate(id, {
      onSuccess: () => {
        setIsOpen(false)
        onSuccess()
      },
    })
  }

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DropdownMenuItem
        onSelect={(e) => e.preventDefault()}
        onClick={() => setIsOpen(true)}
      >
        <Ban />
        {t("apiKeys.revokeKey")}
      </DropdownMenuItem>
      <DialogContent className="sm:max-w-md">
        <form onSubmit={handleSubmit(onSubmit)}>
          <DialogHeader>
            <DialogTitle>{t("apiKeys.revokeTitle")}</DialogTitle>
            <DialogDescription>{t("apiKeys.revokeDesc")}</DialogDescription>
          </DialogHeader>

          <DialogFooter className="mt-4">
            <DialogClose asChild>
              <Button
                variant="outline"
                disabled={revokeApiKeyMutation.isPending}
              >
                {t("common.cancel")}
              </Button>
            </DialogClose>
            <LoadingButton
              variant="destructive"
              type="submit"
              loading={revokeApiKeyMutation.isPending}
            >
              {t("apiKeys.revokeKey")}
            </LoadingButton>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

export default RevokeApiKey
