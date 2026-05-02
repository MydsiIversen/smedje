import { motion } from "motion/react"
import { springDefault } from "../../lib/motion"

interface ProgressBarProps {
  progress: number // 0 to 1
  className?: string
}

export function ProgressBar({ progress, className }: ProgressBarProps) {
  return (
    <div className={`h-1 bg-border overflow-hidden ${className ?? ""}`}>
      <motion.div
        className="h-full bg-forge"
        initial={{ width: 0 }}
        animate={{ width: `${Math.min(progress * 100, 100)}%` }}
        transition={springDefault}
      />
    </div>
  )
}
