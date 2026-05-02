import { motion } from "motion/react"

interface ForgeButtonProps {
  onClick: () => void
  forging: boolean
  done: boolean
}

export function ForgeButton({ onClick, forging, done }: ForgeButtonProps) {
  const label = done ? "Forged ✓" : forging ? "Forging…" : "⚒ Forge"

  return (
    <motion.button
      onClick={onClick}
      disabled={forging}
      aria-label="Forge values"
      aria-busy={forging}
      data-forge-button
      className="bg-forge text-foreground font-mono text-base py-3 w-full rounded-[4px] transition-colors duration-150 hover:border-forge border border-transparent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-forge disabled:opacity-70"
      animate={forging ? { opacity: [0.7, 1, 0.7] } : {}}
      transition={forging ? { duration: 1.5, repeat: Infinity, ease: "easeInOut" } : {}}
    >
      {label}
    </motion.button>
  )
}
