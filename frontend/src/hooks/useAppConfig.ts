import { useQuery } from "@tanstack/react-query"

import pb from "@/lib/pocketbase"

export interface PaymentProvider {
  name: string
  display_name: string
}

export interface AppConfig {
  appName: string
  supportEmail: string
  maintenanceMode: boolean
  paymentProviders: PaymentProvider[]
  bannerText: string
  bannerUrl: string
}

const DEFAULT_CONFIG: AppConfig = {
  appName: "Vendel",
  supportEmail: "support@example.com",
  maintenanceMode: false,
  paymentProviders: [],
  bannerText: "",
  bannerUrl: "",
}

export function useAppConfig() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["app-settings"],
    queryFn: async () => {
      return (await pb.send("/api/utils/app-settings", {})) as Record<
        string,
        any
      >
    },
    staleTime: 1000 * 60 * 60, // Cache for 1 hour
    retry: 1,
  })

  const config: AppConfig = {
    appName: data?.app_name ?? DEFAULT_CONFIG.appName,
    supportEmail: data?.support_email ?? DEFAULT_CONFIG.supportEmail,
    maintenanceMode: data?.maintenance_mode === "true",
    paymentProviders:
      data?.payment_providers ?? DEFAULT_CONFIG.paymentProviders,
    bannerText: data?.banner_text ?? DEFAULT_CONFIG.bannerText,
    bannerUrl: data?.banner_url ?? DEFAULT_CONFIG.bannerUrl,
  }

  return {
    config,
    isLoading,
    error,
  }
}

export default useAppConfig
