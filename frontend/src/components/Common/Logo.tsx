import { Link } from "@tanstack/react-router"

import { cn } from "@/lib/utils"
import icon from "/assets/images/vendel-icon.svg"

interface LogoProps {
  variant?: "full" | "icon" | "responsive"
  className?: string
  asLink?: boolean
}

function VendelWordmark({ className }: { className?: string }) {
  return (
    <span className={cn("flex items-baseline", className)}>
      <span className="text-brand text-2xl font-bold relative -top-[0.15em]">
        :
      </span>
      <span className="font-serif font-bold text-2xl tracking-tight uppercase">
        Vendel
      </span>
      <span className="text-brand text-2xl font-bold relative -top-[0.15em]">
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
      <img src={icon} alt="Vendel" className={cn("size-5", className)} />
    ) : variant === "responsive" ? (
      <>
        <VendelWordmark
          className={cn("group-data-[collapsible=icon]:hidden", className)}
        />
        <img
          src={icon}
          alt="Vendel"
          className={cn(
            "size-5 hidden group-data-[collapsible=icon]:block",
            className,
          )}
        />
      </>
    ) : (
      <VendelWordmark className={className} />
    )

  if (!asLink) {
    return content
  }

  return <Link to="/">{content}</Link>
}
