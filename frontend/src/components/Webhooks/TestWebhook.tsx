import { FlaskConical, Loader2 } from "lucide-react"

import { DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { useTestWebhook } from "@/hooks/useWebhookTestMutation"

interface TestWebhookProps {
  webhookId: string
  onSuccess: () => void
}

const TestWebhook = ({ webhookId, onSuccess }: TestWebhookProps) => {
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
        <Loader2 className="animate-spin" />
      ) : (
        <FlaskConical />
      )}
      Test Webhook
    </DropdownMenuItem>
  )
}

export default TestWebhook
