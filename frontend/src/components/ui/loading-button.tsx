import * as React from "react"
import { Button, type buttonVariants } from "@/components/ui/button"
import { Spinner } from "@/components/ui/spinner"
import type { VariantProps } from "class-variance-authority"

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
  VariantProps<typeof buttonVariants> {
  asChild?: boolean
  loading?: boolean
}

function LoadingButton({
  loading = false,
  children,
  disabled,
  ...props
}: ButtonProps) {
  return (
    <Button disabled={loading || disabled} {...props}>
      {loading && <Spinner data-icon="inline-start" />}
      {children}
    </Button>
  )
}

export { LoadingButton }
