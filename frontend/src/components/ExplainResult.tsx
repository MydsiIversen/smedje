import { useState } from "react"
import { motion } from "motion/react"
import type { Variants } from "motion/react"
import type { ExplainResponse, LayoutSegment } from "../lib/types"
import { Stagger } from "./primitives"

const colorMap: Record<string, string> = {
  time: "125, 211, 252",
  random: "167, 139, 250",
  version: "251, 191, 36",
  counter: "52, 211, 153",
  type: "244, 114, 182",
  meta: "148, 163, 184",
}

const cssVarMap: Record<string, string> = {
  time: "var(--color-syntax-time)",
  random: "var(--color-syntax-random)",
  version: "var(--color-syntax-version)",
  counter: "var(--color-syntax-counter)",
  type: "var(--color-syntax-type)",
  meta: "var(--color-syntax-meta)",
}

// Map detected format names to generator addresses.
function formatToGenerator(detected: string): string | null {
  const lower = detected.toLowerCase()
  if (lower.includes("uuidv7") || lower.includes("uuid v7")) return "uuid.v7"
  if (lower.includes("uuidv4") || lower.includes("uuid v4")) return "uuid.v4"
  if (lower.includes("uuidv1") || lower.includes("uuid v1")) return "uuid.v1"
  if (lower.includes("uuidv6") || lower.includes("uuid v6")) return "uuid.v6"
  if (lower.includes("uuidv8") || lower.includes("uuid v8")) return "uuid.v8"
  if (lower.includes("uuid")) return "uuid.v4"
  if (lower.includes("ulid")) return "ulid"
  if (lower.includes("snowflake")) return "snowflake"
  if (lower.includes("nanoid")) return "nanoid"
  return null
}

const sectionVariant: Variants = {
  hidden: { opacity: 0, y: 8 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.25, ease: "easeOut" as const },
  },
}

interface ExplainResultProps {
  result: ExplainResponse
  input: string
  onForgeAnother?: (generator: string) => void
}

// Build an array of spans covering the full input string, with colored
// segments from layout and unstyled spans for gaps.
function buildSegmentedSpans(input: string, layout: LayoutSegment[]) {
  const sorted = [...layout].sort((a, b) => a.start - b.start)
  const spans: { text: string; segment?: LayoutSegment }[] = []
  let cursor = 0

  for (const seg of sorted) {
    if (seg.start > cursor) {
      spans.push({ text: input.slice(cursor, seg.start) })
    }
    spans.push({ text: input.slice(seg.start, seg.end), segment: seg })
    cursor = seg.end
  }

  if (cursor < input.length) {
    spans.push({ text: input.slice(cursor) })
  }

  return spans
}

export function ExplainResult({ result, input, onForgeAnother }: ExplainResultProps) {
  const [altOpen, setAltOpen] = useState(false)

  if (result.detected === "unknown") {
    return (
      <div className="px-6 py-4">
        <p className="text-muted-foreground text-sm">Format not recognized</p>
        <p className="text-muted-foreground text-xs mt-1">
          Supported formats: UUID (v1, v4, v6, v7, v8), ULID, Snowflake, NanoID
        </p>
      </div>
    )
  }

  const spans = buildSegmentedSpans(input, result.layout)
  const generator = formatToGenerator(result.detected)

  // Filter out "format" from fields since it's shown in the header.
  const fieldEntries = Object.entries(result.fields).filter(
    ([key]) => key.toLowerCase() !== "format"
  )

  const hasAlternateForms =
    result.alternateForms && Object.keys(result.alternateForms).length > 0

  return (
    <Stagger className="px-6 py-4 space-y-4">
      {/* Header: detected format + spec link */}
      <motion.div variants={sectionVariant} className="flex items-baseline gap-2">
        <span className="font-mono text-lg text-foreground">{result.detected}</span>
        {result.spec && (
          <a
            href={result.spec}
            target="_blank"
            rel="noopener noreferrer"
            className="text-muted-foreground hover:text-forge text-xs transition-colors duration-150 flex items-center gap-1"
          >
            spec
            <svg
              viewBox="0 0 16 16"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
              className="w-3 h-3"
            >
              <path d="M6 3h7v7M13 3L6 10" />
            </svg>
          </a>
        )}
      </motion.div>

      {/* Visual layout: colored spans */}
      {spans.length > 0 && (
        <motion.div
          variants={sectionVariant}
          className="font-mono text-base tracking-[0.05em] flex flex-wrap"
        >
          {spans.map((s, i) => {
            if (!s.segment) {
              return <span key={i}>{s.text}</span>
            }
            const seg = s.segment
            const rgb = colorMap[seg.type] || colorMap.meta
            const cssColor = cssVarMap[seg.type] || cssVarMap.meta
            const tooltipParts = [seg.label]
            if (seg.description) tooltipParts.push(seg.description)
            if (seg.value) tooltipParts.push(seg.value)
            const tooltip = tooltipParts.join(" — ")

            return (
              <span
                key={i}
                title={tooltip}
                className="cursor-default"
                style={{
                  backgroundColor: `rgba(${rgb}, 0.2)`,
                  color: cssColor,
                  transition: "background-color 100ms, transform 100ms",
                  display: "inline-block",
                  padding: "2px 1px",
                  borderRadius: 2,
                }}
                onMouseEnter={(e) => {
                  const el = e.currentTarget
                  el.style.backgroundColor = `rgba(${rgb}, 0.35)`
                  el.style.transform = "scale(1.02)"
                }}
                onMouseLeave={(e) => {
                  const el = e.currentTarget
                  el.style.backgroundColor = `rgba(${rgb}, 0.2)`
                  el.style.transform = "scale(1)"
                }}
              >
                {s.text}
              </span>
            )
          })}
        </motion.div>
      )}

      {/* Fields table */}
      {fieldEntries.length > 0 && (
        <motion.div variants={sectionVariant}>
          {fieldEntries.map(([key, value], i) => (
            <div
              key={key}
              className={`flex py-1 ${i < fieldEntries.length - 1 ? "border-b border-border" : ""}`}
            >
              <span className="text-muted-foreground text-sm w-40 shrink-0">{key}</span>
              <span className="font-mono text-sm break-all">{value}</span>
            </div>
          ))}
        </motion.div>
      )}

      {/* Alternate forms */}
      {hasAlternateForms && (
        <motion.div variants={sectionVariant}>
          <button
            onClick={() => setAltOpen(!altOpen)}
            className="text-muted-foreground text-sm hover:text-foreground transition-colors duration-150"
          >
            Alternate forms {altOpen ? "▾" : "▸"}
          </button>
          {altOpen && (
            <div className="mt-2 space-y-1">
              {Object.entries(result.alternateForms!).map(([label, value]) => (
                <div key={label} className="flex gap-3">
                  <span className="text-muted-foreground text-sm w-40 shrink-0">
                    {label}
                  </span>
                  <span className="font-mono text-sm break-all">{value}</span>
                </div>
              ))}
            </div>
          )}
        </motion.div>
      )}

      {/* Actions */}
      {generator && onForgeAnother && (
        <motion.div variants={sectionVariant}>
          <button
            onClick={() => onForgeAnother(generator)}
            className="text-sm text-forge hover:text-forge-dim transition-colors duration-150 font-mono"
          >
            Forge another like this
          </button>
        </motion.div>
      )}
    </Stagger>
  )
}
