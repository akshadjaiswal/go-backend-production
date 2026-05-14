'use client'
import { Sun, Moon, Monitor } from 'lucide-react'
import { useTheme } from './theme-provider'

type Theme = 'light' | 'dark' | 'system'

const CYCLE: Record<Theme, Theme> = { light: 'dark', dark: 'system', system: 'light' }
const LABELS: Record<Theme, string> = {
  light: 'Switch to dark mode',
  dark: 'Switch to system mode',
  system: 'Switch to light mode',
}

export function ThemeToggle() {
  const { theme, resolvedTheme, setTheme } = useTheme()

  return (
    <button
      onClick={() => setTheme(CYCLE[theme])}
      aria-label={LABELS[theme]}
      title={LABELS[theme]}
      className="inline-flex items-center justify-center px-3 py-1.5 border border-foreground dark:border-[#FAFAFA] hover:bg-foreground hover:text-background dark:hover:bg-[#FAFAFA] dark:hover:text-[#0A0A0A] transition-colors duration-100"
    >
      {theme === 'system'
        ? <Monitor size={13} strokeWidth={1.5} />
        : resolvedTheme === 'dark'
          ? <Sun size={13} strokeWidth={1.5} />
          : <Moon size={13} strokeWidth={1.5} />
      }
    </button>
  )
}
