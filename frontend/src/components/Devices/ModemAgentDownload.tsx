import { Download, ExternalLink, Usb } from "lucide-react"
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

const PLATFORMS = [
  { value: "linux_amd64", label: "Linux x86_64", ext: ".tar.gz" },
  { value: "linux_arm64", label: "Linux ARM64", ext: ".tar.gz" },
  { value: "linux_arm", label: "Linux ARM", ext: ".tar.gz" },
  { value: "darwin_amd64", label: "macOS Intel", ext: ".tar.gz" },
  { value: "darwin_arm64", label: "macOS Apple Silicon", ext: ".tar.gz" },
  { value: "windows_amd64", label: "Windows x86_64", ext: ".zip" },
  { value: "windows_arm64", label: "Windows ARM64", ext: ".zip" },
] as const

function detectPlatform(): string {
  const ua = navigator.userAgent.toLowerCase()
  const isArm =
    /arm|aarch64/i.test(navigator.userAgent) ||
    // eslint-disable-next-line -- navigator.platform is the only reliable way to detect Apple Silicon
    ((navigator as { platform?: string }).platform === "MacIntel" &&
      navigator.maxTouchPoints > 1)

  if (ua.includes("mac")) return isArm ? "darwin_arm64" : "darwin_amd64"
  if (ua.includes("win")) return isArm ? "windows_arm64" : "windows_amd64"
  if (ua.includes("linux")) return isArm ? "linux_arm64" : "linux_amd64"
  return "linux_amd64"
}

function getDownloadUrl(platform: string): string {
  const p = PLATFORMS.find((p) => p.value === platform)
  const ext = p?.ext ?? ".tar.gz"
  return `https://github.com/JimScope/vendel/releases/latest/download/vendel-modem-agent_${platform}${ext}`
}

export default function ModemAgentDownload() {
  const [platform, setPlatform] = useState(detectPlatform)

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Usb className="size-4 text-brand" />
          <CardTitle className="text-sm">Modem Agent</CardTitle>
          <Badge variant="secondary" className="text-[10px] px-1.5 py-0">
            Latest
          </Badge>
        </div>
        <CardAction>
          <a
            href="https://github.com/JimScope/vendel/releases"
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
          <Select value={platform} onValueChange={setPlatform}>
            <SelectTrigger size="sm" className="min-w-[170px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {PLATFORMS.map((p) => (
                <SelectItem key={p.value} value={p.value}>
                  {p.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <a href={getDownloadUrl(platform)} download>
            <Button size="sm">
              <Download className="size-3.5" />
              Download
            </Button>
          </a>
        </div>
        <p className="text-muted-foreground text-xs mt-3">
          See the{" "}
          <a
            href="https://vendel.cc/docs/usb-modems"
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
