import { useState, useEffect, useRef, useCallback } from "react"
import { X } from "lucide-react"
import { fetchRecommendations } from "../lib/api"
import type { Recommendation } from "../lib/types"

interface RecommendModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  initialTopic?: string
  onUseGenerator: (address: string) => void
}

const TOPICS = [
  "id",
  "ssh-key",
  "tls-cert",
  "password",
  "hash",
  "jwt",
  "secret",
  "vpn-key",
]

// Parse a smedje CLI command string into a generator address.
// "smedje uuid v7" -> "uuid.v7"
// "smedje password --length 16" -> "password"
// "smedje ssh ed25519" -> "ssh.ed25519"
function commandToAddress(command: string): string {
  const parts = command.trim().split(/\s+/)
  // Strip leading "smedje" if present.
  const start = parts[0] === "smedje" ? 1 : 0
  const segments: string[] = []
  for (let i = start; i < parts.length; i++) {
    if (parts[i].startsWith("-")) break
    segments.push(parts[i])
  }
  return segments.join(".")
}

export function RecommendModal({
  open,
  onOpenChange,
  initialTopic,
  onUseGenerator,
}: RecommendModalProps) {
  const [topic, setTopic] = useState(initialTopic || TOPICS[0])
  const [useCase, setUseCase] = useState("")
  const [debouncedUseCase, setDebouncedUseCase] = useState("")
  const [recommendations, setRecommendations] = useState<Recommendation[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const overlayRef = useRef<HTMLDivElement>(null)

  // Sync initialTopic when modal opens.
  useEffect(() => {
    if (open && initialTopic) {
      setTopic(initialTopic)
    }
  }, [open, initialTopic])

  // Reset state when modal opens.
  useEffect(() => {
    if (open) {
      setUseCase("")
      setDebouncedUseCase("")
      setError(null)
    }
  }, [open])

  // Debounce the use-case input.
  useEffect(() => {
    if (debounceRef.current) {
      clearTimeout(debounceRef.current)
    }
    debounceRef.current = setTimeout(() => {
      setDebouncedUseCase(useCase)
    }, 300)
    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current)
      }
    }
  }, [useCase])

  // Fetch recommendations when topic or debounced use-case changes.
  useEffect(() => {
    if (!open) return

    let cancelled = false
    setLoading(true)
    setError(null)

    fetchRecommendations(topic, debouncedUseCase || undefined)
      .then((recs) => {
        if (!cancelled) {
          setRecommendations(recs)
        }
      })
      .catch((err: Error) => {
        if (!cancelled) {
          setError(err.message)
          setRecommendations([])
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false)
        }
      })

    return () => {
      cancelled = true
    }
  }, [open, topic, debouncedUseCase])

  // Close on Escape.
  useEffect(() => {
    if (!open) return

    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") {
        onOpenChange(false)
      }
    }
    document.addEventListener("keydown", handleKeyDown)
    return () => document.removeEventListener("keydown", handleKeyDown)
  }, [open, onOpenChange])

  const handleOverlayClick = useCallback(
    (e: React.MouseEvent) => {
      if (e.target === overlayRef.current) {
        onOpenChange(false)
      }
    },
    [onOpenChange],
  )

  const handleUse = useCallback(
    (command: string) => {
      const address = commandToAddress(command)
      onOpenChange(false)
      onUseGenerator(address)
    },
    [onOpenChange, onUseGenerator],
  )

  if (!open) return null

  return (
    <div
      ref={overlayRef}
      className="fixed inset-0 bg-black/50 z-[100] flex items-center justify-center"
      onClick={handleOverlayClick}
    >
      <div className="bg-panel border border-border max-w-[640px] w-[90vw] rounded-none shadow-2xl flex flex-col max-h-[80vh]">
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-border">
          <h2 className="text-foreground text-sm font-medium">
            Recommendations
          </h2>
          <button
            onClick={() => onOpenChange(false)}
            className="text-muted-foreground hover:text-foreground transition-colors"
          >
            <X size={16} />
          </button>
        </div>

        {/* Body */}
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {/* Topic selector */}
          <div className="flex flex-wrap gap-1">
            {TOPICS.map((t) => (
              <button
                key={t}
                onClick={() => setTopic(t)}
                className={`text-xs font-mono px-3 py-1.5 transition-colors ${
                  t === topic
                    ? "bg-forge text-foreground"
                    : "border border-border text-muted-foreground hover:text-foreground"
                }`}
              >
                {t}
              </button>
            ))}
          </div>

          {/* Use-case filter */}
          <input
            type="text"
            value={useCase}
            onChange={(e) => setUseCase(e.target.value)}
            placeholder="Filter by use case..."
            className="w-full bg-transparent border border-border px-3 py-2 text-sm font-mono text-foreground placeholder:text-muted-foreground focus:outline-none focus:border-forge transition-colors"
          />

          {/* Loading */}
          {loading && (
            <p className="text-muted-foreground text-sm">Loading...</p>
          )}

          {/* Error */}
          {error && (
            <p className="text-destructive text-sm">
              Failed to load recommendations: {error}
            </p>
          )}

          {/* Results */}
          {!loading &&
            !error &&
            recommendations.map((rec, i) => (
              <div key={i} className="space-y-3">
                {/* Primary card */}
                <div className="bg-panel border-l-2 border-l-forge p-4 space-y-2">
                  <div className="flex items-start justify-between gap-2">
                    <div className="space-y-1 min-w-0">
                      <p className="font-mono text-base text-foreground">
                        {rec.primary}
                      </p>
                      <p className="text-muted-foreground text-sm">
                        {rec.use_case}
                      </p>
                    </div>
                    <button
                      onClick={() => handleUse(rec.command)}
                      className="text-forge text-sm hover:underline shrink-0"
                    >
                      Use this
                    </button>
                  </div>
                  <p className="text-sm text-foreground">{rec.why}</p>
                  <code className="font-mono text-xs bg-anvil px-2 py-1 rounded inline-block text-foreground">
                    {rec.command}
                  </code>
                </div>

                {/* Alternatives */}
                {rec.alternatives && rec.alternatives.length > 0 && (
                  <div className="pl-3 space-y-1">
                    {rec.alternatives.map((alt, j) => (
                      <div
                        key={j}
                        className="text-sm text-muted-foreground"
                      >
                        <span className="font-mono text-foreground text-xs">
                          {alt.name}
                        </span>
                        <span className="ml-2 text-xs">
                          when {alt.when}
                        </span>
                      </div>
                    ))}
                  </div>
                )}

                {/* Avoid section */}
                {rec.avoid && rec.avoid.length > 0 && (
                  <div className="pl-3">
                    <ul className="text-muted-foreground text-xs list-disc list-inside space-y-0.5">
                      {rec.avoid.map((item, j) => (
                        <li key={j}>{item}</li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>
            ))}

          {/* Empty state */}
          {!loading && !error && recommendations.length === 0 && (
            <p className="text-muted-foreground text-sm text-center py-4">
              No recommendations found for this topic.
            </p>
          )}
        </div>
      </div>
    </div>
  )
}
