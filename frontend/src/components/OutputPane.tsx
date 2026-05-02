import type { Artifact, GeneratorSchema } from "../lib/types"
import { Stamp, Stagger, TypewriterText, ProgressBar } from "./primitives"
import { ArtifactCard } from "./ArtifactCard"
import { Copy, Check } from "lucide-react"
import { useState } from "react"

interface OutputPaneProps {
  preview: Artifact | null
  results: Artifact[]
  forging: boolean
  progress: { current: number; total: number; opsPerSec: number } | null
  status: string
  schema: GeneratorSchema
  onCopyAll: () => void
  durationMs: number | null
  formatted?: string | null
}

/** True if the artifact has multiple named fields beyond just "value". */
function isMultiField(artifact: Artifact): boolean {
  const keys = Object.keys(artifact.fields)
  return keys.length > 1 || (keys.length === 1 && !artifact.fields["value"])
}

export function OutputPane({
  preview,
  results,
  forging,
  progress,
  status,
  onCopyAll,
  durationMs,
  formatted,
}: OutputPaneProps) {
  const [copied, setCopied] = useState(false)
  const showPreview = !forging && results.length === 0 && preview !== null
  const showResults = results.length > 0
  const showProgress = forging && progress !== null

  function handleCopyAll() {
    onCopyAll()
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }

  return (
    <div className="flex flex-col h-full gap-3">
      {/* Status line */}
      {status && (
        <div className="text-xs text-muted-foreground font-mono" aria-live="polite">
          <TypewriterText text={status} speed={10} />
        </div>
      )}

      {/* Progress bar */}
      {showProgress && progress && (
        <div
          className="space-y-1"
          role="progressbar"
          aria-valuenow={progress.current}
          aria-valuemax={progress.total}
        >
          <ProgressBar progress={progress.current / progress.total} />
          <div className="flex justify-between text-xs text-muted-foreground font-mono">
            <span>
              {progress.current} / {progress.total}
            </span>
            <span>{Math.round(progress.opsPerSec).toLocaleString()} ops/s</span>
          </div>
        </div>
      )}

      {/* Preview mode */}
      {showPreview && preview && (
        <div className="flex-1 overflow-y-auto">
          {isMultiField(preview) ? (
            <div className="space-y-2">
              {Object.entries(preview.fields).map(([key, val]) => (
                <ArtifactCard key={key} label={key} value={val} />
              ))}
            </div>
          ) : (
            <div>
              <span className="text-xs text-muted-foreground">(preview)</span>
              <p className="font-mono text-lg text-foreground mt-1 break-all">
                {preview.value}
              </p>
            </div>
          )}
        </div>
      )}

      {/* Formatted batch output (SQL, CSV, JSON, env) */}
      {showResults && formatted && !forging && (
        <div className="flex-1 overflow-y-auto">
          <pre className="font-mono text-sm text-foreground whitespace-pre-wrap break-all m-0">
            {formatted}
          </pre>
        </div>
      )}

      {/* Forge results (individual artifacts) */}
      {showResults && !formatted && (
        <div className="flex-1 overflow-y-auto space-y-2">
          <Stagger>
            {results.map((artifact, i) => (
              <Stamp key={i}>
                {isMultiField(artifact) ? (
                  <div className="space-y-2 mb-3">
                    {Object.entries(artifact.fields).map(([key, val]) => (
                      <ArtifactCard key={key} label={key} value={val} />
                    ))}
                  </div>
                ) : (
                  <p className="font-mono text-sm text-foreground py-0.5 break-all">
                    {artifact.value}
                  </p>
                )}
              </Stamp>
            ))}
          </Stagger>
        </div>
      )}

      {/* Empty state */}
      {!showPreview && !showResults && !forging && (
        <div className="flex-1 flex items-center justify-center">
          <p className="text-muted-foreground text-sm">Press Forge to generate</p>
        </div>
      )}

      {/* Footer after forge complete */}
      {!forging && results.length > 0 && (
        <div className="flex items-center justify-between border-t border-border pt-3">
          <span className="text-xs text-muted-foreground font-mono">
            {results.length} item{results.length !== 1 ? "s" : ""}
            {durationMs !== null && ` in ${durationMs}ms`}
          </span>
          <button
            onClick={handleCopyAll}
            className="flex items-center gap-1.5 text-xs font-mono text-muted-foreground hover:text-foreground border border-border px-3 py-1 rounded-[4px] transition-colors duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-forge"
          >
            {copied ? <Check className="w-3 h-3" /> : <Copy className="w-3 h-3" />}
            {copied ? "Copied" : "Copy All"}
          </button>
        </div>
      )}
    </div>
  )
}
