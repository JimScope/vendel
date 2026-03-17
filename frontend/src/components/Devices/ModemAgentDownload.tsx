import { Download, ExternalLink, Usb } from "lucide-react"
import { useState } from "react"

import { Button } from "@/components/ui/button"
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
  const uaData = (navigator as { userAgentData?: { architecture?: string } })
    .userAgentData
  const arch = uaData?.architecture?.toLowerCase() ?? ""
  const isArm =
    arch === "arm" || /arm|aarch64/i.test(navigator.userAgent)

  if (ua.includes("mac")) {
    // Chromium exposes architecture via userAgentData
    // Safari does not — default to Apple Silicon (most modern Macs)
    if (arch === "x86") return "darwin_amd64"
    return "darwin_arm64"
  }
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
    <div className="space-y-2">
      <div className="flex items-center gap-2 text-sm font-medium">
        <Usb className="size-4 text-brand" />
        Modem Agent
        <a
          href="https://github.com/JimScope/vendel/releases"
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
    </div>
  )
}
