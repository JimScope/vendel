import { createFileRoute, Link as RouterLink } from "@tanstack/react-router"
import { CheckCircle, Loader2, XCircle } from "lucide-react"
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
        errorMessage: "No verification token provided",
      }
    }
    try {
      await pb.collection("users").confirmVerification(token)
      return { status: "success" as const, errorMessage: "" }
    } catch (error: any) {
      return {
        status: "error" as const,
        errorMessage:
          error?.response?.data?.detail ||
          "Failed to verify email. The link may be invalid or expired.",
      }
    }
  },
})

function VerifyEmailPending() {
  return (
    <AuthLayout>
      <div className="flex flex-col items-center gap-6 text-center">
        <Loader2 className="h-16 w-16 animate-spin text-primary" />
        <h1 className="text-2xl">Verifying your email...</h1>
        <p className="text-muted-foreground">
          Please wait while we verify your email address.
        </p>
      </div>
    </AuthLayout>
  )
}

function VerifyEmail() {
  const { status, errorMessage } = Route.useLoaderData()

  return (
    <AuthLayout>
      <div className="flex flex-col items-center gap-6 text-center">
        {status === "success" && (
          <>
            <CheckCircle className="h-16 w-16 text-green-500" />
            <h1 className="text-2xl">Email Verified!</h1>
            <p className="text-muted-foreground">
              Your email has been successfully verified. You can now log in to
              your account.
            </p>
            <Button asChild className="mt-4">
              <RouterLink to="/login">Go to Login</RouterLink>
            </Button>
          </>
        )}

        {status === "error" && (
          <>
            <XCircle className="h-16 w-16 text-destructive" />
            <h1 className="text-2xl">Verification Failed</h1>
            <p className="text-muted-foreground">{errorMessage}</p>
            <div className="flex gap-4 mt-4">
              <Button variant="outline" asChild>
                <RouterLink to="/login">Go to Login</RouterLink>
              </Button>
              <Button asChild>
                <RouterLink to="/signup">Sign Up Again</RouterLink>
              </Button>
            </div>
          </>
        )}
      </div>
    </AuthLayout>
  )
}

export default VerifyEmail
