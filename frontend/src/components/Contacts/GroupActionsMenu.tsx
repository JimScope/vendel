import { EllipsisVertical } from "lucide-react"
import { useState } from "react"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import type { ContactGroup } from "@/types/collections"
import DeleteGroup from "./DeleteGroup"
import EditGroup from "./EditGroup"

interface GroupActionsMenuProps {
  group: ContactGroup
}

export const GroupActionsMenu = ({ group }: GroupActionsMenuProps) => {
  const [open, setOpen] = useState(false)

  return (
    <DropdownMenu open={open} onOpenChange={setOpen}>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon">
          <EllipsisVertical />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <EditGroup group={group} onSuccess={() => setOpen(false)} />
        <DeleteGroup id={group.id} onSuccess={() => setOpen(false)} />
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
