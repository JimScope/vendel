import { useQuery } from "@tanstack/react-query"
import { FaGithub, FaGoogle } from "react-icons/fa"

import pb from "@/lib/pocketbase"
import { Button } from "@/components/ui/button"
import useCustomToast from "@/hooks/useCustomToast"

interface OAuthButtonsProps {
  disabled?: boolean
}

export function OAuthButtons({ disabled }: OAuthButtonsProps) {
  const { showErrorToast } = useCustomToast()
  const { data: providers, isLoading } = useQuery({
    queryKey: ["oauth-providers"],
    queryFn: async () => {
      const methods = await pb.collection("users").listAuthMethods()
      return { providers: methods.oauth2?.providers?.map((p: any) => ({ name: p.name, enabled: true })) ?? [] }
    },
    staleTime: 1000 * 60 * 5, // Cache for 5 minutes
  })

  const handleOAuthLogin = async (providerName: string) => {
    try {
      await pb.collection("users").authWithOAuth2({ provider: providerName })
      window.location.href = "/"
    } catch (error: any) {
      showErrorToast(error?.message || "OAuth authentication failed")
    }
  }

  const enabledProviders =
    providers?.providers?.filter((p: Record<string, any>) => p.enabled) ?? []

  if (isLoading || enabledProviders.length === 0) {
    return null
  }

  return (
    <div className="grid gap-3">
      {enabledProviders.map((provider: Record<string, any>) => (
        <Button
          key={provider.name}
          variant="outline"
          type="button"
          disabled={disabled}
          onClick={() => handleOAuthLogin(provider.name)}
          className="w-full"
        >
          {provider.name === "google" && <FaGoogle className="mr-2 h-4 w-4" />}
          {provider.name === "github" && <FaGithub className="mr-2 h-4 w-4" />}
          Continue with{" "}
          {provider.name.charAt(0).toUpperCase() + provider.name.slice(1)}
        </Button>
      ))}
    </div>
  )
}
