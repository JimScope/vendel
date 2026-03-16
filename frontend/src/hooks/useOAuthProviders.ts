import { useQuery } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"

interface OAuthProvider {
  name: string
  enabled: boolean
}

export function useOAuthProviders() {
  return useQuery({
    queryKey: ["oauth-providers"],
    queryFn: async (): Promise<OAuthProvider[]> => {
      const methods = await pb.collection("users").listAuthMethods()
      return (
        methods.oauth2?.providers?.map((p: { name: string }) => ({
          name: p.name,
          enabled: true,
        })) ?? []
      )
    },
    staleTime: 1000 * 60 * 5,
  })
}
