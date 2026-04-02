import { EllipsisVertical } from "lucide-react"
import { useState } from "react"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import type { Contact } from "@/types/collections"
import DeleteContact from "./DeleteContact"
import EditContact from "./EditContact"

interface ContactActionsMenuProps {
  contact: Contact
}

export const ContactActionsMenu = ({ contact }: ContactActionsMenuProps) => {
  const [open, setOpen] = useState(false)

  return (
    <DropdownMenu open={open} onOpenChange={setOpen}>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon">
          <EllipsisVertical />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <EditContact contact={contact} onSuccess={() => setOpen(false)} />
        <DeleteContact id={contact.id} onSuccess={() => setOpen(false)} />
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
