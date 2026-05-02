const STORAGE_KEY = "smedje-recents"
const MAX_ITEMS = 50

export interface RecentItem {
  generator: string
  value: string
  timestamp: number
}

/** Retrieve the list of recently generated values from localStorage. */
export function getRecents(): RecentItem[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return []
    return JSON.parse(raw) as RecentItem[]
  } catch {
    return []
  }
}

/** Add a generated value to the recents list, deduplicating by generator+value. */
export function addRecent(generator: string, value: string): void {
  const items = getRecents()
  const filtered = items.filter(
    (i) => !(i.generator === generator && i.value === value),
  )
  filtered.unshift({ generator, value, timestamp: Date.now() })
  const trimmed = filtered.slice(0, MAX_ITEMS)
  localStorage.setItem(STORAGE_KEY, JSON.stringify(trimmed))
}

/** Clear all recent items from localStorage. */
export function clearRecents(): void {
  localStorage.removeItem(STORAGE_KEY)
}
