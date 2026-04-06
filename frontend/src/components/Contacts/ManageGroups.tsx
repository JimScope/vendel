import { FolderOpen, Settings } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { useContactGroupList } from "@/hooks/useContactGroupList"
import type { ContactGroup } from "@/types/collections"
import { GroupActionsMenu } from "./GroupActionsMenu"

const ManageGroups = () => {
  const { t } = useTranslation()
  const [isOpen, setIsOpen] = useState(false)
  const { data: groups } = useContactGroupList()

  const groupList = (groups?.data ?? []) as unknown as ContactGroup[]

  return (
    <Dialog open={isOpen} onOpenChange={setIsOpen}>
      <DialogTrigger asChild>
        <Button variant="outline">
          <Settings className="size-4" />
          {t("contacts.manageGroups")}
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t("contacts.manageGroups")}</DialogTitle>
          <DialogDescription>{t("contacts.description")}</DialogDescription>
        </DialogHeader>
        <div className="py-4">
          {groupList.length === 0 ? (
            <div className="flex flex-col items-center justify-center text-center py-8">
              <div className="rounded-full bg-muted p-4 mb-4">
                <FolderOpen className="h-8 w-8 text-muted-foreground" />
              </div>
              <p className="text-sm font-medium">{t("contacts.noGroups")}</p>
              <p className="text-sm text-muted-foreground">
                {t("contacts.noGroupsDesc")}
              </p>
            </div>
          ) : (
            <div className="space-y-1">
              {groupList.map((group) => (
                <div
                  key={group.id}
                  className="flex items-center justify-between rounded-md px-3 py-2 hover:bg-muted"
                >
                  <span className="text-sm font-medium">{group.name}</span>
                  <GroupActionsMenu group={group} />
                </div>
              ))}
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}

export default ManageGroups
