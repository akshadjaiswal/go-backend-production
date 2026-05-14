'use client'
import { useEffect, useState } from 'react'

export function ReadingProgress() {
  const [progress, setProgress] = useState(0)

  useEffect(() => {
    const pane = document.getElementById('docs-pane')
    const target = pane && pane.clientWidth > 0 ? pane : null

    function onScroll() {
      const el = target ?? document.documentElement
      const scrolled = el.scrollTop
      const total = el.scrollHeight - el.clientHeight
      setProgress(total > 0 ? (scrolled / total) * 100 : 0)
    }

    const eventTarget: Window | HTMLElement = target ?? window
    eventTarget.addEventListener('scroll', onScroll, { passive: true })
    return () => eventTarget.removeEventListener('scroll', onScroll)
  }, [])

  return (
    <div
      className="fixed top-0 left-0 z-50 h-[2px] bg-accent transition-none"
      style={{ width: `${progress}%` }}
      role="progressbar"
      aria-valuenow={Math.round(progress)}
      aria-valuemin={0}
      aria-valuemax={100}
    />
  )
}
