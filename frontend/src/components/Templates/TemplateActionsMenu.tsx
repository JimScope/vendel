import { EllipsisVertical } from "lucide-react"
import { useState } from "react"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import type { SMSTemplate } from "@/types/collections"
import DeleteTemplate from "./DeleteTemplate"
import EditTemplate from "./EditTemplate"

interface TemplateActionsMenuProps {
  template: SMSTemplate
}

export const TemplateActionsMenu = ({ template }: TemplateActionsMenuProps) => {
  const [open, setOpen] = useState(false)

  return (
    <DropdownMenu open={open} onOpenChange={setOpen}>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon">
          <EllipsisVertical />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <EditTemplate template={template} onSuccess={() => setOpen(false)} />
        <DropdownMenuSeparator />
        <DeleteTemplate id={template.id} onSuccess={() => setOpen(false)} />
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
