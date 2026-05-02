import { useState } from "react"
import { Copy, Check } from "lucide-react"

interface ArtifactCardProps {
  label: string
  value: string
  sensitive?: boolean
}

export function ArtifactCard({ label, value, sensitive }: ArtifactCardProps) {
  const [copied, setCopied] = useState(false)
  const [revealed, setRevealed] = useState(!sensitive)

  function handleCopy() {
    navigator.clipboard.writeText(value).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    })
  }

  return (
    <div className="bg-panel border border-border p-4 rounded-none relative group">
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs text-muted-foreground uppercase tracking-wide">
          {label}
        </span>
        <button
          onClick={handleCopy}
          className="opacity-0 group-hover:opacity-100 transition-opacity duration-150 text-muted-foreground hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-forge rounded-[4px] p-1"
          aria-label={`Copy ${label}`}
        >
          {copied ? <Check className="w-3.5 h-3.5" /> : <Copy className="w-3.5 h-3.5" />}
        </button>
      </div>
      {sensitive && !revealed ? (
        <button
          onClick={() => setRevealed(true)}
          className="font-mono text-sm text-muted-foreground hover:text-foreground transition-colors duration-150"
        >
          Click to reveal
        </button>
      ) : (
        <pre className="font-mono text-sm whitespace-pre-wrap break-all text-foreground m-0">
          {value}
        </pre>
      )}
    </div>
  )
}
