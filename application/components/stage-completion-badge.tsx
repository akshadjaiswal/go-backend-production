'use client'
import { Check } from 'lucide-react'
import { useCompletedStages } from '@/hooks/use-bookmark'

interface StageCompletionBadgeProps {
  slug: string
}

export function StageCompletionBadge({ slug }: StageCompletionBadgeProps) {
  const { isCompleted } = useCompletedStages()
  if (!isCompleted(slug)) return null

  return (
    <span className="absolute top-2 right-2 w-5 h-5 bg-accent flex items-center justify-center z-10">
      <Check size={11} strokeWidth={2.5} className="text-background dark:text-[#0A0A0A]" />
    </span>
  )
}
