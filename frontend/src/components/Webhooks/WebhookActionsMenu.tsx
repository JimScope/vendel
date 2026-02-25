import { EllipsisVertical } from "lucide-react"
import { useState } from "react"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import DeleteWebhook from "./DeleteWebhook"
import EditWebhook from "./EditWebhook"
import TestWebhook from "./TestWebhook"
import WebhookLogs from "./WebhookLogs"

interface WebhookActionsMenuProps {
  webhook: Record<string, any>
}

export const WebhookActionsMenu = ({ webhook }: WebhookActionsMenuProps) => {
  const [open, setOpen] = useState(false)

  return (
    <DropdownMenu open={open} onOpenChange={setOpen}>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon">
          <EllipsisVertical />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <TestWebhook webhookId={webhook.id} onSuccess={() => setOpen(false)} />
        <WebhookLogs webhook={webhook} onSuccess={() => setOpen(false)} />
        <DropdownMenuSeparator />
        <EditWebhook webhook={webhook} onSuccess={() => setOpen(false)} />
        <DeleteWebhook id={webhook.id} onSuccess={() => setOpen(false)} />
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
