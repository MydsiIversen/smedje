function AnvilMark({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" className={className}>
      <path d="M4 8h16v2H4zm2 2h12v3H6zm3 3h6v2H9zm1 2h4v3h-4zm-6-7l2-4h12l2 4" />
    </svg>
  )
}

interface TopBarProps {
  onPaletteOpen: () => void
  version?: string
}

export function TopBar({ onPaletteOpen, version }: TopBarProps) {
  return (
    <header className="fixed top-0 left-0 right-0 z-50 h-10 bg-panel border-b border-border flex items-center justify-between px-4">
      <div className="flex items-center gap-2">
        <AnvilMark className="w-6 h-6 text-forge" />
        <span className="font-mono font-semibold tracking-tight text-foreground">
          smedje
        </span>
      </div>

      <div className="flex items-center gap-3">
        <button
          onClick={onPaletteOpen}
          aria-label="Open command palette"
          className="flex items-center gap-1 text-muted-foreground text-xs hover:text-foreground transition-colors duration-150"
        >
          <kbd className="px-1.5 py-0.5 border border-border text-xs font-mono rounded-[4px]">
            ⌘K
          </kbd>
          <span>to forge</span>
        </button>

        {version && (
          <a
            href="https://github.com/MydsiIversen/smedje"
            target="_blank"
            rel="noopener noreferrer"
            className="text-muted-foreground text-xs font-mono hover:text-foreground transition-colors duration-150"
          >
            {version}
          </a>
        )}
      </div>
    </header>
  )
}
