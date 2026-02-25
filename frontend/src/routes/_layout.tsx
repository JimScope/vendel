import {
  createFileRoute,
  Outlet,
  redirect,
  useRouterState,
} from "@tanstack/react-router"

import { Footer } from "@/components/Common/Footer"
import AppSidebar from "@/components/Sidebar/AppSidebar"
import { Separator } from "@/components/ui/separator"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { isLoggedIn } from "@/hooks/useAuth"

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

function Layout() {
  const router = useRouterState()
  const currentPath = router.location.pathname
  const pageTitle = PAGE_TITLES[currentPath] ?? ""

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
