import type { Transition, Variants } from "motion/react"

export const springDefault: Transition = {
  type: "spring",
  stiffness: 400,
  damping: 30,
  mass: 1,
}

export const staggerDelay = 0.03 // 30ms

export const fadeIn: Variants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: springDefault },
}

export const stampVariant: Variants = {
  hidden: { opacity: 0, y: -12, scale: 0.96 },
  visible: {
    opacity: 1,
    y: 0,
    scale: 1,
    transition: { duration: 0.4, ease: "easeOut" },
  },
}

export const staggerContainer: Variants = {
  hidden: {},
  visible: {
    transition: {
      staggerChildren: staggerDelay,
    },
  },
}

export const glowKeyframes = {
  boxShadow: [
    "0 0 0px rgba(226, 104, 58, 0)",
    "0 0 12px rgba(226, 104, 58, 0.6)",
    "0 0 0px rgba(226, 104, 58, 0)",
  ],
}

export const glowTransition: Transition = {
  duration: 0.6,
  ease: "easeOut",
}
