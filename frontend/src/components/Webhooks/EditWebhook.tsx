import { zodResolver } from "@hookform/resolvers/zod"
import { ChevronDown, Pencil } from "lucide-react"
import { useState } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { z } from "zod"

import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { DropdownMenuItem } from "@/components/ui/dropdown-menu"
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import { LoadingButton } from "@/components/ui/loading-button"
import { useUpdateWebhook } from "@/hooks/useWebhookMutations"
import {
  WEBHOOK_EVENT_DESCRIPTIONS,
  WEBHOOK_EVENT_LABELS,
  WEBHOOK_EVENT_PAYLOADS,
  WEBHOOK_EVENTS,
  type WebhookEvent,
} from "@/lib/webhook-events"
import type { WebhookConfig } from "@/types/collections"

const formSchema = z.object({
  url: z
    .string()
    .min(1, { message: "URL is required" })
    .url({ message: "Must be a valid URL" }),
  secret_key: z.string().optional().nullable(),
  events: z
    .array(z.enum(WEBHOOK_EVENTS))
    .min(1, { message: "Select at least one event" }),
  active: z.boolean().optional().nullable(),
})

type FormData = z.infer<typeof formSchema>

interface EditWebhookProps {
  webhook: WebhookConfig
  onSuccess: () => void
}

const EditWebhook = ({ webhook, onSuccess }: EditWebhookProps) => {
  const { t } = useTranslation()
  const [isOpen, setIsOpen] = useState(false)
  const [previewEvent, setPreviewEvent] = useState<WebhookEvent | null>(null)

  const defaultEvents = (webhook.events ?? []).filter((e): e is WebhookEvent =>
    WEBHOOK_EVENTS.includes(e as WebhookEvent),
  )

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    mode: "onBlur",
    criteriaMode: "all",
    defaultValues: {
      url: webhook.url,
      secret_key: webhook.secret_key ?? "",
      events: defaultEvents.length > 0 ? defaultEvents : ["sms_received"],
      active: webhook.active ?? true,
    },
  })

  const updateWebhookMutation = useUpdateWebhook(webhook.id)

  const selectedEvents = form.watch("events")

  const onSubmit = (data: FormData) => {
    updateWebhookMutation.mutate(
      {
        url: data.url,
        secret_key: data.secret_key ?? undefined,
        events: data.events,
        active: data.active ?? undefined,
      },
      {
        onSuccess: () => {
          setIsOpen(false)
          onSuccess()
        },
      },
    )
  }

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DropdownMenuItem
        onSelect={(e) => e.preventDefault()}
        onClick={() => setIsOpen(true)}
      >
        <Pencil />
        {t("webhooks.editWebhook")}
      </DropdownMenuItem>
      <DialogContent className="sm:max-w-lg">
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)}>
            <DialogHeader>
              <DialogTitle>{t("webhooks.editWebhook")}</DialogTitle>
              <DialogDescription>{t("webhooks.editDesc")}</DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <FormField
                control={form.control}
                name="url"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>
                      {t("webhooks.url")}{" "}
                      <span className="text-destructive">*</span>
                    </FormLabel>
                    <FormControl>
                      <Input
                        placeholder={t("webhooks.urlPlaceholder")}
                        type="url"
                        {...field}
                        required
                      />
                    </FormControl>
                    <FormDescription>{t("webhooks.urlDesc")}</FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="secret_key"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("webhooks.secretKey")}</FormLabel>
                    <FormControl>
                      <Input
                        placeholder={t("webhooks.secretPlaceholder")}
                        type="text"
                        {...field}
                        value={field.value ?? ""}
                      />
                    </FormControl>
                    <FormDescription>
                      {t("webhooks.secretDesc")}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="events"
                render={() => (
                  <FormItem>
                    <FormLabel>
                      {t("webhooks.events")}{" "}
                      <span className="text-destructive">*</span>
                    </FormLabel>
                    <FormDescription>
                      {t("webhooks.eventsDesc")}
                    </FormDescription>
                    <div className="grid gap-2 pt-1">
                      {WEBHOOK_EVENTS.map((event) => (
                        <FormField
                          key={event}
                          control={form.control}
                          name="events"
                          render={({ field }) => (
                            <FormItem className="flex items-start gap-3 space-y-0">
                              <FormControl>
                                <Checkbox
                                  checked={field.value?.includes(event)}
                                  onCheckedChange={(checked) => {
                                    const current = field.value ?? []
                                    field.onChange(
                                      checked
                                        ? [...current, event]
                                        : current.filter((e) => e !== event),
                                    )
                                  }}
                                />
                              </FormControl>
                              <div className="grid gap-0.5 leading-none">
                                <FormLabel className="font-normal cursor-pointer">
                                  {t(WEBHOOK_EVENT_LABELS[event])}
                                </FormLabel>
                                <p className="text-xs text-muted-foreground">
                                  {t(WEBHOOK_EVENT_DESCRIPTIONS[event])}
                                </p>
                              </div>
                            </FormItem>
                          )}
                        />
                      ))}
                    </div>
                    <FormMessage />
                  </FormItem>
                )}
              />

              {selectedEvents.length > 0 && (
                <Collapsible>
                  <CollapsibleTrigger asChild>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      className="gap-1.5 px-0 text-muted-foreground hover:text-foreground"
                    >
                      <ChevronDown className="size-4" />
                      {t("webhooks.viewPayload")}
                    </Button>
                  </CollapsibleTrigger>
                  <CollapsibleContent>
                    <div className="space-y-2 pt-1">
                      {selectedEvents.length > 1 && (
                        <div className="flex gap-1 flex-wrap">
                          {selectedEvents.map((event) => (
                            <Button
                              key={event}
                              type="button"
                              variant={
                                previewEvent === event ? "default" : "outline"
                              }
                              size="sm"
                              className="h-6 text-xs"
                              onClick={() => setPreviewEvent(event)}
                            >
                              {t(WEBHOOK_EVENT_LABELS[event])}
                            </Button>
                          ))}
                        </div>
                      )}
                      <pre className="rounded-md bg-muted p-3 text-xs font-mono overflow-x-auto max-h-48">
                        {JSON.stringify(
                          WEBHOOK_EVENT_PAYLOADS[
                            previewEvent &&
                            selectedEvents.includes(previewEvent)
                              ? previewEvent
                              : selectedEvents[0]
                          ],
                          null,
                          2,
                        )}
                      </pre>
                    </div>
                  </CollapsibleContent>
                </Collapsible>
              )}

              <FormField
                control={form.control}
                name="active"
                render={({ field }) => (
                  <FormItem className="flex items-center gap-3 space-y-0">
                    <FormControl>
                      <Checkbox
                        checked={field.value ?? false}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                    <FormLabel className="font-normal">
                      {t("webhooks.statusActive")}
                    </FormLabel>
                  </FormItem>
                )}
              />
            </div>

            <DialogFooter>
              <DialogClose asChild>
                <Button
                  variant="outline"
                  disabled={updateWebhookMutation.isPending}
                >
                  {t("common.cancel")}
                </Button>
              </DialogClose>
              <LoadingButton
                type="submit"
                loading={updateWebhookMutation.isPending}
              >
                {t("common.save")}
              </LoadingButton>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}

export default EditWebhook
