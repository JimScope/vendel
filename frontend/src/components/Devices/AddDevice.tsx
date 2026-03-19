import { zodResolver } from "@hookform/resolvers/zod"
import { Cuer } from "cuer"
import { Check, Copy, Plus } from "lucide-react"
import { useState } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { z } from "zod"

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
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import { LoadingButton } from "@/components/ui/loading-button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import useAppConfig from "@/hooks/useAppConfig"
import { useCopyToClipboard } from "@/hooks/useCopyToClipboard"
import { useCreateDevice } from "@/hooks/useDeviceMutations"

const formSchema = z.object({
  device_type: z.enum(["android", "modem"]),
  name: z
    .string()
    .min(1, { message: "Name is required" })
    .max(255, { message: "Name must be at most 255 characters" }),
  phone_number: z
    .e164()
    .min(1, { message: "Phone number is required" })
    .max(20, { message: "Phone number must be at most 20 characters" }),
})

type FormData = z.infer<typeof formSchema>

const QR_PAYLOAD_VERSION = "0.1"

interface AddDeviceProps {
  open?: boolean
  onOpenChange?: (open: boolean) => void
}

const AddDevice = ({ open, onOpenChange }: AddDeviceProps) => {
  const { t } = useTranslation()
  const [internalOpen, setInternalOpen] = useState(false)
  const isOpen = open ?? internalOpen
  const setIsOpen = onOpenChange ?? setInternalOpen
  const [apiKey, setApiKey] = useState<string | null>(null)
  const [copiedText, copyToClipboard] = useCopyToClipboard()
  const { config } = useAppConfig()

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    mode: "onBlur",
    criteriaMode: "all",
    defaultValues: {
      device_type: "android",
      name: "",
      phone_number: "",
    },
  })

  const createDeviceMutation = useCreateDevice()

  const onSubmit = (data: FormData) => {
    createDeviceMutation.mutate(data, {
      onSuccess: (response) => {
        setApiKey(response?.api_key ?? null)
        form.reset()
      },
    })
  }

  const handleClose = (open: boolean) => {
    if (!open) {
      setApiKey(null)
      form.reset()
    }
    setIsOpen(open)
  }

  const getQrPayload = (deviceApiKey: string) => {
    const payload = {
      server_instance: import.meta.env.VITE_API_URL,
      api_key: deviceApiKey,
      version: QR_PAYLOAD_VERSION,
    }
    return JSON.stringify(payload)
  }

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogTrigger asChild>
        <Button className="my-4">
          <Plus />
          {t("devices.addDevice")}
        </Button>
      </DialogTrigger>
      <DialogContent
        className="sm:max-w-md"
        onInteractOutside={(e) => e.preventDefault()}
      >
        {apiKey ? (
          <>
            <DialogHeader>
              <DialogTitle>{t("devices.deviceCreated")}</DialogTitle>
              <DialogDescription>
                {t("devices.deviceCreatedMsg")}
              </DialogDescription>
            </DialogHeader>
            <div className="flex flex-col gap-4 py-4">
              <div className="flex flex-col items-center gap-3">
                <div
                  className="rounded-lg border bg-white p-4"
                  role="img"
                  aria-label="QR code containing device connection details"
                >
                  <Cuer.Root value={getQrPayload(apiKey)} size={200}>
                    <Cuer.Finder fill="black" />
                    <Cuer.Cells fill="black" />
                  </Cuer.Root>
                </div>
                <p className="text-sm text-muted-foreground text-center">
                  {t("devices.scanQrCode")} {config.appName}{" "}
                  {t("devices.toConnectAutomatically")}
                </p>
              </div>
              <div className="flex items-center gap-2">
                <Input value={apiKey} readOnly className="font-mono text-sm" />
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() => copyToClipboard(apiKey)}
                  aria-label={
                    copiedText ? t("common.copied") : t("devices.copyApiKey")
                  }
                >
                  {copiedText ? (
                    <Check className="h-4 w-4" />
                  ) : (
                    <Copy className="h-4 w-4" />
                  )}
                </Button>
              </div>
            </div>
            <DialogFooter>
              <Button onClick={() => handleClose(false)}>
                {t("common.done")}
              </Button>
            </DialogFooter>
          </>
        ) : (
          <>
            <DialogHeader>
              <DialogTitle>{t("devices.addDevice")}</DialogTitle>
              <DialogDescription>{t("devices.description")}</DialogDescription>
            </DialogHeader>
            <Form {...form}>
              <form onSubmit={form.handleSubmit(onSubmit)}>
                <div className="grid gap-4 py-4">
                  <FormField
                    control={form.control}
                    name="device_type"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>{t("devices.deviceType")}</FormLabel>
                        <Select
                          onValueChange={field.onChange}
                          defaultValue={field.value}
                        >
                          <FormControl>
                            <SelectTrigger>
                              <SelectValue
                                placeholder={t("devices.selectType")}
                              />
                            </SelectTrigger>
                          </FormControl>
                          <SelectContent>
                            <SelectItem value="android">
                              {t("devices.androidPhone")}
                            </SelectItem>
                            <SelectItem value="modem">
                              {t("devices.usbModem")}
                            </SelectItem>
                          </SelectContent>
                        </Select>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="name"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>
                          {t("common.name")}{" "}
                          <span className="text-destructive">*</span>
                        </FormLabel>
                        <FormControl>
                          <Input
                            placeholder={t("devices.namePlaceholder")}
                            type="text"
                            {...field}
                            required
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />

                  <FormField
                    control={form.control}
                    name="phone_number"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>
                          {t("devices.phoneNumber")}{" "}
                          <span className="text-destructive">*</span>
                        </FormLabel>
                        <FormControl>
                          <Input
                            placeholder={t("devices.phonePlaceholder")}
                            type="tel"
                            {...field}
                            required
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>

                <DialogFooter>
                  <DialogClose asChild>
                    <Button
                      variant="outline"
                      disabled={createDeviceMutation.isPending}
                    >
                      {t("common.cancel")}
                    </Button>
                  </DialogClose>
                  <LoadingButton
                    type="submit"
                    loading={createDeviceMutation.isPending}
                  >
                    {t("devices.createDevice")}
                  </LoadingButton>
                </DialogFooter>
              </form>
            </Form>
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}

export default AddDevice
