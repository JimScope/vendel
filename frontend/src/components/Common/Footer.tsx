import { Heart } from "lucide-react"
import { FaGithub } from "react-icons/fa"
import useAppConfig from "@/hooks/useAppConfig"

const socialLinks = [
  {
    icon: FaGithub,
    href: "https://github.com/JimScope/vendel",
    label: "GitHub",
  },
]

export function Footer() {
  const currentYear = new Date().getFullYear()
  const { config } = useAppConfig()

  return (
    <footer className="border-t py-4 px-6">
      <div className="flex flex-col items-center justify-between gap-4 sm:flex-row">
        <div className="flex flex-col items-center gap-1 sm:items-start">
          <p className="text-muted-foreground text-sm">Vendel - {currentYear}</p>
          <p className="flex items-center gap-1 text-xs text-muted-foreground">
            Hecho con <Heart className="size-3 fill-red-500 text-red-500" /> desde Cuba
          </p>
        </div>
        <div className="flex items-center gap-4">
          {config.supportEmail && config.supportEmail !== "support@example.com" && (
            <a
              href={`mailto:${config.supportEmail}`}
              className="text-muted-foreground hover:text-foreground transition-colors text-sm"
            >
              {config.supportEmail}
            </a>
          )}
          {socialLinks.map(({ icon: Icon, href, label }) => (
            <a
              key={label}
              href={href}
              target="_blank"
              rel="noopener noreferrer"
              aria-label={label}
              className="text-muted-foreground hover:text-foreground transition-colors"
            >
              <Icon className="h-5 w-5" />
            </a>
          ))}
        </div>
      </div>
    </footer>
  )
}
