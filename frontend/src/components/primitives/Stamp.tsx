import { motion } from "motion/react"
import { stampVariant } from "../../lib/motion"

interface StampProps {
  children: React.ReactNode
  className?: string
}

export function Stamp({ children, className }: StampProps) {
  return (
    <motion.div
      variants={stampVariant}
      initial="hidden"
      animate="visible"
      className={className}
    >
      {children}
    </motion.div>
  )
}
