import { ChevronDown, ChevronRight, FileText } from "lucide-react"
import { useState } from "react"

import { Badge } from "@/components/ui/badge"
import { DropdownMenuItem } from "@/components/ui/dropdown-menu"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import { useWebhookLogs } from "@/hooks/useWebhookLogs"
import { cn } from "@/lib/utils"

interface WebhookLogsProps {
  webhook: Record<string, any>
  onSuccess: () => void
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleString()
}

function LogEntry({ log }: { log: Record<string, any> }) {
  const [expanded, setExpanded] = useState(false)
  const isSuccess = log.delivery_status === "success"

  return (
    <div className="border-b last:border-b-0">
      <button
        type="button"
        onClick={() => setExpanded(!expanded)}
        className="flex w-full items-center gap-3 px-4 py-3 text-left hover:bg-muted/50 transition-colors"
      >
        {expanded ? (
          <ChevronDown className="size-4 shrink-0 text-muted-foreground" />
        ) : (
          <ChevronRight className="size-4 shrink-0 text-muted-foreground" />
        )}
        <div className="flex flex-1 flex-col gap-1 min-w-0">
          <div className="flex items-center gap-2">
            <Badge
              variant={isSuccess ? "default" : "destructive"}
              className="text-[10px]"
            >
              {isSuccess ? "OK" : "FAIL"}
            </Badge>
            <span className="text-sm font-medium">{log.event}</span>
            {log.response_status > 0 && (
              <span className="text-xs text-muted-foreground">
                HTTP {log.response_status}
              </span>
            )}
            <span className="text-xs text-muted-foreground ml-auto">
              {log.duration_ms}ms
            </span>
          </div>
          <span className="text-xs text-muted-foreground truncate">
            {formatDate(log.created)}
          </span>
        </div>
      </button>
      {expanded && (
        <div className="px-4 pb-3 space-y-2">
          {log.error_message && (
            <div>
              <p className="text-xs font-medium text-destructive">Error</p>
              <p className="text-xs text-muted-foreground">
                {log.error_message}
              </p>
            </div>
          )}
          <div>
            <p className="text-xs font-medium mb-1">Request Body</p>
            <pre className="text-xs bg-muted rounded p-2 overflow-x-auto max-h-40">
              {typeof log.request_body === "string"
                ? log.request_body
                : JSON.stringify(log.request_body, null, 2)}
            </pre>
          </div>
          {log.response_body && (
            <div>
              <p className="text-xs font-medium mb-1">Response Body</p>
              <pre className="text-xs bg-muted rounded p-2 overflow-x-auto max-h-40">
                {log.response_body}
              </pre>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

const WebhookLogs = ({ webhook, onSuccess }: WebhookLogsProps) => {
  const [isOpen, setIsOpen] = useState(false)
  const { data, isLoading } = useWebhookLogs(webhook.id)

  return (
    <>
      <DropdownMenuItem
        onSelect={(e) => e.preventDefault()}
        onClick={() => {
          setIsOpen(true)
          onSuccess()
        }}
      >
        <FileText />
        View Logs
      </DropdownMenuItem>
      <Sheet open={isOpen} onOpenChange={setIsOpen}>
        <SheetContent className={cn("sm:max-w-lg overflow-y-auto")}>
          <SheetHeader>
            <SheetTitle>Delivery Logs</SheetTitle>
            <SheetDescription>
              Recent webhook deliveries to {webhook.url}
            </SheetDescription>
          </SheetHeader>
          <div className="flex-1 overflow-y-auto">
            {isLoading ? (
              <div className="flex items-center justify-center py-8 text-sm text-muted-foreground">
                Loading...
              </div>
            ) : !data?.data?.length ? (
              <div className="flex flex-col items-center justify-center py-8 text-sm text-muted-foreground gap-2">
                <FileText className="size-8 opacity-50" />
                <p>No delivery logs yet</p>
                <p className="text-xs">
                  Send a test or wait for an event to trigger
                </p>
              </div>
            ) : (
              <div className="border rounded-md">
                {data.data.map((log) => (
                  <LogEntry key={log.id} log={log} />
                ))}
              </div>
            )}
          </div>
        </SheetContent>
      </Sheet>
    </>
  )
}

export default WebhookLogs
