import * as React from "react"
import { X } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"

export interface TagSuggestion {
  label: string
  value: string
}

export interface TagInputProps
  extends Omit<React.InputHTMLAttributes<HTMLInputElement>, "value" | "onChange"> {
  value: string[]
  onChange: (value: string[]) => void
  suggestions?: TagSuggestion[]
}

const TagInput = React.forwardRef<HTMLInputElement, TagInputProps>(
  ({ className, value, onChange, placeholder, suggestions, ...props }, ref) => {
    const [inputValue, setInputValue] = React.useState("")
    const [showSuggestions, setShowSuggestions] = React.useState(false)
    const [activeIndex, setActiveIndex] = React.useState(-1)
    const containerRef = React.useRef<HTMLDivElement>(null)

    const filtered = React.useMemo(() => {
      if (!suggestions || inputValue.length < 2) return []
      const query = inputValue.toLowerCase()
      return suggestions.filter(
        (s) =>
          !value.includes(s.value) &&
          (s.label.toLowerCase().includes(query) ||
            s.value.includes(query)),
      )
    }, [suggestions, inputValue, value])

    const addTag = (tag: string) => {
      const trimmed = tag.trim()
      if (trimmed && !value.includes(trimmed)) {
        onChange([...value, trimmed])
      }
      setInputValue("")
      setShowSuggestions(false)
      setActiveIndex(-1)
    }

    const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (filtered.length > 0 && showSuggestions) {
        if (e.key === "ArrowDown") {
          e.preventDefault()
          setActiveIndex((i) => (i + 1) % filtered.length)
          return
        }
        if (e.key === "ArrowUp") {
          e.preventDefault()
          setActiveIndex((i) => (i <= 0 ? filtered.length - 1 : i - 1))
          return
        }
        if (e.key === "Enter" && activeIndex >= 0) {
          e.preventDefault()
          addTag(filtered[activeIndex].value)
          return
        }
        if (e.key === "Escape") {
          setShowSuggestions(false)
          setActiveIndex(-1)
          return
        }
      }

      if (e.key === " " || e.key === "Enter") {
        e.preventDefault()
        addTag(inputValue)
      } else if (e.key === "Backspace" && !inputValue && value.length > 0) {
        onChange(value.slice(0, -1))
      }
    }

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
      setInputValue(e.target.value)
      setShowSuggestions(true)
      setActiveIndex(-1)
    }

    const handleBlur = () => {
      // Delay to allow suggestion click to fire
      setTimeout(() => {
        addTag(inputValue)
        setShowSuggestions(false)
      }, 150)
    }

    const removeTag = (tagToRemove: string) => {
      onChange(value.filter((tag) => tag !== tagToRemove))
    }

    return (
      <div ref={containerRef} className="relative">
        <div
          className={cn(
            "flex min-h-10 w-full flex-wrap gap-2 rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background disabled:cursor-not-allowed disabled:opacity-50 focus-within:ring-2 focus-within:ring-ring focus-within:ring-offset-2",
            className,
          )}
        >
          {value.map((tag) => (
            <Badge key={tag} variant="secondary" className="gap-1 px-1">
              {tag}
              <button
                type="button"
                className="rounded-full outline-none ring-offset-background focus:ring-2 focus:ring-ring focus:ring-offset-2"
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    removeTag(tag)
                  }
                }}
                onMouseDown={(e) => {
                  e.preventDefault()
                  e.stopPropagation()
                }}
                onClick={() => removeTag(tag)}
              >
                <X className="h-3 w-3 text-muted-foreground hover:text-foreground" />
              </button>
            </Badge>
          ))}
          <input
            {...props}
            ref={ref}
            value={inputValue}
            onChange={handleChange}
            onKeyDown={handleKeyDown}
            onBlur={handleBlur}
            onFocus={() => inputValue.length >= 2 && setShowSuggestions(true)}
            placeholder={value.length === 0 ? placeholder : ""}
            className="flex-1 bg-transparent outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed"
          />
        </div>

        {showSuggestions && filtered.length > 0 && (
          <ul className="absolute z-50 mt-1 max-h-48 w-full overflow-auto rounded-md border bg-popover p-1 shadow-md">
            {filtered.map((suggestion, i) => (
              <li
                key={suggestion.value}
                onMouseDown={(e) => {
                  e.preventDefault()
                  addTag(suggestion.value)
                }}
                className={cn(
                  "flex cursor-pointer items-center justify-between rounded-sm px-2 py-1.5 text-sm",
                  i === activeIndex
                    ? "bg-accent text-accent-foreground"
                    : "hover:bg-accent/50",
                )}
              >
                <span>{suggestion.label}</span>
                <span className="font-mono text-xs text-muted-foreground">
                  {suggestion.value}
                </span>
              </li>
            ))}
          </ul>
        )}
      </div>
    )
  },
)
TagInput.displayName = "TagInput"

export { TagInput }
