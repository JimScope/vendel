import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { z } from "zod"

import ChangePassword from "@/components/UserSettings/ChangePassword"
import DeleteAccount from "@/components/UserSettings/DeleteAccount"
import PlanSettings from "@/components/UserSettings/PlanSettings"
import UserInformation from "@/components/UserSettings/UserInformation"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import useAuth from "@/hooks/useAuth"

const tabsConfig = [
  { value: "my-profile", title: "My profile", component: UserInformation },
  { value: "password", title: "Password", component: ChangePassword },
  { value: "plan", title: "Plan & Usage", component: PlanSettings },
  { value: "danger-zone", title: "Danger zone", component: DeleteAccount },
]

const searchSchema = z.object({
  tab: z.string().optional().catch(undefined),
})

export const Route = createFileRoute("/_layout/settings")({
  component: UserSettings,
  validateSearch: searchSchema,
  head: () => ({
    meta: [
      {
        title: "Settings - Vendel",
      },
    ],
  }),
})

function UserSettings() {
  const { user: currentUser } = useAuth()
  const { tab } = Route.useSearch()
  const navigate = useNavigate()
  const finalTabs = currentUser?.is_superuser
    ? tabsConfig.slice(0, 4)
    : tabsConfig

  const activeTab = finalTabs.some((t) => t.value === tab) ? tab : "my-profile"

  if (!currentUser) {
    return null
  }

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-2xl">User Settings</h1>
        <p className="text-muted-foreground">
          Manage your account settings and preferences
        </p>
      </div>

      <Tabs
        value={activeTab}
        onValueChange={(value) =>
          navigate({
            to: "/settings",
            search: { tab: value },
            replace: true,
          })
        }
      >
        <TabsList>
          {finalTabs.map((tab) => (
            <TabsTrigger key={tab.value} value={tab.value}>
              {tab.title}
            </TabsTrigger>
          ))}
        </TabsList>
        {finalTabs.map((tab) => (
          <TabsContent key={tab.value} value={tab.value}>
            <tab.component />
          </TabsContent>
        ))}
      </Tabs>
    </div>
  )
}
