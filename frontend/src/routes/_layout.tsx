import {
  createFileRoute,
  Outlet,
  redirect,
  useRouterState,
} from "@tanstack/react-router"
import { Construction } from "lucide-react"

import { Footer } from "@/components/Common/Footer"
import { Logo } from "@/components/Common/Logo"
import AppSidebar from "@/components/Sidebar/AppSidebar"
import { Badge } from "@/components/ui/badge"
import { Separator } from "@/components/ui/separator"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import useAppConfig from "@/hooks/useAppConfig"
import useAuth, { isLoggedIn } from "@/hooks/useAuth"

const PAGE_TITLES: Record<string, string> = {
  "/": "Dashboard",
  "/sms": "SMS",
  "/templates": "Templates",
  "/scheduled": "Scheduled",
  "/devices": "Devices",
  "/webhooks": "Webhooks",
  "/integrations": "Integrations",
  "/billing": "Billing",
  "/settings": "Settings",
  "/admin": "Admin",
}

export const Route = createFileRoute("/_layout")({
  component: Layout,
  beforeLoad: async () => {
    if (!isLoggedIn()) {
      throw redirect({
        to: "/login",
      })
    }
  },
})

function MaintenancePage() {
  const { config } = useAppConfig()

  return (
    <div className="flex min-h-svh flex-col items-center justify-center gap-6 p-6 text-center">
      <Logo asLink={false} />
      <div className="flex flex-col items-center gap-3">
        <div className="flex size-16 items-center justify-center rounded-full bg-muted">
          <Construction className="size-8 text-muted-foreground" />
        </div>
        <h1 className="text-2xl font-serif font-bold">Under Maintenance</h1>
        <p className="max-w-md text-muted-foreground">
          We're performing scheduled maintenance to improve your experience.
          Please check back shortly.
        </p>
        {config.supportEmail && config.supportEmail !== "support@example.com" && (
          <p className="text-sm text-muted-foreground">
            Need help?{" "}
            <a
              href={`mailto:${config.supportEmail}`}
              className="text-brand underline underline-offset-4"
            >
              Contact support
            </a>
          </p>
        )}
      </div>
    </div>
  )
}

function Layout() {
  const router = useRouterState()
  const currentPath = router.location.pathname
  const pageTitle = PAGE_TITLES[currentPath] ?? ""
  const { user } = useAuth()
  const { config } = useAppConfig()

  if (config.maintenanceMode && !user?.is_superuser) {
    return <MaintenancePage />
  }

  return (
    <SidebarProvider>
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:fixed focus:top-4 focus:left-4 focus:z-50 focus:rounded-md focus:bg-primary focus:px-4 focus:py-2 focus:text-primary-foreground focus:outline-none"
      >
        Skip to content
      </a>
      <AppSidebar />
      <SidebarInset>
        <header className="sticky top-0 z-10 flex h-16 shrink-0 items-center gap-2 border-b bg-background px-4">
          <SidebarTrigger className="-ml-1 text-muted-foreground" />
          {pageTitle && (
            <>
              <Separator orientation="vertical" className="mx-1 h-4" />
              <span className="text-sm text-muted-foreground">{pageTitle}</span>
            </>
          )}
          {config.maintenanceMode && (
            <Badge variant="destructive" className="ml-auto gap-1.5">
              <Construction className="size-3" />
              Maintenance Mode
            </Badge>
          )}
        </header>
        <main id="main-content" className="flex-1 p-6 md:p-8">
          <div className="mx-auto max-w-7xl">
            <Outlet />
          </div>
        </main>
        <Footer />
      </SidebarInset>
    </SidebarProvider>
  )
}

export default Layout
