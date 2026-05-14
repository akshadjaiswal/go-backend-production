'use client'
import { useRef } from 'react'

interface ShortcutsModalProps {
  onClose: () => void
}

const shortcuts = [
  { keys: ['/'], description: 'Open search' },
  { keys: ['↑', '↓'], description: 'Navigate search results' },
  { keys: ['Enter'], description: 'Open selected result' },
  { keys: ['Esc'], description: 'Close modal / search' },
  { keys: ['?'], description: 'Open keyboard shortcuts' },
  { keys: ['b'], description: 'Bookmark current stage' },
]

export function ShortcutsModal({ onClose }: ShortcutsModalProps) {
  const modalRef = useRef<HTMLDivElement>(null)

  function trapFocus(e: React.KeyboardEvent) {
    if (e.key !== 'Tab') return
    const focusable = modalRef.current?.querySelectorAll<HTMLElement>(
      'button, input, a[href], [tabindex]:not([tabindex="-1"])'
    )
    if (!focusable || focusable.length === 0) return
    const first = focusable[0]
    const last = focusable[focusable.length - 1]
    if (e.shiftKey) {
      if (document.activeElement === first) { e.preventDefault(); last.focus() }
    } else {
      if (document.activeElement === last) { e.preventDefault(); first.focus() }
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    trapFocus(e)
    if (e.key === 'Escape') onClose()
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center pt-24 px-4"
      onClick={onClose}
    >
      <div className="absolute inset-0 bg-foreground/40 dark:bg-[#FAFAFA]/20" />
      <div
        ref={modalRef}
        className="relative w-full max-w-sm border-2 border-foreground dark:border-[#FAFAFA] bg-background dark:bg-[#0A0A0A]"
        onClick={(e) => e.stopPropagation()}
        onKeyDown={handleKeyDown}
      >
        <div className="flex items-center justify-between px-4 py-3 border-b border-border-light dark:border-[#2A2A2A]">
          <span className="font-mono text-xs tracking-widest uppercase">Keyboard Shortcuts</span>
          <button
            onClick={onClose}
            className="font-mono text-xs text-muted-foreground dark:text-[#A3A3A3] hover:text-foreground dark:hover:text-[#FAFAFA]"
          >
            ESC
          </button>
        </div>

        <ul className="py-2">
          {shortcuts.map((s, i) => (
            <li key={i} className="flex items-center justify-between px-4 py-2.5">
              <span className="font-body text-sm">{s.description}</span>
              <span className="flex items-center gap-1">
                {s.keys.map((k) => (
                  <kbd
                    key={k}
                    className="font-mono text-xs px-2 py-0.5 border border-foreground dark:border-[#FAFAFA] bg-muted dark:bg-[#1A1A1A]"
                  >
                    {k}
                  </kbd>
                ))}
              </span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}
