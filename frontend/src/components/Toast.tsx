import { Stamp } from "./primitives"

interface ToastProps {
  message: string
  visible: boolean
}

export function Toast({ message, visible }: ToastProps) {
  if (!visible) return null

  return (
    <div className="fixed bottom-6 right-6 z-50">
      <Stamp>
        <div className="bg-panel border border-border px-4 py-2 font-mono text-sm text-foreground">
          {message}
        </div>
      </Stamp>
    </div>
  )
}
