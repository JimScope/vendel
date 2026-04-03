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

const VARIABLE_REGEX = /\{\{([a-zA-Z_][a-zA-Z0-9_]*)\}\}/g
const RESERVED_VARIABLES = new Set(["name", "phone"])

export interface SelectedTemplate {
  id: string
  body: string
  customVariables: string[]
}

function extractCustomVariables(body: string): string[] {
  const seen = new Set<string>()
  const vars: string[] = []
  for (const match of body.matchAll(VARIABLE_REGEX)) {
    const name = match[1]
    if (!seen.has(name) && !RESERVED_VARIABLES.has(name)) {
      seen.add(name)
      vars.push(name)
    }
  }
  return vars
}

interface TemplateSelectProps {
  onSelect: (template: SelectedTemplate | null) => void
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
          onSelect(null)
          return
        }
        setSelected(value)
        const template = templates.data.find((t) => t.id === value)
        if (template) {
          onSelect({
            id: template.id,
            body: template.body,
            customVariables: extractCustomVariables(template.body),
          })
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
