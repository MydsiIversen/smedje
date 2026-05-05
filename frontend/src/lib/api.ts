import type {
  GeneratorInfo,
  GeneratorSchema,
  ExplainResponse,
  Recommendation,
  VersionInfo,
  SSEEvent,
} from "./types"

const BASE = ""

export async function fetchGenerators(): Promise<GeneratorInfo[]> {
  const res = await fetch(`${BASE}/api/generators`)
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return res.json()
}

export async function fetchGeneratorSchema(address: string): Promise<GeneratorSchema> {
  const res = await fetch(`${BASE}/api/generators/${address}`)
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return res.json()
}

export async function explain(value: string): Promise<ExplainResponse | null> {
  const res = await fetch(`${BASE}/api/explain`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ value }),
  })
  if (!res.ok) return null
  return res.json()
}

export async function fetchRecommendations(
  topic: string,
  useCase?: string,
): Promise<Recommendation[]> {
  const params = new URLSearchParams({ topic })
  if (useCase) params.set("use-case", useCase)
  const res = await fetch(`${BASE}/api/recommend?${params}`)
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return res.json()
}

export async function fetchVersion(): Promise<VersionInfo> {
  const res = await fetch(`${BASE}/api/version`)
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return res.json()
}

export function generateSSE(
  generator: string,
  count: number,
  params: Record<string, string>,
  format: string,
  seed: string,
  onEvent: (event: SSEEvent) => void,
): AbortController {
  const ctrl = new AbortController()

  fetch(`${BASE}/api/generate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ generator, count, format, params, seed: seed || undefined }),
    signal: ctrl.signal,
  })
    .then(async (res) => {
      if (!res.ok || !res.body) {
        onEvent({ type: "error", message: `HTTP ${res.status}` })
        return
      }
      const reader = res.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ""

      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })

        const lines = buffer.split("\n")
        buffer = lines.pop() || ""

        let currentEvent = ""
        for (const line of lines) {
          if (line.startsWith("event: ")) {
            currentEvent = line.slice(7).trim()
          } else if (line.startsWith("data: ") && currentEvent) {
            try {
              const data = JSON.parse(line.slice(6))
              onEvent({ type: currentEvent, ...data } as SSEEvent)
            } catch {
              // Skip malformed data lines.
            }
            currentEvent = ""
          }
        }
      }
    })
    .catch((err: Error) => {
      if (err.name !== "AbortError") {
        onEvent({ type: "error", message: err.message })
      }
    })

  return ctrl
}

export async function generateSingle(
  generator: string,
  params: Record<string, string>,
  format: string,
  seed: string,
): Promise<{ value: string; fields: Record<string, string>; sensitiveKeys?: string[]; formatted?: string }> {
  const res = await fetch(`${BASE}/api/generate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ generator, count: 1, format, params, seed: seed || undefined }),
  })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return res.json()
}
