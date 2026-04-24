'use client'

import { useCallback, useEffect, useRef, useState } from 'react'

const SPLIT_RATIO_KEY = 'gbp_split_ratio'
const DEFAULT_RATIO = 55
const MIN_RATIO = 25
const MAX_RATIO = 75

interface SplitLayoutProps {
  docs: React.ReactNode
  code: React.ReactNode
}

export function SplitLayout({ docs, code }: SplitLayoutProps) {
  const [ratio, setRatio] = useState(DEFAULT_RATIO)
  const [dragging, setDragging] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  // Load persisted ratio on mount
  useEffect(() => {
    try {
      const stored = localStorage.getItem(SPLIT_RATIO_KEY)
      if (stored) {
        const n = Number(stored)
        if (!isNaN(n)) setRatio(Math.max(MIN_RATIO, Math.min(MAX_RATIO, n)))
      }
    } catch {}
  }, [])

  const startDrag = useCallback((e: React.PointerEvent<HTMLDivElement>) => {
    e.preventDefault()
    e.currentTarget.setPointerCapture(e.pointerId)
    setDragging(true)
    document.body.classList.add('dragging-split')
  }, [])

  const onPointerMove = useCallback(
    (e: React.PointerEvent<HTMLDivElement>) => {
      if (!dragging || !containerRef.current) return
      const rect = containerRef.current.getBoundingClientRect()
      const newRatio = ((e.clientX - rect.left) / rect.width) * 100
      const clamped = Math.max(MIN_RATIO, Math.min(MAX_RATIO, newRatio))
      setRatio(clamped)
    },
    [dragging],
  )

  const endDrag = useCallback(
    (ratio: number) => {
      setDragging(false)
      document.body.classList.remove('dragging-split')
      try {
        localStorage.setItem(SPLIT_RATIO_KEY, String(Math.round(ratio)))
      } catch {}
    },
    [],
  )

  return (
    <>
      {/* ── Desktop split (lg+) ─────────────────────────────────────────── */}
      <div
        ref={containerRef}
        className="hidden lg:flex h-[calc(100vh-var(--header-h))] w-full overflow-hidden"
        onPointerMove={onPointerMove}
        onPointerUp={() => endDrag(ratio)}
        onPointerLeave={() => { if (dragging) endDrag(ratio) }}
      >
        {/* Docs pane — scrolls normally */}
        <div
          id="docs-pane"
          className="docs-pane overflow-y-auto h-full shrink-0"
          style={{ width: `${ratio}%` }}
        >
          {docs}
        </div>

        {/* Drag handle */}
        <div
          className={[
            'w-1 shrink-0 h-full cursor-col-resize transition-colors duration-100 relative z-10',
            dragging
              ? 'bg-accent'
              : 'bg-foreground dark:bg-[#2A2A2A] hover:bg-accent',
          ].join(' ')}
          onPointerDown={startDrag}
          title="Drag to resize"
        />

        {/* Code pane — sticky, fills height */}
        <div
          className="code-pane-wrapper overflow-hidden h-full shrink-0"
          style={{ width: `${100 - ratio}%` }}
        >
          {code}
        </div>
      </div>

      {/* ── Mobile stacked (below lg) ────────────────────────────────────── */}
      <div className="lg:hidden">
        {docs}
        <div className="border-t-4 border-foreground dark:border-[#2A2A2A] mt-0">
          <div className="bg-[#252526] px-4 py-2 font-mono text-xs tracking-widest uppercase text-[#858585]">
            Source Files
          </div>
          {code}
        </div>
      </div>
    </>
  )
}
