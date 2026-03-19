import { CreditCard, Loader2, Settings2 } from "lucide-react"
import { useTranslation } from "react-i18next"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { useSystemConfig, useUpdateSystemConfig } from "@/hooks/useSystemConfig"
import type { SystemConfig } from "@/types/collections"

function SystemSettings() {
  const { t } = useTranslation()
  const { data: configs, isLoading } = useSystemConfig()
  const updateConfigMutation = useUpdateSystemConfig()

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  const getConfigValue = (key: string): string => {
    const config = configs?.data?.find((c: SystemConfig) => c.key === key)
    return config?.value ?? ""
  }

  const handleConfigChange = (key: string, value: string) => {
    updateConfigMutation.mutate({ key, value })
  }

  const handleBooleanChange = (key: string) => (checked: boolean) => {
    updateConfigMutation.mutate({ key, value: checked ? "true" : "false" })
  }

  return (
    <div className="space-y-6">
      {/* General Settings */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Settings2 className="h-5 w-5" />
            {t("admin.general")}
          </CardTitle>
          <CardDescription>{t("admin.generalDesc")}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-2">
            <Label htmlFor="app-name">{t("admin.appName")}</Label>
            <Input
              id="app-name"
              defaultValue={getConfigValue("app_name")}
              onBlur={(e) => handleConfigChange("app_name", e.target.value)}
              placeholder={t("admin.appNamePlaceholder")}
              className="w-[300px]"
            />
            <p className="text-sm text-muted-foreground">
              {t("admin.appNameDesc")}
            </p>
          </div>

          <div className="grid gap-2">
            <Label htmlFor="support-email">{t("admin.supportEmail")}</Label>
            <Input
              id="support-email"
              type="email"
              defaultValue={getConfigValue("support_email")}
              onBlur={(e) =>
                handleConfigChange("support_email", e.target.value)
              }
              placeholder={t("admin.supportEmailPlaceholder")}
              className="w-[300px]"
            />
            <p className="text-sm text-muted-foreground">
              {t("admin.supportEmailDesc")}
            </p>
          </div>

          <div className="flex items-center justify-between rounded-lg border p-4">
            <div className="space-y-0.5">
              <Label htmlFor="maintenance-mode">
                {t("maintenance.maintenanceMode")}
              </Label>
              <p className="text-sm text-muted-foreground">
                {t("admin.maintenanceDesc")}
              </p>
            </div>
            <Switch
              id="maintenance-mode"
              checked={getConfigValue("maintenance_mode") === "true"}
              onCheckedChange={handleBooleanChange("maintenance_mode")}
              disabled={updateConfigMutation.isPending}
            />
          </div>
        </CardContent>
      </Card>

      {/* Payment Settings */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <CreditCard className="h-5 w-5" />
            {t("admin.payments")}
          </CardTitle>
          <CardDescription>{t("admin.paymentsDesc")}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-2">
            <Label htmlFor="payment-method">
              {t("admin.defaultPaymentMethod")}
            </Label>
            <Select
              value={getConfigValue("default_payment_method")}
              onValueChange={(value) =>
                handleConfigChange("default_payment_method", value)
              }
              disabled={updateConfigMutation.isPending}
            >
              <SelectTrigger id="payment-method" className="w-[300px]">
                <SelectValue placeholder={t("admin.selectPaymentMethod")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="invoice">
                  <div className="flex flex-col items-start">
                    <span>{t("admin.payPerPeriod")}</span>
                    <span className="text-xs text-muted-foreground">
                      {t("admin.payPerPeriodDesc")}
                    </span>
                  </div>
                </SelectItem>
                <SelectItem value="authorized">
                  <div className="flex flex-col items-start">
                    <span>{t("admin.autoRenew")}</span>
                    <span className="text-xs text-muted-foreground">
                      {t("admin.autoRenewDesc")}
                    </span>
                  </div>
                </SelectItem>
              </SelectContent>
            </Select>
            <p className="text-sm text-muted-foreground">
              {t("admin.paymentMethodDesc")}
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

export default SystemSettings
