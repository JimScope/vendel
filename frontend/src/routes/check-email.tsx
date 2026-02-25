import { createFileRoute, Link as RouterLink } from "@tanstack/react-router"
import { ExternalLink, Mail } from "lucide-react"
import { PiMicrosoftOutlookLogoDuotone } from "react-icons/pi"
import { SiGmail, SiProtonmail } from "react-icons/si"
import { AuthLayout } from "@/components/Common/AuthLayout"
import { Button } from "@/components/ui/button"

const EMAIL_PROVIDERS = [
  {
    name: "Gmail",
    icon: SiGmail,
    url: "https://mail.google.com",
    domains: ["gmail.com", "googlemail.com"],
    color: "#EA4335",
  },
  {
    name: "Outlook",
    icon: PiMicrosoftOutlookLogoDuotone,
    url: "https://outlook.live.com",
    domains: ["outlook.com", "hotmail.com", "live.com", "msn.com"],
    color: "#0078D4",
  },
  {
    name: "ProtonMail",
    icon: SiProtonmail,
    url: "https://mail.proton.me",
    domains: ["proton.me", "protonmail.com", "pm.me"],
    color: "#6D4AFF",
  },
]

function getEmailDomain(email: string): string {
  const parts = email.split("@")
  return parts.length === 2 ? parts[1].toLowerCase() : ""
}

export const Route = createFileRoute("/check-email")({
  component: CheckEmail,
  validateSearch: (search: Record<string, unknown>) => ({
    email: (search.email as string) || "",
  }),
  head: () => ({
    meta: [
      {
        title: "Check Your Email",
      },
    ],
  }),
})

function CheckEmail() {
  const { email } = Route.useSearch()
  const domain = getEmailDomain(email)
  const matchedProvider = EMAIL_PROVIDERS.find((p) =>
    p.domains.includes(domain),
  )

  return (
    <AuthLayout>
      <div className="flex flex-col items-center gap-6 text-center">
        <div className="rounded-full bg-brand/10 p-4">
          <Mail className="h-12 w-12 text-brand" />
        </div>

        <h1 className="text-2xl">Check your email</h1>

        <p className="text-muted-foreground">
          We've sent a verification link to{" "}
          {email ? (
            <span className="font-medium text-foreground">{email}</span>
          ) : (
            "your email address"
          )}
          . Click the link in the email to verify your account.
        </p>

        {matchedProvider ? (
          <a
            href={matchedProvider.url}
            target="_blank"
            rel="noopener noreferrer"
            className="group flex w-full items-center justify-center gap-3 rounded-lg border-2 border-border bg-card px-5 py-3.5 font-medium transition-all hover:border-[var(--provider-color)] hover:shadow-sm"
            style={
              {
                "--provider-color": matchedProvider.color,
              } as React.CSSProperties
            }
          >
            <matchedProvider.icon
              className="size-5"
              style={{ color: matchedProvider.color }}
            />
            Open {matchedProvider.name}
            <ExternalLink className="size-3.5 text-muted-foreground transition-transform group-hover:translate-x-0.5" />
          </a>
        ) : (
          <div className="flex w-full flex-col gap-2">
            {EMAIL_PROVIDERS.map((provider) => (
              <a
                key={provider.name}
                href={provider.url}
                target="_blank"
                rel="noopener noreferrer"
                className="group flex items-center justify-center gap-3 rounded-lg border border-border bg-card px-5 py-3 text-sm font-medium transition-all hover:border-[var(--provider-color)] hover:shadow-sm"
                style={
                  {
                    "--provider-color": provider.color,
                  } as React.CSSProperties
                }
              >
                <provider.icon
                  className="size-4"
                  style={{ color: provider.color }}
                />
                Open {provider.name}
                <ExternalLink className="size-3 text-muted-foreground transition-transform group-hover:translate-x-0.5" />
              </a>
            ))}
          </div>
        )}

        <div className="text-sm text-muted-foreground">
          <p>Didn't receive the email?</p>
          <p>Check your spam folder or try signing up again.</p>
        </div>

        <div className="flex gap-4">
          <Button variant="outline" asChild>
            <RouterLink to="/login">Go to Login</RouterLink>
          </Button>
          <Button variant="outline" asChild>
            <RouterLink to="/signup">Sign Up Again</RouterLink>
          </Button>
        </div>
      </div>
    </AuthLayout>
  )
}

export default CheckEmail
