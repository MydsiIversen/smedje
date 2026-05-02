// Privacy-first analytics helpers.
// These are no-ops unless window.umami is present (injected by self-hosted
// Umami on the public demo only). Generated values are never sent.

declare global {
  interface Window {
    umami?: {
      track: (event: string, data?: Record<string, string | number>) => void
    }
  }
}

function track(event: string, data?: Record<string, string | number>) {
  if (typeof window !== "undefined" && window.umami) {
    window.umami.track(event, data)
  }
}

export function trackForge(generator: string) {
  track("forge", { generator })
}

export function trackExplainerUsed() {
  track("explainer_used")
}

export function trackPaletteOpened() {
  track("palette_opened")
}

export function trackInstallLinkClicked() {
  track("install_link_clicked")
}
