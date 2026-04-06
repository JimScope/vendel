import { createFileRoute, Link as RouterLink } from "@tanstack/react-router"
import { CircleCheckBig, LoaderCircle, CircleX } from "lucide-react"
import { useTranslation } from "react-i18next"
import { AuthLayout } from "@/components/Common/AuthLayout"
import { Button } from "@/components/ui/button"
import pb from "@/lib/pocketbase"

export const Route = createFileRoute("/verify-email")({
  component: VerifyEmail,
  pendingComponent: VerifyEmailPending,
  pendingMs: 0,
  validateSearch: (search: Record<string, unknown>) => ({
    token: (search.token as string) || "",
  }),
  head: () => ({
    meta: [
      {
        title: "Verify Email",
      },
    ],
  }),
  loaderDeps: ({ search }) => ({ token: search.token }),
  loader: async ({ deps: { token } }) => {
    if (!token) {
      return {
        status: "error" as const,
        errorKey: "auth.noVerificationToken" as const,
      }
    }
    try {
      await pb.collection("users").confirmVerification(token)
      return { status: "success" as const, errorKey: "" as const }
    } catch (error: any) {
      return {
        status: "error" as const,
        errorMessage: error?.response?.data?.detail,
        errorKey: "auth.verificationFailedMsg" as const,
      }
    }
  },
})

function VerifyEmailPending() {
  const { t } = useTranslation()
  return (
    <AuthLayout>
      <div className="flex flex-col items-center gap-6 text-center">
        <LoaderCircle className="h-16 w-16 animate-spin text-primary" />
        <h1 className="text-2xl">{t("auth.verifyingEmail")}</h1>
        <p className="text-muted-foreground">{t("auth.pleaseWait")}</p>
      </div>
    </AuthLayout>
  )
}

function VerifyEmail() {
  const { t } = useTranslation()
  const loaderData = Route.useLoaderData()
  const { status, errorKey } = loaderData
  const errorMessage =
    "errorMessage" in loaderData ? loaderData.errorMessage : undefined

  return (
    <AuthLayout>
      <div className="flex flex-col items-center gap-6 text-center">
        {status === "success" && (
          <>
            <CircleCheckBig className="h-16 w-16 text-green-500" />
            <h1 className="text-2xl">{t("auth.emailVerified")}</h1>
            <p className="text-muted-foreground">
              {t("auth.emailVerifiedSuccess")}
            </p>
            <Button asChild className="mt-4">
              <RouterLink to="/login">{t("auth.goToLogin")}</RouterLink>
            </Button>
          </>
        )}

        {status === "error" && (
          <>
            <CircleX className="h-16 w-16 text-destructive" />
            <h1 className="text-2xl">{t("auth.verificationFailed")}</h1>
            <p className="text-muted-foreground">
              {errorMessage || t(errorKey)}
            </p>
            <div className="flex gap-4 mt-4">
              <Button variant="outline" asChild>
                <RouterLink to="/login">{t("auth.goToLogin")}</RouterLink>
              </Button>
              <Button asChild>
                <RouterLink to="/signup">{t("auth.signUpAgain")}</RouterLink>
              </Button>
            </div>
          </>
        )}
      </div>
    </AuthLayout>
  )
}

export default VerifyEmail
