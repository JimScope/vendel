import { Pause, Play } from "lucide-react"

import { DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { useUpdateScheduledSMS } from "@/hooks/useScheduledSMSMutations"
import type { ScheduledSMS } from "@/types/collections"

interface PauseScheduledSMSProps {
  schedule: ScheduledSMS
  onSuccess: () => void
}

const PauseScheduledSMS = ({ schedule, onSuccess }: PauseScheduledSMSProps) => {
  const isPaused = schedule.status === "paused"
  const updateMutation = useUpdateScheduledSMS(schedule.id)

  const handleToggle = () => {
    updateMutation.mutate(
      { status: isPaused ? "active" : "paused" },
      { onSuccess },
    )
  }

  return (
    <DropdownMenuItem onClick={handleToggle}>
      {isPaused ? <Play /> : <Pause />}
      {isPaused ? "Resume" : "Pause"}
    </DropdownMenuItem>
  )
}

export default PauseScheduledSMS
