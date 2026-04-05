import { zodResolver } from "@hookform/resolvers/zod"
import { Plus } from "lucide-react"
import { useState } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { z } from "zod"

import { MultiSelect } from "@/components/Common/MultiSelect"
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
import { Textarea } from "@/components/ui/textarea"
import { useContactGroupList } from "@/hooks/useContactGroupList"
import { useCreateContact } from "@/hooks/useContactMutations"

const formSchema = z.object({
  name: z
    .string()
    .min(1, { message: "Name is required" })
    .max(255, { message: "Name must be at most 255 characters" }),
  phone_number: z
    .string()
    .min(1, { message: "Phone number is required" })
    .max(20, { message: "Phone number must be at most 20 characters" }),
  groups: z.array(z.string()).optional(),
  notes: z.string().max(1000).optional(),
})

type FormData = z.infer<typeof formSchema>

interface AddContactProps {
  open?: boolean
  onOpenChange?: (open: boolean) => void
}

const AddContact = ({ open, onOpenChange }: AddContactProps) => {
  const { t } = useTranslation()
  const [internalOpen, setInternalOpen] = useState(false)
  const isOpen = open ?? internalOpen
  const setIsOpen = onOpenChange ?? setInternalOpen
  const { data: groups } = useContactGroupList()

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    mode: "onBlur",
    criteriaMode: "all",
    defaultValues: {
      name: "",
      phone_number: "",
      groups: [],
      notes: "",
    },
  })

  const createContactMutation = useCreateContact()

  const onSubmit = (data: FormData) => {
    createContactMutation.mutate(
      {
        name: data.name,
        phone_number: data.phone_number,
        groups: data.groups,
        notes: data.notes,
      },
      {
        onSuccess: () => {
          form.reset()
          setIsOpen(false)
        },
      },
    )
  }

  const handleClose = (open: boolean) => {
    if (!open) {
      form.reset()
    }
    setIsOpen(open)
  }

  return (
    <Dialog open={isOpen} onOpenChange={handleClose}>
      <DialogTrigger asChild>
        <Button className="my-4">
          <Plus />
          {t("contacts.addContact")}
        </Button>
      </DialogTrigger>
      <DialogContent
        className="sm:max-w-md"
        onInteractOutside={(e) => e.preventDefault()}
      >
        <DialogHeader>
          <DialogTitle>{t("contacts.addContact")}</DialogTitle>
          <DialogDescription>{t("contacts.description")}</DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)}>
            <div className="grid gap-4 py-4">
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>
                      {t("contacts.name")}{" "}
                      <span className="text-destructive">*</span>
                    </FormLabel>
                    <FormControl>
                      <Input
                        placeholder="John Doe"
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
                      {t("contacts.phoneNumber")}{" "}
                      <span className="text-destructive">*</span>
                    </FormLabel>
                    <FormControl>
                      <Input
                        placeholder="+1234567890"
                        type="tel"
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
                name="groups"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("contacts.groups")}</FormLabel>
                    <FormControl>
                      <MultiSelect
                        options={(groups?.data || []).map((group) => ({
                          label: group.name,
                          value: group.id,
                        }))}
                        onValueChange={field.onChange}
                        defaultValue={field.value || []}
                        placeholder={t("contacts.groups")}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="notes"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("contacts.notes")}</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder={t("contacts.notes")}
                        rows={2}
                        {...field}
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
                  disabled={createContactMutation.isPending}
                >
                  {t("common.cancel")}
                </Button>
              </DialogClose>
              <LoadingButton
                type="submit"
                loading={createContactMutation.isPending}
              >
                {t("common.create")}
              </LoadingButton>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}

export default AddContact
