import { useQuery } from "@tanstack/react-query"

import pb from "@/lib/pocketbase"

export interface AppConfig {
  appName: string
  supportEmail: string
}

const DEFAULT_CONFIG: AppConfig = {
  appName: "Ender",
  supportEmail: "support@example.com",
}

export function useAppConfig() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["app-settings"],
    queryFn: async () => {
      return (await pb.send("/api/utils/app-settings", {})) as Record<
        string,
        string
      >
    },
    staleTime: 1000 * 60 * 60, // Cache for 1 hour
    retry: 1,
  })

  const config: AppConfig = {
    appName: data?.app_name ?? DEFAULT_CONFIG.appName,
    supportEmail: data?.support_email ?? DEFAULT_CONFIG.supportEmail,
  }

  return {
    config,
    isLoading,
    error,
  }
}

export default useAppConfig
