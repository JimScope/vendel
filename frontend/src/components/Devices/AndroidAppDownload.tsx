import { Download, ExternalLink, LoaderCircle, Smartphone } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useLatestApks } from "@/hooks/useLatestApks"

const ABI_VALUES = ["arm64-v8a", "armeabi-v7a", "x86_64", "universal"] as const

const ABI_LABEL_KEYS = {
  "arm64-v8a": "devices.arm64",
  "armeabi-v7a": "devices.arm32",
  x86_64: "devices.x86_64",
  universal: "devices.universal",
} as const

export default function AndroidAppDownload() {
  const { t } = useTranslation()
  const [abi, setAbi] = useState("arm64-v8a")
  const { data: apks, isLoading } = useLatestApks()
  const downloadUrl = apks?.get(abi)

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2 text-sm font-medium">
        <Smartphone className="size-4 text-brand" />
        {t("devices.androidApp")}
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
            {ABI_VALUES.map((value) => (
              <SelectItem key={value} value={value}>
                {t(ABI_LABEL_KEYS[value])}
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
              <LoaderCircle className="size-3.5 animate-spin" />
            ) : (
              <Download className="size-3.5" />
            )}
            {t("devices.downloadApk")}
          </Button>
        </a>
      </div>
    </div>
  )
}
