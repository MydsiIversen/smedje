import { useEffect, useState } from "react"
import { fetchGenerators } from "../lib/api"
import type { GeneratorInfo } from "../lib/types"

interface SidebarProps {
  selected: string | null
  onSelect: (address: string) => void
}

interface CategoryGroup {
  label: string
  generators: GeneratorInfo[]
}

function groupGenerators(generators: GeneratorInfo[]): CategoryGroup[] {
  const groups: CategoryGroup[] = [
    { label: "Identifiers", generators: [] },
    { label: "Crypto Keys", generators: [] },
    { label: "Certificates", generators: [] },
    { label: "Secrets", generators: [] },
    { label: "Network", generators: [] },
  ]

  for (const gen of generators) {
    if (gen.category === "id") {
      groups[0].generators.push(gen)
    } else if (gen.category === "crypto" && gen.group !== "tls") {
      groups[1].generators.push(gen)
    } else if (gen.category === "crypto" && gen.group === "tls") {
      groups[2].generators.push(gen)
    } else if (gen.category === "secret") {
      groups[3].generators.push(gen)
    } else if (gen.category === "network") {
      groups[4].generators.push(gen)
    }
  }

  for (const group of groups) {
    group.generators.sort((a, b) => a.address.localeCompare(b.address))
  }

  return groups.filter((g) => g.generators.length > 0)
}

export function Sidebar({ selected, onSelect }: SidebarProps) {
  const [groups, setGroups] = useState<CategoryGroup[]>([])

  useEffect(() => {
    fetchGenerators()
      .then((gens) => setGroups(groupGenerators(gens)))
      .catch(() => {})
  }, [])

  return (
    <aside className="fixed top-10 left-0 bottom-0 w-[220px] bg-panel border-r border-border overflow-y-auto">
      <nav className="py-2">
        {groups.map((group) => (
          <div key={group.label}>
            <h3 className="uppercase text-xs text-muted-foreground tracking-widest px-4 py-2">
              {group.label}
            </h3>
            {group.generators.map((gen) => {
              const isSelected = selected === gen.address
              return (
                <button
                  key={gen.address}
                  onClick={() => onSelect(gen.address)}
                  className={`w-full text-left px-4 py-1.5 font-mono text-sm transition-colors duration-150 ${
                    isSelected
                      ? "border-l-2 border-forge text-foreground bg-[#1A1C22]"
                      : "border-l-2 border-transparent text-muted-foreground hover:bg-[#1A1C22] hover:text-foreground"
                  }`}
                >
                  {gen.address}
                </button>
              )
            })}
          </div>
        ))}
      </nav>
    </aside>
  )
}
