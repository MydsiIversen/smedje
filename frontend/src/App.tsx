import { useState, useEffect } from "react"
import { TopBar } from "./components/TopBar"
import { Sidebar } from "./components/Sidebar"
import { fetchVersion } from "./lib/api"

function App() {
  const [selected, setSelected] = useState<string | null>(null)
  const [_paletteOpen, setPaletteOpen] = useState(false)
  const [version, setVersion] = useState<string>("")

  useEffect(() => {
    fetchVersion()
      .then((v) => setVersion(v.version))
      .catch(() => {})
  }, [])

  return (
    <div className="min-h-screen bg-anvil text-foreground flex flex-col">
      <TopBar onPaletteOpen={() => setPaletteOpen(true)} version={version} />
      <div className="flex flex-1" style={{ marginTop: 40 }}>
        <Sidebar selected={selected} onSelect={setSelected} />
        <main className="flex-1 ml-[220px]">
          {selected ? (
            <div className="p-6">
              <p className="font-mono text-lg">{selected}</p>
              <p className="text-muted-foreground text-sm mt-1">
                Generator panel will go here
              </p>
            </div>
          ) : (
            <div className="flex items-center justify-center h-full">
              <div className="text-center">
                <p className="text-muted-foreground text-sm">
                  Select a generator from the sidebar
                </p>
                <p className="text-muted-foreground text-xs mt-1">
                  or press ⌘K to search
                </p>
              </div>
            </div>
          )}
        </main>
      </div>
    </div>
  )
}

export default App
