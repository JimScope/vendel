import { EllipsisVertical } from "lucide-react"
import { useState } from "react"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import DeleteScheduledSMS from "./DeleteScheduledSMS"
import EditScheduledSMS from "./EditScheduledSMS"
import PauseScheduledSMS from "./PauseScheduledSMS"

interface ScheduledSMSActionsMenuProps {
  schedule: Record<string, any>
}

export const ScheduledSMSActionsMenu = ({
  schedule,
}: ScheduledSMSActionsMenuProps) => {
  const [open, setOpen] = useState(false)

  const isCompleted = schedule.status === "completed"

  return (
    <DropdownMenu open={open} onOpenChange={setOpen}>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon">
          <EllipsisVertical />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {!isCompleted && (
          <PauseScheduledSMS
            schedule={schedule}
            onSuccess={() => setOpen(false)}
          />
        )}
        <EditScheduledSMS
          schedule={schedule}
          onSuccess={() => setOpen(false)}
        />
        <DropdownMenuSeparator />
        <DeleteScheduledSMS id={schedule.id} onSuccess={() => setOpen(false)} />
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
