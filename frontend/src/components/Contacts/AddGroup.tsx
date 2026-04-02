import { zodResolver } from "@hookform/resolvers/zod"
import { FolderPlus } from "lucide-react"
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
import { useCreateContactGroup } from "@/hooks/useContactGroupMutations"

const formSchema = z.object({
  name: z
    .string()
    .min(1, { message: "Name is required" })
    .max(255, { message: "Name must be at most 255 characters" }),
})

type FormData = z.infer<typeof formSchema>

const AddGroup = () => {
  const { t } = useTranslation()
  const [isOpen, setIsOpen] = useState(false)

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
    mode: "onBlur",
    criteriaMode: "all",
    defaultValues: {
      name: "",
    },
  })

  const createGroupMutation = useCreateContactGroup()

  const onSubmit = (data: FormData) => {
    createGroupMutation.mutate(data, {
      onSuccess: () => {
        form.reset()
        setIsOpen(false)
      },
    })
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
        <Button variant="outline">
          <FolderPlus className="size-4" />
          {t("contacts.addGroup")}
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t("contacts.addGroup")}</DialogTitle>
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
                      {t("contacts.groupName")}{" "}
                      <span className="text-destructive">*</span>
                    </FormLabel>
                    <FormControl>
                      <Input
                        placeholder={t("contacts.groupName")}
                        type="text"
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
                  disabled={createGroupMutation.isPending}
                >
                  {t("common.cancel")}
                </Button>
              </DialogClose>
              <LoadingButton
                type="submit"
                loading={createGroupMutation.isPending}
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

export default AddGroup
