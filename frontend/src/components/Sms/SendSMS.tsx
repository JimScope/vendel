import { zodResolver } from "@hookform/resolvers/zod"
import { Plus } from "lucide-react"
import { useEffect, useState } from "react"
import { Controller, useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { z } from "zod"
import { MultiSelect } from "@/components/Common/MultiSelect"
import {
  type SelectedTemplate,
  TemplateSelect,
} from "@/components/Templates/TemplateSelect"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Field, FieldError, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { LoadingButton } from "@/components/ui/loading-button"
import { TagInput } from "@/components/ui/tag-input"
import { Textarea } from "@/components/ui/textarea"
import { useContactGroupList } from "@/hooks/useContactGroupList"
import { useContactList } from "@/hooks/useContactList"
import { useDeviceList } from "@/hooks/useDeviceList"
import { useSendSMS, useSendSMSTemplate } from "@/hooks/useSMSMutations"
import type { Contact } from "@/types/collections"

const formSchema = z.object({
  recipients: z.array(z.e164().min(1, "Recipient is required")),
  from: z.array(z.string()).min(1, "Device is required"),
  body: z.string(),
  group_ids: z.array(z.string()).optional(),
})

type FormData = z.infer<typeof formSchema>

const RESERVED_VAR_REGEX = /\{\{(name|phone)\}\}/

const SendSMS = () => {
  const { t } = useTranslation()
  const [isOpen, setIsOpen] = useState(false)
  const [selectedTemplate, setSelectedTemplate] =
    useState<SelectedTemplate | null>(null)
  const [templateVars, setTemplateVars] = useState<Record<string, string>>({})
  const { data: devices } = useDeviceList()
  const { data: contactGroups } = useContactGroupList()
  const { data: contacts } = useContactList()

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    mode: "onBlur",
    defaultValues: {
      recipients: [],
      from: [],
      body: "",
      group_ids: [],
    },
  })

  useEffect(() => {
    if (devices?.data?.length === 1 && form.getValues("from").length === 0) {
      form.setValue("from", [devices.data[0].id])
    }
  }, [devices, form])

  const sendSMSMutation = useSendSMS()
  const sendTemplateMutation = useSendSMSTemplate()
  const isSending = sendSMSMutation.isPending || sendTemplateMutation.isPending

  const onSubmit = (data: FormData) => {
    const hasRecipients =
      data.recipients.length > 0 ||
      (data.group_ids && data.group_ids.length > 0)
    if (!hasRecipients) {
      form.setError("recipients", {
        message: "At least one recipient or group required",
      })
      return
    }

    if (!selectedTemplate && !data.body) {
      form.setError("body", { message: "Message body or template required" })
      return
    }

    if (selectedTemplate) {
      const missing = selectedTemplate.customVariables.filter(
        (v) => !templateVars[v]?.trim(),
      )
      if (missing.length > 0) {
        form.setError("body", {
          message: `Missing variables: ${missing.join(", ")}`,
        })
        return
      }
    }

    const onSuccess = () => {
      form.reset()
      setSelectedTemplate(null)
      setTemplateVars({})
      setIsOpen(false)
    }

    if (selectedTemplate) {
      sendTemplateMutation.mutate(
        {
          recipients: data.recipients,
          template_id: selectedTemplate.id,
          variables: templateVars,
          device_id: data.from[0],
          group_ids: data.group_ids,
        },
        { onSuccess },
      )
    } else {
      sendSMSMutation.mutate(
        {
          recipients: data.recipients,
          body: data.body,
          device_id: data.from[0],
          group_ids: data.group_ids,
        },
        { onSuccess },
      )
    }
  }

  const handleOpenChange = (open: boolean) => {
    setIsOpen(open)
    if (!open) {
      setSelectedTemplate(null)
      setTemplateVars({})
    }
  }

  const hasReservedVars = selectedTemplate
    ? RESERVED_VAR_REGEX.test(selectedTemplate.body)
    : false

  return (
    <Dialog open={isOpen} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button className="my-4">
          <Plus className="h-4 w-4" />
          {t("sms.sendSms")}
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t("sms.sendSms")}</DialogTitle>
          <DialogDescription>{t("sms.sendSmsDesc")}</DialogDescription>
        </DialogHeader>

        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4 py-4">
          {/* Send to Group Field */}
          <Controller
            name="group_ids"
            control={form.control}
            render={({ field }) => (
              <Field>
                <FieldLabel htmlFor={field.name}>
                  {t("contacts.sendToGroup")}
                </FieldLabel>
                <MultiSelect
                  options={(contactGroups?.data || []).map((group) => ({
                    label: group.name,
                    value: group.id,
                  }))}
                  onValueChange={field.onChange}
                  defaultValue={field.value || []}
                  placeholder={t("contacts.groups")}
                />
              </Field>
            )}
          />

          {/* Recipient Field */}
          <Controller
            name="recipients"
            control={form.control}
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel htmlFor={field.name}>
                  {t("sms.to")} <span className="text-destructive">*</span>
                </FieldLabel>
                <TagInput
                  {...field}
                  id={field.name}
                  placeholder={t("sms.recipientPlaceholder")}
                  aria-invalid={fieldState.invalid}
                  suggestions={(
                    (contacts?.data || []) as unknown as Contact[]
                  ).map((c) => ({
                    label: c.name,
                    value: c.phone_number,
                  }))}
                />
                {fieldState.invalid && (
                  <FieldError errors={[fieldState.error]} />
                )}
              </Field>
            )}
          />

          {/* Device Selection Field */}
          <Controller
            name="from"
            control={form.control}
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel htmlFor={field.name}>
                  {t("sms.device")} <span className="text-destructive">*</span>
                </FieldLabel>
                <MultiSelect
                  options={(devices?.data || []).map((device) => ({
                    label: device.name || device.id,
                    value: device.id,
                  }))}
                  onValueChange={field.onChange}
                  defaultValue={field.value}
                />
                {fieldState.invalid && (
                  <FieldError errors={[fieldState.error]} />
                )}
              </Field>
            )}
          />

          {/* Template Select */}
          <TemplateSelect
            onSelect={(template) => {
              setSelectedTemplate(template)
              setTemplateVars({})
              form.setValue("body", "")
              form.clearErrors("body")
            }}
          />

          {selectedTemplate ? (
            <>
              {/* Template Preview */}
              <Field>
                <FieldLabel>{t("templates.templatePreview")}</FieldLabel>
                <div className="rounded-md border bg-muted/50 p-3 text-sm whitespace-pre-wrap">
                  {selectedTemplate.body}
                </div>
                {hasReservedVars && (
                  <p className="text-muted-foreground text-xs mt-1">
                    {t("templates.autoFilledHint")}
                  </p>
                )}
              </Field>

              {/* Custom Variable Inputs */}
              {selectedTemplate.customVariables.map((v) => (
                <Field key={v}>
                  <FieldLabel>
                    {v} <span className="text-destructive">*</span>
                  </FieldLabel>
                  <Input
                    value={templateVars[v] || ""}
                    onChange={(e) =>
                      setTemplateVars((prev) => ({
                        ...prev,
                        [v]: e.target.value,
                      }))
                    }
                    placeholder={t("templates.variableLabel", { var: v })}
                    required
                  />
                </Field>
              ))}
            </>
          ) : (
            /* Message Body Field */
            <Controller
              name="body"
              control={form.control}
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor={field.name}>
                    {t("sms.body")} <span className="text-destructive">*</span>
                  </FieldLabel>
                  <Textarea
                    {...field}
                    id={field.name}
                    placeholder={t("sms.bodyPlaceholder")}
                    rows={3}
                    aria-invalid={fieldState.invalid}
                  />
                  {fieldState.invalid && (
                    <FieldError errors={[fieldState.error]} />
                  )}
                </Field>
              )}
            />
          )}

          <DialogFooter>
            <DialogClose asChild>
              <Button variant="outline" type="button" disabled={isSending}>
                {t("common.cancel")}
              </Button>
            </DialogClose>
            <LoadingButton type="submit" loading={isSending}>
              {t("common.send")}
            </LoadingButton>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

export default SendSMS
