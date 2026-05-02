import { motion } from "motion/react"
import { glowKeyframes, glowTransition } from "../../lib/motion"

interface GlowPulseProps {
  children: React.ReactNode
  active?: boolean
  className?: string
}

export function GlowPulse({ children, active = false, className }: GlowPulseProps) {
  return (
    <motion.div
      animate={active ? glowKeyframes : {}}
      transition={glowTransition}
      className={className}
    >
      {children}
    </motion.div>
  )
}
