import { useState, useRef, useCallback, useEffect } from "react"
import { AnimatePresence, motion } from "motion/react"
import { explain } from "../lib/api"
import type { ExplainResponse } from "../lib/types"
import { trackExplainerUsed } from "../lib/analytics"
import { ExplainResult } from "./ExplainResult"

interface ExplainerBarProps {
  onForgeAnother?: (generator: string) => void
}

export function ExplainerBar({ onForgeAnother }: ExplainerBarProps) {
  const [input, setInput] = useState("")
  const [result, setResult] = useState<ExplainResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const abortRef = useRef<AbortController | null>(null)

  const doExplain = useCallback(async (value: string) => {
    // Cancel any in-flight request.
    if (abortRef.current) {
      abortRef.current.abort()
    }

    if (!value.trim()) {
      setResult(null)
      setLoading(false)
      return
    }

    setLoading(true)
    const ctrl = new AbortController()
    abortRef.current = ctrl

    try {
      const res = await explain(value)
      // Only update if this request wasn't aborted.
      if (!ctrl.signal.aborted) {
        setResult(res)
        if (res) trackExplainerUsed()
      }
    } catch {
      if (!ctrl.signal.aborted) {
        setResult(null)
      }
    } finally {
      if (!ctrl.signal.aborted) {
        setLoading(false)
      }
    }
  }, [])

  const handleChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value
      setInput(value)

      if (debounceRef.current) {
        clearTimeout(debounceRef.current)
      }

      if (!value.trim()) {
        setResult(null)
        setLoading(false)
        return
      }

      debounceRef.current = setTimeout(() => {
        doExplain(value)
      }, 200)
    },
    [doExplain]
  )

  // Cleanup on unmount.
  useEffect(() => {
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current)
      if (abortRef.current) abortRef.current.abort()
    }
  }, [])

  return (
    <div className="w-full border-b border-border bg-panel">
      <div className="px-6 py-4">
        <input
          type="text"
          value={input}
          onChange={handleChange}
          placeholder="Paste an ID, key, or token to decode it..."
          className="w-full bg-panel border border-border rounded-md px-6 py-4 font-mono text-lg text-foreground placeholder:text-muted-foreground focus:outline-none focus:border-forge transition-colors duration-150"
        />
      </div>
      <AnimatePresence mode="wait">
        {loading && !result && (
          <motion.div
            key="loading"
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: "auto" }}
            exit={{ opacity: 0, height: 0 }}
            transition={{ duration: 0.15 }}
            className="overflow-hidden"
          >
            <p className="px-6 pb-4 text-muted-foreground text-sm">Analyzing...</p>
          </motion.div>
        )}
        {result && (
          <motion.div
            key={input}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.25 }}
          >
            <ExplainResult result={result} input={input} onForgeAnother={onForgeAnother} />
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}
