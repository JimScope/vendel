import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useTemplateList } from "@/hooks/useTemplateList"

interface TemplateSelectProps {
  onSelect: (body: string) => void
}

export const TemplateSelect = ({ onSelect }: TemplateSelectProps) => {
  const { data: templates } = useTemplateList()

  if (!templates?.data || templates.data.length === 0) {
    return null
  }

  return (
    <Select
      onValueChange={(value) => {
        const template = templates.data.find(
          (t: Record<string, any>) => t.id === value,
        )
        if (template) {
          onSelect(template.body)
        }
      }}
    >
      <SelectTrigger className="w-full">
        <SelectValue placeholder="Use a template..." />
      </SelectTrigger>
      <SelectContent>
        {templates.data.map((template: Record<string, any>) => (
          <SelectItem key={template.id} value={template.id}>
            {template.name}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
