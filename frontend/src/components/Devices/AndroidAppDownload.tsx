import { Download, ExternalLink, Loader2, Smartphone } from "lucide-react"
import { useState } from "react"

import { Button } from "@/components/ui/button"
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
    <div className="space-y-2">
      <div className="flex items-center gap-2 text-sm font-medium">
        <Smartphone className="size-4 text-brand" />
        Android App
        <a
          href="https://github.com/JimScope/vendel-android/releases"
          target="_blank"
          rel="noopener noreferrer"
          className="ml-auto"
        >
          <Button variant="ghost" size="icon-sm">
            <ExternalLink className="size-3.5" />
          </Button>
        </a>
      </div>
      <div className="flex flex-wrap items-center gap-2">
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
    </div>
  )
}
