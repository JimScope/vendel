import { Cuer } from "cuer"
import { Check, Copy, RefreshCw } from "lucide-react"
import { useState } from "react"

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
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { LoadingButton } from "@/components/ui/loading-button"
import { useRotateApiKey } from "@/hooks/useApiKeyMutations"
import { useCopyToClipboard } from "@/hooks/useCopyToClipboard"

interface RotateApiKeyProps {
  id: string
  onSuccess: () => void
}

const QR_PAYLOAD_VERSION = "0.1"

const RotateApiKey = ({ id, onSuccess }: RotateApiKeyProps) => {
  const [isOpen, setIsOpen] = useState(false)
  const [expiresAt, setExpiresAt] = useState("")
  const [newKey, setNewKey] = useState<string | null>(null)
  const [copiedText, copyToClipboard] = useCopyToClipboard()

  const rotateMutation = useRotateApiKey()

  const handleRotate = () => {
    rotateMutation.mutate(
      { id, expires_at: expiresAt || undefined },
      {
        onSuccess: (result) => {
          setNewKey(result.key)
        },
      },
    )
  }

  const handleClose = (open: boolean) => {
    if (!open) {
      setExpiresAt("")
      setNewKey(null)
      rotateMutation.reset()
      onSuccess()
    }
    setIsOpen(open)
  }

  const getQrPayload = (apiKey: string) => {
    const payload = {
      server_instance: import.meta.env.VITE_API_URL,
      api_key: apiKey,
      version: QR_PAYLOAD_VERSION,
    }
    return JSON.stringify(payload)
  }

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DropdownMenuItem
        onSelect={(e) => e.preventDefault()}
        onClick={() => setIsOpen(true)}
      >
        <RefreshCw />
        Rotate
      </DropdownMenuItem>
      <DialogContent className="sm:max-w-md">
        {newKey ? (
          <>
            <DialogHeader>
              <DialogTitle>Key Rotated</DialogTitle>
              <DialogDescription>
                Your old key has been revoked. Copy your new API key now — you
                won't be able to see it again.
              </DialogDescription>
            </DialogHeader>
            <div className="flex flex-col gap-4 py-4">
              <div className="flex flex-col items-center gap-3">
                <div className="rounded-lg border bg-white p-4">
                  <Cuer.Root value={getQrPayload(newKey)} size={200}>
                    <Cuer.Finder fill="black" />
                    <Cuer.Cells fill="black" />
                  </Cuer.Root>
                </div>
                <p className="text-sm text-muted-foreground text-center">
                  Scan this QR code with the app to connect automatically
                </p>
              </div>
              <div className="flex items-center gap-2">
                <Input value={newKey} readOnly className="font-mono text-sm" />
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() => copyToClipboard(newKey)}
                >
                  {copiedText ? (
                    <Check className="h-4 w-4" />
                  ) : (
                    <Copy className="h-4 w-4" />
                  )}
                </Button>
              </div>
            </div>
            <DialogFooter>
              <DialogClose asChild>
                <Button>Done</Button>
              </DialogClose>
            </DialogFooter>
          </>
        ) : (
          <>
            <DialogHeader>
              <DialogTitle>Rotate API Key</DialogTitle>
              <DialogDescription>
                This will revoke the current key and generate a new one. Any
                integrations using the old key will stop working.
              </DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="expires_at">Expiration (optional)</Label>
                <Input
                  id="expires_at"
                  type="date"
                  value={expiresAt}
                  onChange={(e) => setExpiresAt(e.target.value)}
                  min={new Date().toISOString().split("T")[0]}
                />
                <p className="text-xs text-muted-foreground">
                  Leave empty for a key that never expires
                </p>
              </div>
            </div>
            <DialogFooter>
              <DialogClose asChild>
                <Button variant="outline" disabled={rotateMutation.isPending}>
                  Cancel
                </Button>
              </DialogClose>
              <LoadingButton
                onClick={handleRotate}
                loading={rotateMutation.isPending}
              >
                Rotate Key
              </LoadingButton>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}

export default RotateApiKey
