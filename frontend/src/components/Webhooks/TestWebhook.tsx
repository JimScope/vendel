import { FlaskConical, LoaderCircle } from "lucide-react"
import { useTranslation } from "react-i18next"

import { DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { useTestWebhook } from "@/hooks/useWebhookTestMutation"

interface TestWebhookProps {
  webhookId: string
  onSuccess: () => void
}

const TestWebhook = ({ webhookId, onSuccess }: TestWebhookProps) => {
  const { t } = useTranslation()
  const testMutation = useTestWebhook()

  const handleTest = () => {
    testMutation.mutate(webhookId, {
      onSettled: () => onSuccess(),
    })
  }

  return (
    <DropdownMenuItem
      onSelect={(e) => e.preventDefault()}
      onClick={handleTest}
      disabled={testMutation.isPending}
    >
      {testMutation.isPending ? (
        <LoaderCircle className="animate-spin" />
      ) : (
        <FlaskConical />
      )}
      {t("webhooks.testWebhook")}
    </DropdownMenuItem>
  )
}

export default TestWebhook
