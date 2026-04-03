import { Megaphone, X } from "lucide-react"
import { useState } from "react"
import useAppConfig from "@/hooks/useAppConfig"

function hashText(text: string): string {
  let hash = 0
  for (let i = 0; i < text.length; i++) {
    const char = text.charCodeAt(i)
    hash = (hash << 5) - hash + char
    hash |= 0
  }
  return hash.toString(36)
}

function AnnouncementBanner() {
  const { config } = useAppConfig()
  const dismissKey = `banner-dismissed-${hashText(config.bannerText)}`
  const [dismissed, setDismissed] = useState(
    () => sessionStorage.getItem(dismissKey) === "true",
  )

  if (!config.bannerText || dismissed) return null

  const handleDismiss = () => {
    sessionStorage.setItem(dismissKey, "true")
    setDismissed(true)
  }

  const content = config.bannerUrl ? (
    <a
      href={config.bannerUrl}
      target="_blank"
      rel="noopener noreferrer"
      className="underline underline-offset-2"
    >
      {config.bannerText}
    </a>
  ) : (
    <span>{config.bannerText}</span>
  )

  return (
    <div className="flex w-full items-center gap-2 bg-brand px-4 py-2 text-sm text-neutral-900 dark:text-neutral-900">
      <Megaphone className="size-4 shrink-0" />
      <p className="min-w-0 flex-1 truncate">{content}</p>
      <button
        type="button"
        onClick={handleDismiss}
        className="shrink-0 rounded-sm p-0.5 hover:bg-black/10"
        aria-label="Dismiss"
      >
        <X className="size-4" />
      </button>
    </div>
  )
}

export default AnnouncementBanner
