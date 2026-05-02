import { useState, useEffect, useCallback } from "react"
import { TopBar } from "./components/TopBar"
import { ExplainerBar } from "./components/ExplainerBar"
import { Sidebar } from "./components/Sidebar"
import { GeneratorPanel } from "./components/GeneratorPanel"
import { CommandPalette } from "./components/CommandPalette"
import { RecommendModal } from "./components/RecommendModal"
import { DemoBanner } from "./components/DemoBanner"
import { PrivacyPage } from "./components/PrivacyPage"
import { fetchVersion, fetchGenerators } from "./lib/api"
import type { GeneratorInfo } from "./lib/types"

function MobileMessage() {
  return (
    <div className="md:hidden flex items-center justify-center min-h-screen bg-anvil text-foreground p-8">
      <div className="text-center max-w-xs">
        <h1 className="font-mono text-xl font-semibold mb-2">smedje</h1>
        <p className="text-muted text-sm mb-4">Smedje is best on desktop.</p>
        <p className="text-muted text-xs">
          Install the CLI for mobile-friendly use:
        </p>
        <code className="block mt-2 text-xs font-mono bg-panel border border-border px-3 py-2">
          go install github.com/MydsiIversen/smedje/cmd/smedje@latest
        </code>
      </div>
    </div>
  )
}

function App() {
  const [selected, setSelected] = useState<string | null>(null)
  const [paletteOpen, setPaletteOpen] = useState(false)
  const [recommendOpen, setRecommendOpen] = useState(false)
  const [recommendTopic, setRecommendTopic] = useState<string | undefined>()
  const [version, setVersion] = useState<string>("")
  const [publicMode, setPublicMode] = useState(false)
  const [generators, setGenerators] = useState<GeneratorInfo[]>([])

  useEffect(() => {
    fetchVersion()
      .then((v) => {
        setVersion(v.version)
        setPublicMode(v.publicMode)
      })
      .catch(() => {})
    fetchGenerators()
      .then(setGenerators)
      .catch(() => {})
  }, [])

  // Simple pathname routing for the privacy page.
  if (window.location.pathname === "/privacy") {
    return <PrivacyPage />
  }

  // Global keyboard shortcuts.
  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      // Cmd+K / Ctrl+K: toggle command palette.
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault()
        setPaletteOpen((prev) => !prev)
        return
      }

      // Cmd+Enter / Ctrl+Enter: forge from the focused panel.
      if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
        e.preventDefault()
        const forgeBtn = document.querySelector<HTMLButtonElement>(
          "[data-forge-button]"
        )
        if (forgeBtn && !forgeBtn.disabled) {
          forgeBtn.click()
        }
        return
      }

      // Cmd+Shift+E / Ctrl+Shift+E: focus explainer input.
      if ((e.metaKey || e.ctrlKey) && e.shiftKey && e.key === "E") {
        e.preventDefault()
        const explainerInput = document.querySelector<HTMLInputElement>(
          "[data-explainer-input]"
        )
        if (explainerInput) {
          explainerInput.focus()
        }
        return
      }

      // Esc: close any open modal.
      if (e.key === "Escape") {
        if (paletteOpen) {
          setPaletteOpen(false)
        } else if (recommendOpen) {
          setRecommendOpen(false)
        }
      }
    }
    document.addEventListener("keydown", handleKeyDown)
    return () => document.removeEventListener("keydown", handleKeyDown)
  }, [paletteOpen, recommendOpen])

  const handleSelectGenerator = useCallback((address: string) => {
    setSelected(address)
  }, [])

  const handleExplain = useCallback((_value: string) => {
    // Explain mode is deferred to a later phase.
  }, [])

  const handleRecommend = useCallback((topic: string) => {
    setRecommendTopic(topic)
    setRecommendOpen(true)
  }, [])

  return (
    <>
      <MobileMessage />
      <div className="hidden md:flex min-h-screen bg-anvil text-foreground flex-col">
        <DemoBanner visible={publicMode} onDismiss={() => {}} />
        <TopBar onPaletteOpen={() => setPaletteOpen(true)} version={version} />
        <div style={{ marginTop: 40 }}>
          <ExplainerBar onForgeAnother={(gen) => setSelected(gen)} />
        </div>
        <div className="flex flex-1">
          <Sidebar selected={selected} onSelect={setSelected} />
          <main className="flex-1 ml-[220px] flex flex-col" style={{ height: "calc(100vh - 40px)" }}>
            {selected ? (
              <GeneratorPanel address={selected} maxCount={publicMode ? 100 : 10000} />
            ) : (
              <div className="flex items-center justify-center h-full">
                <div className="text-center">
                  <p className="text-muted-foreground text-sm">
                    Select a generator from the sidebar
                  </p>
                  <p className="text-muted-foreground text-xs mt-1">
                    or press Cmd+K to search
                  </p>
                </div>
              </div>
            )}
          </main>
        </div>

        <CommandPalette
          open={paletteOpen}
          onOpenChange={setPaletteOpen}
          generators={generators}
          onSelectGenerator={handleSelectGenerator}
          onExplain={handleExplain}
          onRecommend={handleRecommend}
        />

        <RecommendModal
          open={recommendOpen}
          onOpenChange={setRecommendOpen}
          initialTopic={recommendTopic}
          onUseGenerator={handleSelectGenerator}
        />
      </div>
    </>
  )
}

export default App
