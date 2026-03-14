import { createFileRoute, redirect, useNavigate } from "@tanstack/react-router"
import { useActionState, useEffect } from "react"
import { AuthLayout } from "@/components/Common/AuthLayout"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { LoadingButton } from "@/components/ui/loading-button"
import { isLoggedIn } from "@/hooks/useAuth"
import useCustomToast from "@/hooks/useCustomToast"
import pb from "@/lib/pocketbase"

export const Route = createFileRoute("/oauth-callback/$provider")({
  component: OAuthCallback,
  validateSearch: (search: Record<string, unknown>) => ({
    access_token: (search.access_token as string) || "",
    is_new_user: search.is_new_user === "true",
    requires_linking: search.requires_linking === "true",
    existing_email: (search.existing_email as string) || "",
    error: (search.error as string) || "",
  }),
  beforeLoad: async () => {
    if (isLoggedIn()) {
      throw redirect({
        to: "/",
      })
    }
  },
})

function OAuthCallback() {
  const { provider } = Route.useParams()
  const { access_token, is_new_user, requires_linking, existing_email, error } =
    Route.useSearch()
  const navigate = useNavigate()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  const [, linkAction, isLinking] = useActionState(
    async (_prev: string | null, formData: FormData) => {
      try {
        const pwd = formData.get("password") as string
        const response = await pb.send(`/api/oauth/${provider}/link`, {
          method: "POST",
          body: { email: existing_email, password: pwd },
        })
        if (response?.access_token) {
          showSuccessToast(`Your ${provider} account has been linked.`)
          navigate({ to: "/" })
        }
        return null
      } catch (err: any) {
        const msg = err.message || "Failed to link account"
        showErrorToast(msg)
        return msg
      }
    },
    null,
  )

  useEffect(() => {
    // Handle OAuth error
    if (error) {
      showErrorToast(error)
      navigate({ to: "/login" })
      return
    }

    // Handle successful OAuth (not requiring linking)
    if (access_token && !requires_linking) {
      if (is_new_user) {
        showSuccessToast("Welcome! Your account has been created.")
      } else {
        showSuccessToast("Welcome back!")
      }
      navigate({ to: "/" })
    }
  }, [
    access_token,
    error,
    is_new_user,
    requires_linking,
    navigate,
    showSuccessToast,
    showErrorToast,
  ])

  // Show linking form if required
  if (requires_linking && existing_email) {
    return (
      <AuthLayout>
        <div className="flex flex-col gap-6">
          <div className="flex flex-col items-center gap-2 text-center">
            <h1 className="text-2xl">Link Your Account</h1>
            <p className="text-muted-foreground text-sm">
              An account with email <strong>{existing_email}</strong> already
              exists. Enter your password to link your {provider} account.
            </p>
          </div>

          <form action={linkAction} className="grid gap-4">
            <div className="grid gap-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                name="password"
                type="password"
                placeholder="Enter your password"
                required
                minLength={8}
              />
            </div>

            <LoadingButton type="submit" loading={isLinking}>
              Link Account
            </LoadingButton>

            <Button
              type="button"
              variant="outline"
              onClick={() => navigate({ to: "/login" })}
            >
              Cancel
            </Button>
          </form>
        </div>
      </AuthLayout>
    )
  }

  // Show loading state
  return (
    <AuthLayout>
      <div className="flex flex-col items-center gap-4 text-center">
        <h1 className="text-2xl">Signing in...</h1>
        <p className="text-muted-foreground">
          Please wait while we complete your sign in with {provider}.
        </p>
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    </AuthLayout>
  )
}
