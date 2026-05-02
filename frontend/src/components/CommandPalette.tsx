import { useState, useEffect, useRef, useCallback } from "react"
import { Command } from "cmdk"
import type { GeneratorInfo } from "../lib/types"
import { generateSingle } from "../lib/api"
import { parseQuery, matchGenerators } from "../lib/palette-parser"
import { getRecents, addRecent } from "../lib/recents"
import type { RecentItem } from "../lib/recents"

interface CommandPaletteProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  generators: GeneratorInfo[]
  onSelectGenerator: (address: string) => void
  onExplain: (value: string) => void
  onRecommend: (topic: string) => void
}

const SUGGESTED = [
  "uuid.v7",
  "password",
  "ssh.ed25519",
  "ulid",
  "nanoid",
  "snowflake",
]

function relativeTime(ts: number): string {
  const diff = Date.now() - ts
  const seconds = Math.floor(diff / 1000)
  if (seconds < 60) return "just now"
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}

export function CommandPalette({
  open,
  onOpenChange,
  generators,
  onSelectGenerator,
  onExplain,
  onRecommend,
}: CommandPaletteProps) {
  const [query, setQuery] = useState("")
  const [highlightedAddress, setHighlightedAddress] = useState("")
  const [preview, setPreview] = useState<string | null>(null)
  const [recents, setRecents] = useState<RecentItem[]>([])
  const previewTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const previewAbortRef = useRef<AbortController | null>(null)

  // Load recents when palette opens.
  useEffect(() => {
    if (open) {
      setRecents(getRecents().slice(0, 8))
      setQuery("")
      setPreview(null)
      setHighlightedAddress("")
    }
  }, [open])

  // Debounced live preview when a generator row is highlighted.
  useEffect(() => {
    if (previewTimerRef.current) {
      clearTimeout(previewTimerRef.current)
    }
    if (previewAbortRef.current) {
      previewAbortRef.current.abort()
      previewAbortRef.current = null
    }
    setPreview(null)

    if (!highlightedAddress || !open) return

    // Only preview if the highlighted address matches a known generator.
    const gen = generators.find((g) => g.address === highlightedAddress)
    if (!gen) return

    const abort = new AbortController()
    previewAbortRef.current = abort

    previewTimerRef.current = setTimeout(() => {
      generateSingle(gen.address, {}, "text", "")
        .then((result) => {
          if (!abort.signal.aborted) {
            setPreview(result.value)
          }
        })
        .catch(() => {
          // Preview is best-effort; ignore failures.
        })
    }, 200)

    return () => {
      if (previewTimerRef.current) {
        clearTimeout(previewTimerRef.current)
      }
      abort.abort()
    }
  }, [highlightedAddress, open, generators])

  const handleSelect = useCallback(
    async (address: string) => {
      try {
        const result = await generateSingle(address, {}, "text", "")
        await navigator.clipboard.writeText(result.value)
        addRecent(address, result.value)
      } catch {
        // Generation or copy failed silently.
      }
      onOpenChange(false)
    },
    [onOpenChange],
  )

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Tab") {
        e.preventDefault()
        if (highlightedAddress) {
          const gen = generators.find((g) => g.address === highlightedAddress)
          if (gen) {
            onSelectGenerator(gen.address)
            onOpenChange(false)
          }
        }
        return
      }

      if (e.key === "Enter") {
        const parsed = parseQuery(query, generators)
        if (parsed.isExplain && parsed.explainValue) {
          e.preventDefault()
          onExplain(parsed.explainValue)
          onOpenChange(false)
          return
        }
        if (parsed.isRecommend && parsed.recommendTopic) {
          e.preventDefault()
          onRecommend(parsed.recommendTopic)
          onOpenChange(false)
          return
        }
        // Otherwise let cmdk handle the select via onSelect on Item.
      }
    },
    [
      highlightedAddress,
      generators,
      query,
      onSelectGenerator,
      onExplain,
      onRecommend,
      onOpenChange,
    ],
  )

  const parsed = parseQuery(query, generators)
  const matched = query.trim()
    ? matchGenerators(query, generators)
    : []

  // When there's a parsed generator from alias/multi-word, ensure it shows.
  const showParsed =
    parsed.generator && !matched.find((g) => g.address === parsed.generator)
  const parsedGen = showParsed
    ? generators.find((g) => g.address === parsed.generator)
    : null

  const suggestedGenerators = generators.filter((g) =>
    SUGGESTED.includes(g.address),
  )

  if (!open) return null

  return (
    <Command.Dialog
      open={open}
      onOpenChange={onOpenChange}
      label="Smedje command palette"
      shouldFilter={false}
      loop
      onValueChange={setHighlightedAddress}
      overlayClassName="fixed inset-0 bg-black/50 z-[100]"
      contentClassName="fixed top-[20%] left-1/2 -translate-x-1/2 w-[60vw] max-w-[720px] bg-panel border border-border rounded-none z-[101] shadow-2xl"
    >
      <div onKeyDown={handleKeyDown}>
        <Command.Input
          value={query}
          onValueChange={setQuery}
          placeholder="Type a generator, alias, or command..."
          className="w-full px-4 py-3 bg-transparent text-foreground text-sm font-mono border-b border-border outline-none placeholder:text-muted-foreground"
          autoFocus
        />

        <Command.List className="max-h-[400px] overflow-y-auto p-2">
          {/* When query is typed and we have matches */}
          {query.trim() && (
            <>
              {parsedGen && (
                <Command.Item
                  key={parsedGen.address}
                  value={parsedGen.address}
                  onSelect={() => handleSelect(parsedGen.address)}
                  className="flex items-center justify-between px-3 py-2 cursor-pointer data-[selected=true]:bg-[#1A1C22] transition-colors duration-100"
                >
                  <div>
                    <span className="font-mono text-sm text-foreground">
                      {parsedGen.address}
                    </span>
                    <span className="text-muted-foreground text-xs ml-3">
                      {parsedGen.description}
                    </span>
                  </div>
                  {highlightedAddress === parsedGen.address && preview && (
                    <span className="font-mono text-xs text-muted-foreground truncate max-w-[200px]">
                      {preview}
                    </span>
                  )}
                </Command.Item>
              )}

              {matched.map((gen) => (
                <Command.Item
                  key={gen.address}
                  value={gen.address}
                  onSelect={() => handleSelect(gen.address)}
                  className="flex items-center justify-between px-3 py-2 cursor-pointer data-[selected=true]:bg-[#1A1C22] transition-colors duration-100"
                >
                  <div>
                    <span className="font-mono text-sm text-foreground">
                      {gen.address}
                    </span>
                    <span className="text-muted-foreground text-xs ml-3">
                      {gen.description}
                    </span>
                  </div>
                  {highlightedAddress === gen.address && preview && (
                    <span className="font-mono text-xs text-muted-foreground truncate max-w-[200px]">
                      {preview}
                    </span>
                  )}
                </Command.Item>
              ))}

              {parsed.isExplain && parsed.explainValue && (
                <Command.Item
                  value="explain-action"
                  onSelect={() => {
                    if (parsed.explainValue) {
                      onExplain(parsed.explainValue)
                      onOpenChange(false)
                    }
                  }}
                  className="px-3 py-2 cursor-pointer data-[selected=true]:bg-[#1A1C22] transition-colors duration-100"
                >
                  <span className="font-mono text-sm text-forge">explain</span>
                  <span className="text-muted-foreground text-xs ml-3">
                    Analyze "{parsed.explainValue}"
                  </span>
                </Command.Item>
              )}

              {parsed.isRecommend && parsed.recommendTopic && (
                <Command.Item
                  value="recommend-action"
                  onSelect={() => {
                    if (parsed.recommendTopic) {
                      onRecommend(parsed.recommendTopic)
                      onOpenChange(false)
                    }
                  }}
                  className="px-3 py-2 cursor-pointer data-[selected=true]:bg-[#1A1C22] transition-colors duration-100"
                >
                  <span className="font-mono text-sm text-forge">
                    recommend
                  </span>
                  <span className="text-muted-foreground text-xs ml-3">
                    Suggest generators for "{parsed.recommendTopic}"
                  </span>
                </Command.Item>
              )}
            </>
          )}

          {/* Empty state: no query */}
          {!query.trim() && (
            <>
              {recents.length > 0 && (
                <Command.Group
                  heading={
                    <span className="text-xs text-muted-foreground uppercase tracking-widest px-1">
                      Recent
                    </span>
                  }
                >
                  {recents.map((item, i) => (
                    <Command.Item
                      key={`recent-${i}`}
                      value={`recent-${item.generator}-${i}`}
                      onSelect={() => handleSelect(item.generator)}
                      className="flex items-center justify-between px-3 py-2 cursor-pointer data-[selected=true]:bg-[#1A1C22] transition-colors duration-100"
                    >
                      <div className="flex items-center gap-3 min-w-0">
                        <span className="font-mono text-sm text-foreground">
                          {item.generator}
                        </span>
                        <span className="font-mono text-xs text-muted-foreground truncate max-w-[200px]">
                          {item.value}
                        </span>
                      </div>
                      <span className="text-xs text-muted-foreground shrink-0 ml-2">
                        {relativeTime(item.timestamp)}
                      </span>
                    </Command.Item>
                  ))}
                </Command.Group>
              )}

              <Command.Group
                heading={
                  <span className="text-xs text-muted-foreground uppercase tracking-widest px-1">
                    Suggested
                  </span>
                }
              >
                {suggestedGenerators.map((gen) => (
                  <Command.Item
                    key={gen.address}
                    value={gen.address}
                    onSelect={() => handleSelect(gen.address)}
                    className="flex items-center justify-between px-3 py-2 cursor-pointer data-[selected=true]:bg-[#1A1C22] transition-colors duration-100"
                  >
                    <div>
                      <span className="font-mono text-sm text-foreground">
                        {gen.address}
                      </span>
                      <span className="text-muted-foreground text-xs ml-3">
                        {gen.description}
                      </span>
                    </div>
                    {highlightedAddress === gen.address && preview && (
                      <span className="font-mono text-xs text-muted-foreground truncate max-w-[200px]">
                        {preview}
                      </span>
                    )}
                  </Command.Item>
                ))}
              </Command.Group>
            </>
          )}

          {/* No matches found for typed query */}
          {query.trim() &&
            matched.length === 0 &&
            !parsedGen &&
            !parsed.isExplain &&
            !parsed.isRecommend && (
              <Command.Empty className="px-3 py-6 text-center text-muted-foreground text-sm">
                No generators match "{query}"
              </Command.Empty>
            )}
        </Command.List>

        {/* Footer with keyboard hints */}
        <div className="border-t border-border px-4 py-2 flex items-center gap-4">
          <span className="text-xs text-muted-foreground">
            <kbd className="font-mono">&#x21B5;</kbd> generate + copy
          </span>
          <span className="text-xs text-muted-foreground">
            <kbd className="font-mono">&#x21E5;</kbd> open panel
          </span>
          <span className="text-xs text-muted-foreground">
            <kbd className="font-mono">esc</kbd> close
          </span>
        </div>
      </div>
    </Command.Dialog>
  )
}
