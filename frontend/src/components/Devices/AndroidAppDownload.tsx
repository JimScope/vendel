import { Download, ExternalLink, Loader2, Smartphone } from "lucide-react"
import { useState } from "react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useLatestApks } from "@/hooks/useLatestApks"

const ABIS = [
  { value: "arm64-v8a", label: "ARM64 (most devices)" },
  { value: "armeabi-v7a", label: "ARM 32-bit" },
  { value: "x86_64", label: "x86_64 (emulators)" },
  { value: "universal", label: "Universal" },
] as const

export default function AndroidAppDownload() {
  const [abi, setAbi] = useState("arm64-v8a")
  const { data: apks, isLoading } = useLatestApks()
  const downloadUrl = apks?.get(abi)

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Smartphone className="size-4 text-brand" />
          <CardTitle className="text-sm">Android App</CardTitle>
          <Badge variant="secondary" className="text-[10px] px-1.5 py-0">
            Latest
          </Badge>
        </div>
        <CardAction>
          <a
            href="https://github.com/JimScope/vendel-android/releases"
            target="_blank"
            rel="noopener noreferrer"
          >
            <Button variant="ghost" size="icon-sm">
              <ExternalLink className="size-3.5" />
            </Button>
          </a>
        </CardAction>
      </CardHeader>
      <CardContent>
        <div className="flex items-center gap-2">
          <Select value={abi} onValueChange={setAbi}>
            <SelectTrigger size="sm" className="min-w-[170px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {ABIS.map((a) => (
                <SelectItem key={a.value} value={a.value}>
                  {a.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <a
            href={downloadUrl ?? "#"}
            download
            aria-disabled={!downloadUrl}
            className={!downloadUrl ? "pointer-events-none" : undefined}
          >
            <Button size="sm" disabled={isLoading || !downloadUrl}>
              {isLoading ? (
                <Loader2 className="size-3.5 animate-spin" />
              ) : (
                <Download className="size-3.5" />
              )}
              Download APK
            </Button>
          </a>
        </div>
        <p className="text-muted-foreground text-xs mt-3">
          See the{" "}
          <a
            href="https://vendel.cc/docs/android"
            target="_blank"
            rel="noopener noreferrer"
            className="text-brand hover:underline"
          >
            setup guide
          </a>{" "}
          for configuration instructions.
        </p>
      </CardContent>
    </Card>
  )
}
