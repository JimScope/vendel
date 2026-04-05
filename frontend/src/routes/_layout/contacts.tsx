import { createFileRoute } from "@tanstack/react-router"
import { Plus, Users } from "lucide-react"
import { Suspense, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

import { DataTable } from "@/components/Common/DataTable"
import AddContact from "@/components/Contacts/AddContact"
import AddGroup from "@/components/Contacts/AddGroup"
import { getColumns } from "@/components/Contacts/columns"
import ImportContacts from "@/components/Contacts/ImportContacts"
import PendingContacts from "@/components/Pending/PendingContacts"
import { Button } from "@/components/ui/button"
import useAppConfig from "@/hooks/useAppConfig"
import { useContactGroupList } from "@/hooks/useContactGroupList"
import { useContactListSuspense } from "@/hooks/useContactList"
import type { Contact, ContactGroup } from "@/types/collections"

export const Route = createFileRoute("/_layout/contacts")({
  component: Contacts,
})

function ContactsEmptyState({ onAddContact }: { onAddContact: () => void }) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-col items-center justify-center text-center py-12">
      <div className="rounded-full bg-muted p-4 mb-4">
        <Users className="h-8 w-8 text-muted-foreground" />
      </div>
      <h2 className="text-lg font-semibold">{t("contacts.noContacts")}</h2>
      <p className="text-muted-foreground">{t("contacts.noContactsDesc")}</p>
      <Button className="my-4" onClick={onAddContact}>
        <Plus />
        {t("contacts.addContact")}
      </Button>
    </div>
  )
}

function ContactsTableContent({ onAddContact }: { onAddContact: () => void }) {
  const { t } = useTranslation()
  const { data: contacts } = useContactListSuspense()
  const { data: groups } = useContactGroupList()
  const columns = useMemo(
    () => getColumns(t, (groups?.data ?? []) as unknown as ContactGroup[]),
    [t, groups],
  )

  if (!contacts?.data || contacts.data.length === 0) {
    return <ContactsEmptyState onAddContact={onAddContact} />
  }

  return (
    <DataTable
      columns={columns}
      data={(contacts?.data ?? []) as unknown as Contact[]}
      caption={t("contacts.title")}
    />
  )
}

function ContactsTable({ onAddContact }: { onAddContact: () => void }) {
  return (
    <Suspense fallback={<PendingContacts />}>
      <ContactsTableContent onAddContact={onAddContact} />
    </Suspense>
  )
}

function Contacts() {
  const { t } = useTranslation()
  const { config } = useAppConfig()
  const [addContactOpen, setAddContactOpen] = useState(false)

  return (
    <div className="flex flex-col gap-6">
      <title>{`${t("contacts.title")} - ${config.appName}`}</title>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl">{t("contacts.title")}</h1>
          <p className="text-muted-foreground">{t("contacts.description")}</p>
        </div>
        <div className="flex items-center gap-2">
          <ImportContacts />
          <AddGroup />
          <AddContact open={addContactOpen} onOpenChange={setAddContactOpen} />
        </div>
      </div>
      <ContactsTable onAddContact={() => setAddContactOpen(true)} />
    </div>
  )
}
