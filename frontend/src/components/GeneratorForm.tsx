import { useEffect, useRef, useCallback } from "react"
import type { GeneratorSchema } from "../lib/types"
import { generateSingle } from "../lib/api"
import { ForgeButton } from "./ForgeButton"
import type { Artifact } from "../lib/types"

interface GeneratorFormProps {
  schema: GeneratorSchema
  values: Record<string, string>
  onChange: (values: Record<string, string>) => void
  onForge: () => void
  onPreview: (artifact: Artifact | null) => void
  forging: boolean
  forgeDone: boolean
  maxCount?: number
}

/** Map slider position (0-1) to a log-scale count value. */
function sliderToCount(pos: number, max: number): number {
  if (pos <= 0) return 1
  if (pos >= 1) return max
  return Math.round(Math.exp(pos * Math.log(max)))
}

/** Map count value to slider position (0-1). */
function countToSlider(count: number, max: number): number {
  if (count <= 1) return 0
  if (count >= max) return 1
  return Math.log(count) / Math.log(max)
}

export function GeneratorForm({
  schema,
  values,
  onChange,
  onForge,
  onPreview,
  forging,
  forgeDone,
  maxCount = 10000,
}: GeneratorFormProps) {
  const previewTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Stable key for preview dependencies — excludes count so slider
  // changes don't re-fetch a new random value ("scramble" effect).
  const previewKey = Object.entries(values)
    .filter(([k]) => k !== "count")
    .map(([k, v]) => `${k}=${v}`)
    .join("&")

  const fetchPreview = useCallback(() => {
    const format = values["format"] || "plain"
    const seed = values["seed"] || ""
    const params: Record<string, string> = {}
    for (const [k, v] of Object.entries(values)) {
      if (k !== "count" && k !== "format" && k !== "seed") {
        params[k] = v
      }
    }
    generateSingle(schema.address, params, format, seed)
      .then((result) => onPreview({ value: result.value, fields: result.fields, sensitiveKeys: result.sensitiveKeys }))
      .catch(() => onPreview(null))
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [schema.address, previewKey, onPreview])

  // Debounced live preview on value change.
  useEffect(() => {
    if (previewTimer.current) clearTimeout(previewTimer.current)
    previewTimer.current = setTimeout(fetchPreview, 200)
    return () => {
      if (previewTimer.current) clearTimeout(previewTimer.current)
    }
  }, [fetchPreview])

  function setValue(key: string, val: string) {
    onChange({ ...values, [key]: val })
  }

  // Separate common flags from generator-specific flags.
  const commonNames = new Set(["count", "format", "seed"])
  const commonFlags = schema.flags.filter((f) => commonNames.has(f.name) && f.name !== "seed")
  const specificFlags = schema.flags.filter((f) => !commonNames.has(f.name))
  const seedFlag = schema.flags.find((f) => f.name === "seed")

  return (
    <div className="flex flex-col gap-4 h-full">
      <div className="flex-1 overflow-y-auto space-y-4 pr-2">
        {/* Common flags: count, format */}
        {commonFlags.map((flag) => (
          <FlagControl
            key={flag.name}
            flag={flag}
            value={values[flag.name] ?? flag.default ?? ""}
            onChange={(v) => setValue(flag.name, v)}
            maxCount={maxCount}
            isCount={flag.name === "count"}
          />
        ))}

        {/* Generator-specific flags */}
        {specificFlags.map((flag) => (
          <FlagControl
            key={flag.name}
            flag={flag}
            value={values[flag.name] ?? flag.default ?? ""}
            onChange={(v) => setValue(flag.name, v)}
            maxCount={maxCount}
            isCount={false}
          />
        ))}

        {/* Seed input (only if supported) */}
        {schema.supports.seed && seedFlag && (
          <FlagControl
            flag={seedFlag}
            value={values["seed"] ?? ""}
            onChange={(v) => setValue("seed", v)}
            maxCount={maxCount}
            isCount={false}
          />
        )}
        {schema.supports.seed && !seedFlag && (
          <div>
            <label className="block text-xs text-muted-foreground mb-1">seed</label>
            <input
              type="text"
              value={values["seed"] ?? ""}
              onChange={(e) => setValue("seed", e.target.value)}
              placeholder="optional seed for determinism"
              className="w-full bg-panel border border-border text-foreground font-mono px-3 py-1.5 rounded-[4px] text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-forge"
            />
          </div>
        )}
      </div>

      <div className="pt-2">
        <ForgeButton onClick={onForge} forging={forging} done={forgeDone} />
      </div>
    </div>
  )
}

// --- Individual flag control ---

interface FlagControlProps {
  flag: { name: string; type: string; default?: string; description: string; options?: string[] }
  value: string
  onChange: (v: string) => void
  maxCount: number
  isCount: boolean
}

function FlagControl({ flag, value, onChange, maxCount, isCount }: FlagControlProps) {
  // Any flag with preset options: segmented control.
  if (flag.options && flag.options.length > 0) {
    return (
      <div>
        <label className="block text-xs text-muted-foreground mb-1">{flag.name}</label>
        <div className="flex flex-wrap gap-1">
          {flag.options.map((opt) => (
            <button
              key={opt}
              onClick={() => onChange(opt)}
              className={`text-xs font-mono px-3 py-1 border transition-colors duration-150 rounded-[4px] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-forge ${
                value === opt
                  ? "bg-forge border-forge text-foreground"
                  : "border-border text-muted-foreground hover:text-foreground hover:border-border-hover"
              }`}
            >
              {opt}
            </button>
          ))}
        </div>
        <p className="text-xs text-muted-foreground mt-1">{flag.description}</p>
      </div>
    )
  }

  // Boolean flag: toggle switch.
  if (flag.type === "bool") {
    const checked = value === "true"
    return (
      <div className="flex items-center justify-between gap-2">
        <div>
          <label className="block text-xs text-muted-foreground">{flag.name}</label>
          <p className="text-xs text-muted-foreground">{flag.description}</p>
        </div>
        <button
          type="button"
          role="switch"
          aria-checked={checked}
          onClick={() => onChange(checked ? "false" : "true")}
          className={`relative inline-flex h-5 w-9 shrink-0 rounded-full border transition-colors duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-forge ${
            checked ? "bg-forge border-forge" : "bg-panel border-border"
          }`}
        >
          <span
            className={`pointer-events-none block h-4 w-4 rounded-full bg-foreground transition-transform duration-150 ${
              checked ? "translate-x-4" : "translate-x-0"
            }`}
          />
        </button>
      </div>
    )
  }

  // Integer flag.
  if (flag.type === "int") {
    const numVal = parseInt(value, 10) || 1
    return (
      <div>
        <label className="block text-xs text-muted-foreground mb-1">{flag.name}</label>
        <input
          type="number"
          min={1}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="w-full bg-panel border border-border text-foreground font-mono px-3 py-1.5 rounded-[4px] text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-forge"
        />
        {isCount && (
          <input
            type="range"
            min={0}
            max={1}
            step={0.001}
            value={countToSlider(numVal, maxCount)}
            onChange={(e) => {
              const count = sliderToCount(parseFloat(e.target.value), maxCount)
              onChange(String(count))
            }}
            className="w-full mt-1 accent-forge"
          />
        )}
        <p className="text-xs text-muted-foreground mt-1">{flag.description}</p>
      </div>
    )
  }

  // Default: text input. Use password type for sensitive fields.
  const isSensitive = flag.name === "passphrase"
  return (
    <div>
      <label className="block text-xs text-muted-foreground mb-1">{flag.name}</label>
      <input
        type={isSensitive ? "password" : "text"}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={flag.default ?? ""}
        className="w-full bg-panel border border-border text-foreground font-mono px-3 py-1.5 rounded-[4px] text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-forge"
      />
      <p className="text-xs text-muted-foreground mt-1">{flag.description}</p>
    </div>
  )
}
