import { motion } from "motion/react"
import { fadeIn } from "../../lib/motion"

interface RevealProps {
  children: React.ReactNode
  className?: string
}

export function Reveal({ children, className }: RevealProps) {
  return (
    <motion.div
      variants={fadeIn}
      initial="hidden"
      animate="visible"
      className={className}
    >
      {children}
    </motion.div>
  )
}
