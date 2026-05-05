import { useState, useEffect, useCallback, useRef } from "react"
import { ChevronDown, ChevronRight } from "lucide-react"
import type { GeneratorSchema, Artifact, SSEEvent } from "../lib/types"
import { fetchGeneratorSchema, generateSingle, generateSSE } from "../lib/api"
import { trackForge } from "../lib/analytics"
import { Reveal } from "./primitives"
import { GeneratorForm } from "./GeneratorForm"
import { OutputPane } from "./OutputPane"
import { Toast } from "./Toast"

interface GeneratorPanelProps {
  address: string
  maxCount?: number
}

export function GeneratorPanel({ address, maxCount }: GeneratorPanelProps) {
  const [schema, setSchema] = useState<GeneratorSchema | null>(null)
  const [values, setValues] = useState<Record<string, string>>({})
  const [preview, setPreview] = useState<Artifact | null>(null)
  const [results, setResults] = useState<Artifact[]>([])
  const [forging, setForging] = useState(false)
  const [forgeDone, setForgeDone] = useState(false)
  const [status, setStatus] = useState("")
  const [progress, setProgress] = useState<{
    current: number
    total: number
    opsPerSec: number
  } | null>(null)
  const [durationMs, setDurationMs] = useState<number | null>(null)
  const [formatted, setFormatted] = useState<string | null>(null)
  const [toastVisible, setToastVisible] = useState(false)
  const [rationaleOpen, setRationaleOpen] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const abortRef = useRef<AbortController | null>(null)

  // Fetch schema when address changes.
  useEffect(() => {
    setSchema(null)
    setValues({})
    setPreview(null)
    setResults([])
    setError(null)
    setForging(false)
    setForgeDone(false)
    setStatus("")
    setProgress(null)
    setDurationMs(null)
    setFormatted(null)
    setRationaleOpen(false)

    fetchGeneratorSchema(address)
      .then((s) => {
        setSchema(s)
        // Initialize values from defaults.
        const defaults: Record<string, string> = {}
        for (const flag of s.flags) {
          if (flag.default !== undefined) {
            defaults[flag.name] = flag.default
          }
        }
        setValues(defaults)
      })
      .catch((err) => setError(err.message))

    return () => {
      if (abortRef.current) abortRef.current.abort()
    }
  }, [address])

  const handlePreview = useCallback((a: Artifact | null) => setPreview(a), [])

  function handleForge() {
    if (!schema || forging) return

    const count = parseInt(values["count"] || "1", 10) || 1
    const format = values["format"] || "plain"
    const seed = values["seed"] || ""
    const params: Record<string, string> = {}
    for (const [k, v] of Object.entries(values)) {
      if (k !== "count" && k !== "format" && k !== "seed") {
        params[k] = v
      }
    }

    setForging(true)
    setForgeDone(false)
    setResults([])
    setStatus("")
    setProgress(null)
    setDurationMs(null)
    setFormatted(null)
    setError(null)

    if (count <= 1) {
      generateSingle(schema.address, params, format, seed)
        .then((result) => {
          setResults([{ value: result.value, fields: result.fields, sensitiveKeys: result.sensitiveKeys }])
          if (result.formatted) setFormatted(result.formatted)
          setForging(false)
          showForgeDone()
        })
        .catch((err) => {
          setError(err.message)
          setForging(false)
        })
      return
    }

    // Stream via SSE for count > 1.
    const ctrl = generateSSE(schema.address, count, params, format, seed, (event: SSEEvent) => {
      switch (event.type) {
        case "status":
          setStatus(event.message)
          break
        case "progress":
          setProgress({
            current: event.current,
            total: event.total,
            opsPerSec: event.opsPerSec,
          })
          break
        case "artifact":
          setResults((prev) => [...prev, { value: event.value, fields: event.fields, sensitiveKeys: event.sensitiveKeys }])
          break
        case "done":
          setDurationMs(event.durationMs)
          if (event.formatted) setFormatted(event.formatted)
          setForging(false)
          showForgeDone()
          break
        case "error":
          setError(event.message)
          setForging(false)
          break
      }
    })
    abortRef.current = ctrl
  }

  function showForgeDone() {
    trackForge(address)
    setForgeDone(true)
    setToastVisible(true)
    setTimeout(() => {
      setForgeDone(false)
      setToastVisible(false)
    }, 1500)
  }

  function handleCopyAll() {
    const text = formatted || results.map((a) => a.value).join("\n")
    navigator.clipboard.writeText(text)
  }

  if (error && !schema) {
    return (
      <div className="p-6">
        <p className="text-red-400 font-mono text-sm">{error}</p>
      </div>
    )
  }

  if (!schema) {
    return (
      <div className="p-6">
        <p className="text-muted-foreground text-sm font-mono">Loading…</p>
      </div>
    )
  }

  return (
    <Reveal className="flex-1 min-h-0 flex flex-col">
      <div className="flex flex-col h-full">
        {/* Header */}
        <div className="px-6 py-4 border-b border-border">
          <div className="flex items-center gap-3">
            <h1 className="font-mono text-2xl text-foreground">{schema.name}</h1>
            <span className="bg-forge text-foreground text-xs px-2 py-0.5 rounded-none">
              {schema.category}
            </span>
          </div>
          <p className="text-muted-foreground text-sm mt-1">{schema.description}</p>
          {schema.rationale && (
            <button
              onClick={() => setRationaleOpen(!rationaleOpen)}
              className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground mt-2 transition-colors duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-forge rounded-[4px]"
            >
              {rationaleOpen ? (
                <ChevronDown className="w-3 h-3" />
              ) : (
                <ChevronRight className="w-3 h-3" />
              )}
              Why this default
            </button>
          )}
          {rationaleOpen && schema.rationale && (
            <div className="mt-2 bg-panel border border-border p-3 text-xs text-muted-foreground font-mono">
              {schema.rationale}
            </div>
          )}
        </div>

        {/* Body: split pane */}
        <div className="flex flex-1 min-h-0">
          {/* Left: form */}
          <div className="w-[45%] p-6 border-r border-border overflow-y-auto">
            <GeneratorForm
              schema={schema}
              values={values}
              onChange={setValues}
              onForge={handleForge}
              onPreview={handlePreview}
              forging={forging}
              forgeDone={forgeDone}
              maxCount={maxCount}
            />
          </div>

          {/* Right: output */}
          <div className="w-[55%] p-6 overflow-y-auto">
            <OutputPane
              preview={preview}
              results={results}
              forging={forging}
              progress={progress}
              status={status}
              schema={schema}
              onCopyAll={handleCopyAll}
              durationMs={durationMs}
              formatted={formatted}
            />
          </div>
        </div>

        {/* Error display */}
        {error && (
          <div className="px-6 py-2 border-t border-border">
            <p className="text-red-400 text-xs font-mono">{error}</p>
          </div>
        )}

        <Toast message="Forged ✓" visible={toastVisible} />
      </div>
    </Reveal>
  )
}
