import { EllipsisVertical } from "lucide-react"
import { useState } from "react"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import type { ApiKey } from "@/types/collections"
import DeleteApiKey from "./DeleteApiKey"
import RevokeApiKey from "./RevokeApiKey"
import RotateApiKey from "./RotateApiKey"

interface ApiKeyActionsMenuProps {
  apiKey: ApiKey
}

export const ApiKeyActionsMenu = ({ apiKey }: ApiKeyActionsMenuProps) => {
  const [open, setOpen] = useState(false)

  return (
    <DropdownMenu open={open} onOpenChange={setOpen}>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon">
          <EllipsisVertical />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {apiKey.is_active && (
          <>
            <RotateApiKey id={apiKey.id} onSuccess={() => setOpen(false)} />
            <RevokeApiKey id={apiKey.id} onSuccess={() => setOpen(false)} />
            <DropdownMenuSeparator />
          </>
        )}
        <DeleteApiKey id={apiKey.id} onSuccess={() => setOpen(false)} />
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
