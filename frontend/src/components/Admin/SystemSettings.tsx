import {
  Banknote,
  Check,
  Coins,
  CreditCard,
  Loader2,
  Settings2,
} from "lucide-react"
import { useEffect, useState } from "react"
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
import { LoadingButton } from "@/components/ui/loading-button"
import { Switch } from "@/components/ui/switch"
import useCustomToast from "@/hooks/useCustomToast"
import { useSaveSystemConfig, useSystemConfig } from "@/hooks/useSystemConfig"
import type { SystemConfig } from "@/types/collections"

function EnvHint({ envSet }: { envSet: boolean }) {
  const { t } = useTranslation()
  if (!envSet) return null
  return (
    <span className="inline-flex items-center gap-1 rounded-full bg-brand/15 px-2 py-0.5 text-xs font-medium text-[var(--neutral-700)] dark:bg-brand/20 dark:text-brand">
      <Check className="size-3 text-brand" />
      {t("admin.envVarSet")}
    </span>
  )
}

function SystemSettings() {
  const { t } = useTranslation()
  const { data: configs, isLoading } = useSystemConfig()
  const saveMutation = useSaveSystemConfig()
  const { showSuccessToast, showErrorToast } = useCustomToast()

  // Local draft state — only sent to backend on Save
  const [draft, setDraft] = useState<Record<string, string>>({})

  // Sync draft from server data whenever it changes (initial load + after save refetch)
  useEffect(() => {
    if (!configs?.data) return
    const server: Record<string, string> = {}
    for (const c of configs.data as SystemConfig[]) {
      server[c.key] = c.value
    }
    setDraft(server)
  }, [configs?.data])

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  const get = (key: string) => draft[key] ?? ""
  const set = (key: string, value: string) =>
    setDraft((prev) => ({ ...prev, [key]: value }))
  const toggle = (key: string) => (checked: boolean) =>
    set(key, checked ? "true" : "false")

  const envHints: Record<string, boolean> = configs?.env_hints ?? {}
  const hasEnv = (key: string) => envHints[key] === true

  // Diff: only send keys that changed from server state
  const serverValues: Record<string, string> = {}
  for (const c of (configs?.data ?? []) as SystemConfig[]) {
    serverValues[c.key] = c.value
  }
  const changedKeys = Object.keys(draft).filter(
    (k) => draft[k] !== (serverValues[k] ?? ""),
  )
  const isDirty = changedKeys.length > 0

  const handleSave = () => {
    if (!isDirty) return
    const payload: Record<string, string> = {}
    for (const k of changedKeys) {
      payload[k] = draft[k]
    }
    saveMutation.mutate(payload, {
      onSuccess: () => showSuccessToast(t("admin.configSaved")),
      onError: () => showErrorToast(t("admin.configSaveFailed")),
    })
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
              value={get("app_name")}
              onChange={(e) => set("app_name", e.target.value)}
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
              value={get("support_email")}
              onChange={(e) => set("support_email", e.target.value)}
              placeholder={t("admin.supportEmailPlaceholder")}
              className="w-[300px]"
            />
            <p className="text-sm text-muted-foreground">
              {t("admin.supportEmailDesc")}
            </p>
          </div>

          <div className="grid gap-2">
            <Label htmlFor="sender-email">{t("admin.senderEmail")}</Label>
            <Input
              id="sender-email"
              type="email"
              value={get("sender_email")}
              onChange={(e) => set("sender_email", e.target.value)}
              placeholder={t("admin.senderEmailPlaceholder")}
              className="w-[300px]"
            />
            <p className="text-sm text-muted-foreground">
              {t("admin.senderEmailDesc")}
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
              checked={get("maintenance_mode") === "true"}
              onCheckedChange={toggle("maintenance_mode")}
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="banner-text">{t("admin.bannerText")}</Label>
            <Input
              id="banner-text"
              value={get("banner_text")}
              onChange={(e) => set("banner_text", e.target.value)}
              placeholder={t("admin.bannerTextPlaceholder")}
              className="w-full"
              maxLength={200}
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="banner-url">{t("admin.bannerUrl")}</Label>
            <Input
              id="banner-url"
              type="url"
              value={get("banner_url")}
              onChange={(e) => set("banner_url", e.target.value)}
              placeholder={t("admin.bannerUrlPlaceholder")}
              className="w-[300px]"
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
          <p className="text-xs text-muted-foreground">
            {t("admin.envOverrideHint")}
          </p>

          {/* TronDealer */}
          <div className="rounded-lg border">
            <div className="flex items-center justify-between p-4">
              <div className="flex items-center gap-3">
                <Coins className="h-5 w-5 text-muted-foreground" />
                <div className="space-y-0.5">
                  <Label htmlFor="provider-trondealer">TronDealer</Label>
                  <p className="text-sm text-muted-foreground">
                    {t("admin.cryptoPayments")}
                  </p>
                </div>
              </div>
              <Switch
                id="provider-trondealer"
                checked={get("provider_trondealer_enabled") === "true"}
                onCheckedChange={toggle("provider_trondealer_enabled")}
              />
            </div>
            {get("provider_trondealer_enabled") === "true" && (
              <div className="border-t px-4 py-3 space-y-3">
                <div className="grid gap-1.5">
                  <div className="flex items-center justify-between">
                    <Label htmlFor="td-api-key" className="text-xs">
                      API Key
                    </Label>
                    <EnvHint envSet={hasEnv("trondealer_api_key")} />
                  </div>
                  <Input
                    id="td-api-key"
                    type="password"
                    value={get("trondealer_api_key")}
                    onChange={(e) => set("trondealer_api_key", e.target.value)}
                    placeholder={
                      hasEnv("trondealer_api_key")
                        ? t("admin.envVarConfigured")
                        : "td_..."
                    }
                    className="h-8 text-xs"
                  />
                </div>
                <div className="grid gap-1.5">
                  <div className="flex items-center justify-between">
                    <Label htmlFor="td-webhook-secret" className="text-xs">
                      Webhook Secret
                    </Label>
                    <EnvHint envSet={hasEnv("trondealer_webhook_secret")} />
                  </div>
                  <Input
                    id="td-webhook-secret"
                    type="password"
                    value={get("trondealer_webhook_secret")}
                    onChange={(e) =>
                      set("trondealer_webhook_secret", e.target.value)
                    }
                    placeholder={
                      hasEnv("trondealer_webhook_secret")
                        ? t("admin.envVarConfigured")
                        : "Webhook secret"
                    }
                    className="h-8 text-xs"
                  />
                </div>
              </div>
            )}
          </div>

          {/* QvaPay */}
          <div className="rounded-lg border">
            <div className="flex items-center justify-between p-4">
              <div className="flex items-center gap-3">
                <Banknote className="h-5 w-5 text-muted-foreground" />
                <div className="space-y-0.5">
                  <Label htmlFor="provider-qvapay">QvaPay</Label>
                  <p className="text-sm text-muted-foreground">
                    {t("admin.qvapayPayments")}
                  </p>
                </div>
              </div>
              <Switch
                id="provider-qvapay"
                checked={get("provider_qvapay_enabled") === "true"}
                onCheckedChange={toggle("provider_qvapay_enabled")}
              />
            </div>
            {get("provider_qvapay_enabled") === "true" && (
              <div className="border-t px-4 py-3 space-y-3">
                <div className="grid gap-1.5">
                  <div className="flex items-center justify-between">
                    <Label htmlFor="qvapay-app-id" className="text-xs">
                      App ID
                    </Label>
                    <EnvHint envSet={hasEnv("qvapay_app_id")} />
                  </div>
                  <Input
                    id="qvapay-app-id"
                    value={get("qvapay_app_id")}
                    onChange={(e) => set("qvapay_app_id", e.target.value)}
                    placeholder={
                      hasEnv("qvapay_app_id")
                        ? t("admin.envVarConfigured")
                        : "QvaPay App ID"
                    }
                    className="h-8 text-xs"
                  />
                </div>
                <div className="grid gap-1.5">
                  <div className="flex items-center justify-between">
                    <Label htmlFor="qvapay-app-secret" className="text-xs">
                      App Secret
                    </Label>
                    <EnvHint envSet={hasEnv("qvapay_app_secret")} />
                  </div>
                  <Input
                    id="qvapay-app-secret"
                    type="password"
                    value={get("qvapay_app_secret")}
                    onChange={(e) => set("qvapay_app_secret", e.target.value)}
                    placeholder={
                      hasEnv("qvapay_app_secret")
                        ? t("admin.envVarConfigured")
                        : "QvaPay App Secret"
                    }
                    className="h-8 text-xs"
                  />
                </div>
              </div>
            )}
          </div>

          {/* Stripe */}
          <div className="rounded-lg border">
            <div className="flex items-center justify-between p-4">
              <div className="flex items-center gap-3">
                <CreditCard className="h-5 w-5 text-muted-foreground" />
                <div className="space-y-0.5">
                  <Label htmlFor="provider-stripe">Stripe</Label>
                  <p className="text-sm text-muted-foreground">
                    {t("admin.cardPayments")}
                  </p>
                </div>
              </div>
              <Switch
                id="provider-stripe"
                checked={get("provider_stripe_enabled") === "true"}
                onCheckedChange={toggle("provider_stripe_enabled")}
              />
            </div>
            {get("provider_stripe_enabled") === "true" && (
              <div className="border-t px-4 py-3 space-y-3">
                <div className="grid gap-1.5">
                  <div className="flex items-center justify-between">
                    <Label htmlFor="stripe-secret-key" className="text-xs">
                      Secret Key
                    </Label>
                    <EnvHint envSet={hasEnv("stripe_secret_key")} />
                  </div>
                  <Input
                    id="stripe-secret-key"
                    type="password"
                    value={get("stripe_secret_key")}
                    onChange={(e) => set("stripe_secret_key", e.target.value)}
                    placeholder={
                      hasEnv("stripe_secret_key")
                        ? t("admin.envVarConfigured")
                        : "sk_..."
                    }
                    className="h-8 text-xs"
                  />
                </div>
                <div className="grid gap-1.5">
                  <div className="flex items-center justify-between">
                    <Label htmlFor="stripe-webhook-secret" className="text-xs">
                      Webhook Secret
                    </Label>
                    <EnvHint envSet={hasEnv("stripe_webhook_secret")} />
                  </div>
                  <Input
                    id="stripe-webhook-secret"
                    type="password"
                    value={get("stripe_webhook_secret")}
                    onChange={(e) =>
                      set("stripe_webhook_secret", e.target.value)
                    }
                    placeholder={
                      hasEnv("stripe_webhook_secret")
                        ? t("admin.envVarConfigured")
                        : "whsec_..."
                    }
                    className="h-8 text-xs"
                  />
                </div>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Save button */}
      <div className="flex justify-end">
        <LoadingButton
          onClick={handleSave}
          loading={saveMutation.isPending}
          disabled={!isDirty}
        >
          {t("common.save")}
        </LoadingButton>
      </div>
    </div>
  )
}

export default SystemSettings
