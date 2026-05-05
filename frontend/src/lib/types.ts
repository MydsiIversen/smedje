export interface GeneratorInfo {
  name: string
  address: string
  group: string
  category: string
  description: string
  rationale?: string
}

export interface FlagDef {
  name: string
  type: "string" | "int" | "bool"
  default?: string
  description: string
  options?: string[]
}

export interface GeneratorSchema extends GeneratorInfo {
  flags: FlagDef[]
  supports: {
    count: boolean
    seed: boolean
    bench: boolean
  }
  exampleOutput: Record<string, string> | null
}

export interface LayoutSegment {
  start: number
  end: number
  label: string
  type: "time" | "random" | "version" | "counter" | "type" | "meta"
  value: string
  description: string
}

export interface ExplainResponse {
  detected: string
  spec?: string
  layout: LayoutSegment[]
  fields: Record<string, string>
  alternateForms?: Record<string, string>
}

export interface Recommendation {
  use_case: string
  primary: string
  why: string
  command: string
  alternatives?: { name: string; when: string }[]
  avoid?: string[]
}

export interface VersionInfo {
  version: string
  commit: string
  goVersion: string
  publicMode: boolean
}

export interface Artifact {
  value: string
  fields: Record<string, string>
  sensitiveKeys?: string[]
}

export type SSEEventType = "status" | "progress" | "artifact" | "done" | "error"

export interface SSEStatusEvent {
  type: "status"
  message: string
}

export interface SSEProgressEvent {
  type: "progress"
  current: number
  total: number
  opsPerSec: number
}

export interface SSEArtifactEvent {
  type: "artifact"
  value: string
  fields: Record<string, string>
  sensitiveKeys?: string[]
}

export interface SSEDoneEvent {
  type: "done"
  durationMs: number
  count: number
  formatted?: string
}

export interface SSEErrorEvent {
  type: "error"
  message: string
}

export type SSEEvent =
  | SSEStatusEvent
  | SSEProgressEvent
  | SSEArtifactEvent
  | SSEDoneEvent
  | SSEErrorEvent
