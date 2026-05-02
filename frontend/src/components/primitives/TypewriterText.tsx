import { useState, useEffect } from "react"

interface TypewriterTextProps {
  text: string
  speed?: number // ms per character, default 15
  className?: string
  onComplete?: () => void
}

export function TypewriterText({ text, speed = 15, className, onComplete }: TypewriterTextProps) {
  const [displayed, setDisplayed] = useState("")

  useEffect(() => {
    setDisplayed("")
    let i = 0
    const interval = setInterval(() => {
      i++
      setDisplayed(text.slice(0, i))
      if (i >= text.length) {
        clearInterval(interval)
        onComplete?.()
      }
    }, speed)
    return () => clearInterval(interval)
  }, [text, speed, onComplete])

  return <span className={className}>{displayed}</span>
}
