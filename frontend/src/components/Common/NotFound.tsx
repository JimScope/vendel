import { Link } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"
import { Button } from "@/components/ui/button"

const NotFound = () => {
  const { t } = useTranslation()
  return (
    <div
      className="flex min-h-screen items-center justify-center flex-col p-4"
      data-testid="not-found"
    >
      <div className="flex items-center">
        <div className="flex flex-col ml-4 items-center justify-center p-4">
          <span className="text-6xl md:text-8xl font-serif leading-none mb-4">
            404
          </span>
          <span className="text-2xl font-serif mb-2">
            {t("errors.notFoundTitle")}
          </span>
        </div>
      </div>

      <p className="text-lg text-muted-foreground mb-4 text-center">
        {t("errors.notFoundMsg")}
      </p>
      <Link to="/">
        <Button className="mt-4">{t("common.goBack")}</Button>
      </Link>
    </div>
  )
}

export default NotFound
