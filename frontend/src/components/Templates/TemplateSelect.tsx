import { useState } from "react"
import { useTranslation } from "react-i18next"
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
  const { t } = useTranslation()
  const { data: templates } = useTemplateList()
  const [selected, setSelected] = useState<string>("")

  if (!templates?.data || templates.data.length === 0) {
    return null
  }

  return (
    <Select
      value={selected}
      onValueChange={(value) => {
        if (value === "__none__") {
          setSelected("")
          return
        }
        setSelected(value)
        const template = templates.data.find((t) => t.id === value)
        if (template) {
          onSelect(template.body)
        }
      }}
    >
      <SelectTrigger className="w-full">
        <SelectValue placeholder={t("templates.useTemplate")} />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="__none__" className="text-muted-foreground">
          {t("templates.none")}
        </SelectItem>
        {templates.data.map((template) => (
          <SelectItem key={template.id} value={template.id}>
            {template.name}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
