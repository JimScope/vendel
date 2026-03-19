import { zodResolver } from "@hookform/resolvers/zod"
import { Plus } from "lucide-react"
import { useEffect, useState } from "react"
import { Controller, useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { z } from "zod"
import { MultiSelect } from "@/components/Common/MultiSelect"
import { TemplateSelect } from "@/components/Templates/TemplateSelect"
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
import { LoadingButton } from "@/components/ui/loading-button"
import { TagInput } from "@/components/ui/tag-input"
import { Textarea } from "@/components/ui/textarea"
import { useDeviceList } from "@/hooks/useDeviceList"
import { useSendSMS } from "@/hooks/useSMSMutations"

const formSchema = z.object({
  recipients: z
    .array(z.e164().min(1, "Recipient is required"))
    .min(1, "At least one recipient is required"),
  from: z.array(z.string()).min(1, "Device is required"),
  body: z.string().min(1, "Message body is required"),
})

type FormData = z.infer<typeof formSchema>

const SendSMS = () => {
  const { t } = useTranslation()
  const [isOpen, setIsOpen] = useState(false)
  const { data: devices } = useDeviceList()

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    mode: "onBlur",
    defaultValues: {
      recipients: [],
      from: [],
      body: "",
    },
  })

  useEffect(() => {
    if (devices?.data?.length === 1 && form.getValues("from").length === 0) {
      form.setValue("from", [devices.data[0].id])
    }
  }, [devices, form])

  const sendSMSMutation = useSendSMS()

  const onSubmit = (data: FormData) => {
    sendSMSMutation.mutate(
      {
        recipients: data.recipients,
        body: data.body,
        device_id: data.from[0],
      },
      {
        onSuccess: () => {
          form.reset()
          setIsOpen(false)
        },
      },
    )
  }

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
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
          <TemplateSelect onSelect={(body) => form.setValue("body", body)} />

          {/* Message Body Field */}
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

          <DialogFooter>
            <DialogClose asChild>
              <Button
                variant="outline"
                type="button"
                disabled={sendSMSMutation.isPending}
              >
                {t("common.cancel")}
              </Button>
            </DialogClose>
            <LoadingButton type="submit" loading={sendSMSMutation.isPending}>
              {t("common.send")}
            </LoadingButton>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

export default SendSMS
