import { createFileRoute, redirect, useNavigate } from "@tanstack/react-router"
import { useActionState, useEffect } from "react"
import { useTranslation } from "react-i18next"
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
  const { t } = useTranslation()
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
          showSuccessToast(t("toast.accountLinked", { provider }))
          navigate({ to: "/" })
        }
        return null
      } catch (err: any) {
        const msg = err.message || t("toast.oauthFailed")
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
        showSuccessToast(t("toast.authWelcome"))
      } else {
        showSuccessToast(t("toast.authWelcomeBack"))
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
    t,
  ])

  // Show linking form if required
  if (requires_linking && existing_email) {
    return (
      <AuthLayout>
        <div className="flex flex-col gap-6">
          <div className="flex flex-col items-center gap-2 text-center">
            <h1 className="text-2xl">{t("auth.linkAccount")}</h1>
            <p className="text-muted-foreground text-sm">
              {t("auth.linkAccountDesc", { email: existing_email, provider })}
            </p>
          </div>

          <form action={linkAction} className="grid gap-4">
            <div className="grid gap-2">
              <Label htmlFor="password">{t("common.password")}</Label>
              <Input
                id="password"
                name="password"
                type="password"
                placeholder={t("auth.enterPassword")}
                required
                minLength={8}
              />
            </div>

            <LoadingButton type="submit" loading={isLinking}>
              {t("auth.linkAccountButton")}
            </LoadingButton>

            <Button
              type="button"
              variant="outline"
              onClick={() => navigate({ to: "/login" })}
            >
              {t("common.cancel")}
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
        <h1 className="text-2xl">{t("auth.signingIn")}</h1>
        <p className="text-muted-foreground">
          {t("auth.signingInDesc", { provider })}
        </p>
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    </AuthLayout>
  )
}
