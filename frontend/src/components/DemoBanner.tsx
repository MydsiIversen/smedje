import { useState, useEffect } from "react"
import { Copy, Check, X } from "lucide-react"
import { trackInstallLinkClicked } from "../lib/analytics"

const STORAGE_KEY = "smedje-banner-dismissed"
const INSTALL_CMD = "go install github.com/MydsiIversen/smedje/cmd/smedje@latest"

interface DemoBannerProps {
  visible: boolean
  onDismiss: () => void
}

export function DemoBanner({ visible, onDismiss }: DemoBannerProps) {
  const [copied, setCopied] = useState(false)
  const [dismissed, setDismissed] = useState(false)

  useEffect(() => {
    try {
      if (localStorage.getItem(STORAGE_KEY) === "true") {
        setDismissed(true)
      }
    } catch {
      // localStorage unavailable.
    }
  }, [])

  if (!visible || dismissed) return null

  function handleCopy() {
    navigator.clipboard.writeText(INSTALL_CMD)
    setCopied(true)
    trackInstallLinkClicked()
    setTimeout(() => setCopied(false), 1500)
  }

  function handleDismiss() {
    try {
      localStorage.setItem(STORAGE_KEY, "true")
    } catch {
      // localStorage unavailable.
    }
    setDismissed(true)
    onDismiss()
  }

  return (
    <div className="h-8 bg-panel border-b border-border border-l-2 border-l-forge flex items-center px-4 text-xs text-muted-foreground shrink-0">
      <span>
        You're using the public demo. Install locally for unlimited use:{" "}
        <code className="font-mono text-foreground">{INSTALL_CMD}</code>
      </span>
      <button
        onClick={handleCopy}
        className="ml-2 text-muted-foreground hover:text-foreground transition-colors duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-forge rounded-[4px] p-0.5"
        aria-label="Copy install command"
      >
        {copied ? <Check className="w-3 h-3" /> : <Copy className="w-3 h-3" />}
      </button>
      <div className="flex-1" />
      <button
        onClick={handleDismiss}
        className="text-muted-foreground hover:text-foreground transition-colors duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-forge rounded-[4px] p-0.5"
        aria-label="Dismiss banner"
      >
        <X className="w-3 h-3" />
      </button>
    </div>
  )
}
