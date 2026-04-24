'use client'

import { CheckSquare, Square, PartyPopper } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useCompletedStages } from '@/hooks/use-bookmark'

interface CompleteButtonProps {
  slug: string
  groupLabel: string
  allSlugsInGroup: string[]
}

export function CompleteButton({ slug, groupLabel, allSlugsInGroup }: CompleteButtonProps) {
  const { isCompleted, toggle } = useCompletedStages()
  const [celebrateMsg, setCelebrateMsg] = useState<string | null>(null)
  const done = isCompleted(slug)

  function handleToggle() {
    const wasDone = isCompleted(slug)
    toggle(slug)
    if (!wasDone) {
      const otherSlugs = allSlugsInGroup.filter((s) => s !== slug)
      if (otherSlugs.every((s) => isCompleted(s))) {
        setCelebrateMsg(`${groupLabel} complete!`)
      }
    }
  }

  useEffect(() => {
    if (!celebrateMsg) return
    const timer = setTimeout(() => setCelebrateMsg(null), 4000)
    return () => clearTimeout(timer)
  }, [celebrateMsg])

  return (
    <>
      <button
        onClick={handleToggle}
        aria-label={done ? 'Mark as incomplete' : 'Mark stage as complete'}
        title={done ? 'Mark as incomplete' : 'Mark stage as complete'}
        className={`flex items-center gap-1.5 font-mono text-[10px] tracking-widest uppercase border px-3 py-1.5 transition-colors duration-100 ${
          done
            ? 'bg-foreground dark:bg-[#FAFAFA] text-background dark:text-[#0A0A0A] border-foreground dark:border-[#FAFAFA]'
            : 'border-foreground dark:border-[#FAFAFA] hover:bg-foreground hover:text-background dark:hover:bg-[#FAFAFA] dark:hover:text-[#0A0A0A]'
        }`}
      >
        {done
          ? <CheckSquare size={12} strokeWidth={2} />
          : <Square size={12} strokeWidth={1.5} />
        }
        {done ? 'Completed' : 'Mark Complete'}
      </button>

      {celebrateMsg && (
        <div className="fixed bottom-20 right-6 z-50 flex items-center gap-2 bg-foreground dark:bg-[#FAFAFA] text-background dark:text-[#0A0A0A] px-4 py-3 font-mono text-xs tracking-widest uppercase no-print animate-in">
          <PartyPopper size={14} strokeWidth={1.5} />
          {celebrateMsg}
        </div>
      )}
    </>
  )
}
