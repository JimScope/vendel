import type { ColumnDef } from "@tanstack/react-table"
import type { TFunction } from "i18next"
import { Check, Copy } from "lucide-react"
import { useCopyToClipboard } from "@/hooks/useCopyToClipboard"
import type { SMSMessage } from "@/types/collections"
import { Button } from "../ui/button"
import { SMSActionsMenu } from "./SMSActionsMenu"

function CopyId({ id, t }: { id: string; t: TFunction }) {
  const [copiedText, copy] = useCopyToClipboard()
  const isCopied = copiedText === id

  return (
    <div className="flex items-center gap-1.5 group">
      <span className="font-mono text-xs text-muted-foreground">{id}</span>
      <Button
        variant="ghost"
        size="icon"
        className="size-6 opacity-0 group-hover:opacity-100 transition-opacity"
        onClick={() => copy(id)}
      >
        {isCopied ? (
          <Check className="size-3 text-green-500" />
        ) : (
          <Copy className="size-3" />
        )}
        <span className="sr-only">{t("sms.copyId")}</span>
      </Button>
    </div>
  )
}

export function getColumns(t: TFunction): ColumnDef<SMSMessage>[] {
  return [
    {
      accessorKey: "id",
      header: t("sms.id"),
      cell: ({ row }) => <CopyId id={row.original.id} t={t} />,
    },
    {
      accessorKey: "device",
      header: t("sms.deviceId"),
      cell: ({ row }) => (
        <span className="font-medium">{row.original.device}</span>
      ),
    },
    {
      accessorKey: "to",
      header: t("sms.to"),
      cell: ({ row }) => <span className="font-medium">{row.original.to}</span>,
    },
    {
      accessorKey: "message_type",
      header: t("sms.messageType"),
      cell: ({ row }) => (
        <span className="font-medium">{row.original.message_type}</span>
      ),
    },
    {
      accessorKey: "status",
      header: t("common.status"),
      cell: ({ row }) => (
        <span className="font-medium">{row.original.status}</span>
      ),
    },
    {
      id: "actions",
      header: () => <span className="sr-only">{t("common.actions")}</span>,
      cell: ({ row }) => (
        <div className="flex justify-end">
          <SMSActionsMenu sms={row.original} />
        </div>
      ),
    },
  ]
}
