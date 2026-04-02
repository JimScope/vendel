import {
  Clock,
  Contact,
  CreditCard,
  FileText,
  Home,
  Key,
  MessageSquare,
  Smartphone,
  Users,
  Webhook,
} from "lucide-react"

import { SidebarAppearance } from "@/components/Common/Appearance"
import { SidebarLanguage } from "@/components/Common/LanguageSwitcher"
import { Logo } from "@/components/Common/Logo"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
} from "@/components/ui/sidebar"
import useAuth from "@/hooks/useAuth"
import { type Item, Main } from "./Main"
import { User } from "./User"

const baseItems: Item[] = [
  { icon: Home, title: "sidebar.dashboard", path: "/" },
  { icon: MessageSquare, title: "sidebar.sms", path: "/sms" },
  { icon: FileText, title: "sidebar.templates", path: "/templates" },
  { icon: Clock, title: "sidebar.scheduled", path: "/scheduled" },
  { icon: Smartphone, title: "sidebar.devices", path: "/devices" },
  { icon: Contact, title: "sidebar.contacts", path: "/contacts" },
  { icon: Webhook, title: "sidebar.webhooks", path: "/webhooks" },
  { icon: Key, title: "sidebar.integrations", path: "/integrations" },
  { icon: CreditCard, title: "sidebar.billing", path: "/billing" },
]

export function AppSidebar() {
  const { user: currentUser } = useAuth()

  const items = currentUser?.is_superuser
    ? [...baseItems, { icon: Users, title: "sidebar.admin", path: "/admin" }]
    : baseItems

  return (
    <Sidebar collapsible="icon">
      <SidebarHeader className="px-4 py-6 group-data-[collapsible=icon]:px-0 group-data-[collapsible=icon]:items-center">
        <Logo variant="responsive" />
      </SidebarHeader>
      <SidebarContent>
        <Main items={items} />
      </SidebarContent>
      <SidebarFooter>
        <SidebarLanguage />
        <SidebarAppearance />
        <User user={currentUser} />
      </SidebarFooter>
    </Sidebar>
  )
}

export default AppSidebar
