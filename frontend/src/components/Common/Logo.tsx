import { Link } from "@tanstack/react-router"

import { cn } from "@/lib/utils"
import icon from "/assets/images/ender-icon.svg"

interface LogoProps {
  variant?: "full" | "icon" | "responsive"
  className?: string
  asLink?: boolean
}

function EnderWordmark({ className }: { className?: string }) {
  return (
    <span className={cn("flex items-baseline", className)}>
      <span className="text-[#2dd4a8] text-2xl font-bold relative -top-[0.15em]">
        :
      </span>
      <span className="font-serif font-bold text-2xl tracking-tight uppercase">
        Ender
      </span>
      <span className="text-[#2dd4a8] text-2xl font-bold relative -top-[0.15em]">
        :
      </span>
    </span>
  )
}

export function Logo({
  variant = "full",
  className,
  asLink = true,
}: LogoProps) {
  const content =
    variant === "icon" ? (
      <img src={icon} alt="Ender" className={cn("size-5", className)} />
    ) : variant === "responsive" ? (
      <>
        <EnderWordmark
          className={cn("group-data-[collapsible=icon]:hidden", className)}
        />
        <img
          src={icon}
          alt="Ender"
          className={cn(
            "size-5 hidden group-data-[collapsible=icon]:block",
            className,
          )}
        />
      </>
    ) : (
      <EnderWordmark className={className} />
    )

  if (!asLink) {
    return content
  }

  return <Link to="/">{content}</Link>
}
