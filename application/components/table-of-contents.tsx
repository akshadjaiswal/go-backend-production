'use client'

import { useEffect, useState } from 'react'
import type { TocHeading } from '@/types/stage'

interface TableOfContentsProps {
  headings: TocHeading[]
}

export function TableOfContents({ headings }: TableOfContentsProps) {
  const [activeId, setActiveId] = useState<string>('')
  const [scrollPct, setScrollPct] = useState(0)

  useEffect(() => {
    // In split-layout mode the docs pane scrolls instead of window.
    // The docs pane div has id="docs-pane" — detect it by checking clientWidth > 0.
    const paneEl = document.getElementById('docs-pane')
    const inSplitMode = paneEl !== null && paneEl.clientWidth > 0

    const scrollRoot = inSplitMode ? paneEl : null  // null = viewport for IntersectionObserver

    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            setActiveId(entry.target.id)
          }
        }
      },
      { rootMargin: '-80px 0px -80% 0px', root: scrollRoot },
    )

    headings.forEach((h) => {
      const el = document.getElementById(h.slug)
      if (el) observer.observe(el)
    })

    function onScroll() {
      if (inSplitMode && paneEl) {
        const scrollable = paneEl.scrollHeight - paneEl.clientHeight
        setScrollPct(scrollable > 0 ? Math.round((paneEl.scrollTop / scrollable) * 100) : 0)
      } else {
        const scrollable = document.body.scrollHeight - window.innerHeight
        setScrollPct(scrollable > 0 ? Math.round((window.scrollY / scrollable) * 100) : 0)
      }
    }

    onScroll()
    const eventTarget: EventTarget = inSplitMode && paneEl ? paneEl : window
    eventTarget.addEventListener('scroll', onScroll, { passive: true })

    return () => {
      observer.disconnect()
      eventTarget.removeEventListener('scroll', onScroll)
    }
  }, [headings])

  if (headings.length === 0) return null

  return (
    <nav className="sticky top-24">
      <div className="flex items-center justify-between mb-3">
        <h4 className="font-mono text-xs font-bold uppercase tracking-widest">
          On this page
        </h4>
        <span className="font-mono text-[9px] text-muted-foreground dark:text-[#A3A3A3] tabular-nums">
          {scrollPct}%
        </span>
      </div>
      <div className="h-px bg-foreground dark:bg-[#2A2A2A] mb-4 relative overflow-hidden">
        <div
          className="absolute inset-y-0 left-0 bg-accent transition-all duration-150"
          style={{ width: `${scrollPct}%` }}
        />
      </div>
      <ul className="space-y-0.5">
        {headings.map((h) => (
          <li key={h.slug}>
            <a
              href={`#${h.slug}`}
              style={{ paddingLeft: `${(h.level - 2) * 12}px` }}
              className={`block text-sm py-1 px-2 font-body transition-colors duration-100 ${
                activeId === h.slug
                  ? 'bg-foreground dark:bg-[#FAFAFA] text-background dark:text-[#0A0A0A]'
                  : 'text-muted-foreground dark:text-[#A3A3A3] hover:bg-foreground dark:hover:bg-[#FAFAFA] hover:text-background dark:hover:text-[#0A0A0A]'
              }`}
            >
              {h.text}
            </a>
          </li>
        ))}
      </ul>
    </nav>
  )
}
