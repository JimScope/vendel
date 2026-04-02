import { zodResolver } from "@hookform/resolvers/zod"
import { Pencil } from "lucide-react"
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
} from "@/components/ui/dialog"
import { DropdownMenuItem } from "@/components/ui/dropdown-menu"
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
import { useUpdateContact } from "@/hooks/useContactMutations"
import type { Contact } from "@/types/collections"

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

interface EditContactProps {
  contact: Contact
  onSuccess: () => void
}

const EditContact = ({ contact, onSuccess }: EditContactProps) => {
  const { t } = useTranslation()
  const [isOpen, setIsOpen] = useState(false)
  const { data: groups } = useContactGroupList()

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    mode: "onBlur",
    criteriaMode: "all",
    defaultValues: {
      name: contact.name,
      phone_number: contact.phone_number,
      groups: contact.groups || [],
      notes: contact.notes || "",
    },
  })

  const updateContactMutation = useUpdateContact(contact.id)

  const onSubmit = (data: FormData) => {
    updateContactMutation.mutate(data, {
      onSuccess: () => {
        setIsOpen(false)
        onSuccess()
      },
    })
  }

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DropdownMenuItem
        onSelect={(e) => e.preventDefault()}
        onClick={() => setIsOpen(true)}
      >
        <Pencil />
        {t("contacts.editContact")}
      </DropdownMenuItem>
      <DialogContent className="sm:max-w-md">
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)}>
            <DialogHeader>
              <DialogTitle>{t("contacts.editContact")}</DialogTitle>
              <DialogDescription>
                {t("contacts.description")}
              </DialogDescription>
            </DialogHeader>
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
                      <Input placeholder="John Doe" type="text" {...field} required />
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
                  disabled={updateContactMutation.isPending}
                >
                  {t("common.cancel")}
                </Button>
              </DialogClose>
              <LoadingButton
                type="submit"
                loading={updateContactMutation.isPending}
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

export default EditContact
