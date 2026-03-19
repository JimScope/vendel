import { Check, ChevronDown, X } from "lucide-react"
import { useMemo, useState } from "react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { cn } from "@/lib/utils"

interface MultiSelectOption {
  label: string
  value: string
  disabled?: boolean
}

interface MultiSelectProps {
  options: MultiSelectOption[]
  onValueChange: (value: string[]) => void
  defaultValue?: string[]
  placeholder?: string
  maxVisibleBadges?: number
  className?: string
}

function MultiSelect({
  options,
  onValueChange,
  defaultValue = [],
  placeholder = "Select options...",
  maxVisibleBadges = 2,
  className,
}: MultiSelectProps) {
  const [open, setOpen] = useState(false)
  const selectedValues = defaultValue

  const handleSelect = (value: string) => {
    const newValues = selectedValues.includes(value)
      ? selectedValues.filter((v) => v !== value)
      : [...selectedValues, value]
    onValueChange(newValues)
  }

  const handleRemove = (value: string, e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    onValueChange(selectedValues.filter((v) => v !== value))
  }

  const { visibleValues, remainingCount } = useMemo(() => {
    if (selectedValues.length <= maxVisibleBadges) {
      return { visibleValues: selectedValues, remainingCount: 0 }
    }
    return {
      visibleValues: selectedValues.slice(0, maxVisibleBadges),
      remainingCount: selectedValues.length - maxVisibleBadges,
    }
  }, [selectedValues, maxVisibleBadges])

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className={cn(
            "w-full min-h-9 h-auto justify-between hover:bg-transparent",
            className,
          )}
          onClick={(e) => {
            if ((e.target as HTMLElement).closest("[data-badge-remove]")) {
              e.preventDefault()
              e.stopPropagation()
            }
          }}
        >
          <div className="flex items-center gap-1 w-full overflow-hidden">
            {selectedValues.length === 0 ? (
              <span className="text-muted-foreground truncate">
                {placeholder}
              </span>
            ) : (
              <div className="flex flex-wrap gap-1 flex-1 min-w-0">
                {visibleValues.map((value) => {
                  const option = options.find((opt) => opt.value === value)
                  return option ? (
                    <Badge
                      key={value}
                      variant="secondary"
                      className="gap-1 pr-1 shrink-0"
                      onClick={(e) => e.stopPropagation()}
                    >
                      <span className="truncate max-w-[150px]">
                        {option.label}
                      </span>
                      <button
                        type="button"
                        data-badge-remove
                        onClick={(e) => handleRemove(value, e)}
                        className="hover:bg-muted rounded-sm cursor-pointer inline-flex shrink-0"
                      >
                        <X className="size-3" />
                      </button>
                    </Badge>
                  ) : null
                })}
                {remainingCount > 0 && (
                  <Badge variant="outline" className="shrink-0">
                    +{remainingCount} more
                  </Badge>
                )}
              </div>
            )}
          </div>
          <ChevronDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent
        className="w-[var(--radix-popover-trigger-width)] p-0"
        align="start"
      >
        <Command>
          <CommandInput placeholder="Search..." />
          <CommandList>
            <CommandEmpty>No results found.</CommandEmpty>
            <CommandGroup>
              {options.map((option) => {
                const isSelected = selectedValues.includes(option.value)
                return (
                  <CommandItem
                    key={option.value}
                    value={option.value}
                    keywords={[option.label]}
                    onSelect={() => handleSelect(option.value)}
                    disabled={option.disabled}
                  >
                    <div
                      className={cn(
                        "mr-2 flex h-4 w-4 items-center justify-center rounded-sm border border-primary",
                        isSelected
                          ? "bg-primary text-primary-foreground"
                          : "opacity-50 [&_svg]:invisible",
                      )}
                    >
                      <Check className="size-4" />
                    </div>
                    <span>{option.label}</span>
                  </CommandItem>
                )
              })}
            </CommandGroup>
          </CommandList>
        </Command>
        {selectedValues.length > 0 && (
          <div className="p-2 border-t">
            <Button
              variant="ghost"
              size="sm"
              className="w-full justify-center"
              onClick={() => onValueChange([])}
            >
              Clear selection
            </Button>
          </div>
        )}
      </PopoverContent>
    </Popover>
  )
}

export type { MultiSelectOption, MultiSelectProps }
export { MultiSelect }
