import type { GeneratorInfo } from "./types"

export interface ParsedQuery {
  generator: string | null
  count: number
  params: Record<string, string>
  format: string
  isExplain: boolean
  explainValue?: string
  isRecommend: boolean
  recommendTopic?: string
}

const aliases: Record<string, string> = {
  pwd: "password",
  pass: "password",
  wg: "wireguard.keypair",
  ssh: "ssh.ed25519",
  tls: "tls.self-signed",
  cert: "tls.self-signed",
  id: "uuid.v7",
  mac: "mac",
}

/** Parse a command palette query into a structured result. */
export function parseQuery(input: string, generators: GeneratorInfo[]): ParsedQuery {
  const trimmed = input.trim()
  const result: ParsedQuery = {
    generator: null,
    count: 1,
    params: {},
    format: "text",
    isExplain: false,
    isRecommend: false,
  }

  if (!trimmed) return result

  // "explain <value>"
  if (trimmed.startsWith("explain ")) {
    result.isExplain = true
    result.explainValue = trimmed.slice(8).trim()
    return result
  }

  // "recommend <topic>"
  if (trimmed.startsWith("recommend ")) {
    result.isRecommend = true
    result.recommendTopic = trimmed.slice(10).trim()
    return result
  }

  const parts = trimmed.split(/\s+/)
  let remaining = [...parts]

  // Extract --json, --format flags.
  remaining = remaining.filter((p) => {
    if (p === "--json") {
      result.format = "json"
      return false
    }
    if (p.startsWith("--format=")) {
      result.format = p.split("=")[1]
      return false
    }
    return true
  })

  // Check aliases first.
  const firstWord = remaining[0]?.toLowerCase()
  if (firstWord && aliases[firstWord]) {
    result.generator = aliases[firstWord]
    remaining.shift()
  } else {
    // Try "uuid v7" -> "uuid.v7".
    if (remaining.length >= 2) {
      const candidate = `${remaining[0]}.${remaining[1]}`.toLowerCase()
      const match = generators.find((g) => g.address === candidate)
      if (match) {
        result.generator = match.address
        remaining.splice(0, 2)
      }
    }
    // Try bare name.
    if (!result.generator && remaining.length >= 1) {
      const bare = remaining[0].toLowerCase()
      const match = generators.find(
        (g) => g.address === bare || g.address.startsWith(bare + "."),
      )
      if (match) {
        result.generator = match.address
        remaining.shift()
      }
    }
  }

  // Remaining numeric = count (for most), or length (for password/nanoid).
  for (const part of remaining) {
    const num = parseInt(part, 10)
    if (!isNaN(num)) {
      if (result.generator === "password" || result.generator === "nanoid") {
        result.params.length = String(num)
      } else {
        result.count = num
      }
    }
  }

  return result
}

/** Match generators against a search query, ordered by relevance. */
export function matchGenerators(
  query: string,
  generators: GeneratorInfo[],
): GeneratorInfo[] {
  if (!query.trim()) return []
  const q = query.toLowerCase().trim()

  // Exact address match first.
  const exact = generators.filter((g) => g.address === q)
  if (exact.length > 0) return exact

  // Prefix match.
  const prefix = generators.filter(
    (g) =>
      g.address.startsWith(q) ||
      g.name.toLowerCase().startsWith(q) ||
      g.group.startsWith(q),
  )
  if (prefix.length > 0) return prefix

  // Fuzzy: contains match.
  return generators.filter(
    (g) =>
      g.address.includes(q) || g.description.toLowerCase().includes(q),
  )
}
