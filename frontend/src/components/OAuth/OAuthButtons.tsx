import { FaGithub, FaGoogle } from "react-icons/fa"
import { Button } from "@/components/ui/button"
import useCustomToast from "@/hooks/useCustomToast"
import { useOAuthProviders } from "@/hooks/useOAuthProviders"
import pb from "@/lib/pocketbase"

interface OAuthButtonsProps {
  disabled?: boolean
}

export function OAuthButtons({ disabled }: OAuthButtonsProps) {
  const { showErrorToast } = useCustomToast()
  const { data: providers, isLoading } = useOAuthProviders()

  const handleOAuthLogin = async (providerName: string) => {
    try {
      await pb.collection("users").authWithOAuth2({ provider: providerName })
      window.location.href = "/"
    } catch (error: any) {
      showErrorToast(error?.message || "OAuth authentication failed")
    }
  }

  const enabledProviders = providers?.filter((p) => p.enabled) ?? []

  if (isLoading || enabledProviders.length === 0) {
    return null
  }

  return (
    <div className="grid gap-3">
      {enabledProviders.map((provider) => (
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
