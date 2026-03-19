import { InfoIcon } from "lucide-react"
import { useEffect, useState } from "react"
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
import useCustomToast from "@/hooks/useCustomToast"
import { useSMSDetails } from "@/hooks/useSMSDetails"

interface SMSDetailsProps {
  id: string
}

export default function SMSDetails({ id }: SMSDetailsProps) {
  const { t } = useTranslation()
  const [isOpen, setIsOpen] = useState(false)
  const { showErrorToast } = useCustomToast()

  const { data: sms, isLoading, error } = useSMSDetails(id, isOpen)

  useEffect(() => {
    if (error) {
      showErrorToast(error.message)
    }
  }, [error, showErrorToast])

  return (
    <>
      <DropdownMenuItem
        onSelect={(e) => e.preventDefault()}
        onClick={() => setIsOpen(true)}
      >
        <InfoIcon className="mr-2 h-4 w-4" />
        {t("sms.details")}
      </DropdownMenuItem>
      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{t("sms.smsDetails")}</DialogTitle>

            {isLoading && (
              <DialogDescription>{t("sms.loadingMessage")}</DialogDescription>
            )}

            {sms && <DialogDescription>{sms.body}</DialogDescription>}
          </DialogHeader>

          <DialogFooter className="mt-4">
            <DialogClose asChild>
              <Button variant="outline">{t("common.close")}</Button>
            </DialogClose>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
