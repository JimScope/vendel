import { Trash2 } from "lucide-react"
import { useState } from "react"
import { useForm } from "react-hook-form"

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
import { useDeleteScheduledSMS } from "@/hooks/useScheduledSMSMutations"

interface DeleteScheduledSMSProps {
  id: string
  onSuccess: () => void
}

const DeleteScheduledSMS = ({ id, onSuccess }: DeleteScheduledSMSProps) => {
  const [isOpen, setIsOpen] = useState(false)
  const { handleSubmit } = useForm()

  const deleteScheduledSMSMutation = useDeleteScheduledSMS()

  const onSubmit = async () => {
    deleteScheduledSMSMutation.mutate(id, {
      onSuccess: () => {
        setIsOpen(false)
        onSuccess()
      },
    })
  }

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DropdownMenuItem
        variant="destructive"
        onSelect={(e) => e.preventDefault()}
        onClick={() => setIsOpen(true)}
      >
        <Trash2 />
        Delete Schedule
      </DropdownMenuItem>
      <DialogContent className="sm:max-w-md">
        <form onSubmit={handleSubmit(onSubmit)}>
          <DialogHeader>
            <DialogTitle>Delete Scheduled SMS</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete this scheduled SMS? This action
              cannot be undone.
            </DialogDescription>
          </DialogHeader>

          <DialogFooter className="mt-4">
            <DialogClose asChild>
              <Button
                variant="outline"
                disabled={deleteScheduledSMSMutation.isPending}
              >
                Cancel
              </Button>
            </DialogClose>
            <LoadingButton
              variant="destructive"
              type="submit"
              loading={deleteScheduledSMSMutation.isPending}
            >
              Delete
            </LoadingButton>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

export default DeleteScheduledSMS
