import type { StageMeta } from '@/types/stage'
import { BookmarkButton } from '@/components/bookmark-button'
import { CompleteButton } from '@/components/complete-button'
import { PrintButton } from '@/components/print-button'

interface StageHeaderProps {
  stage: StageMeta
  allSlugsInGroup: string[]
}

export function StageHeader({ stage, allSlugsInGroup }: StageHeaderProps) {
  return (
    <header className="mb-12">
      <div className="flex flex-col gap-0.5 mb-3">
        <span className="font-mono text-xs tracking-widest uppercase text-accent">
          {stage.groupLabel} · Stage {stage.number}
        </span>
        <span className="font-mono text-[10px] tracking-widest text-muted-foreground dark:text-[#A3A3A3]">
          {stage.readTime} min read
        </span>
      </div>
      <h1 className="font-heading text-4xl sm:text-5xl md:text-6xl lg:text-5xl xl:text-6xl font-black tracking-tight leading-tight">
        {stage.title}
      </h1>
      <div className="h-2 bg-foreground dark:bg-[#FAFAFA] mt-8" />
      <div className="mt-4 flex items-center gap-3 flex-wrap">
        <BookmarkButton slug={stage.slug} />
        <CompleteButton slug={stage.slug} groupLabel={stage.groupLabel} allSlugsInGroup={allSlugsInGroup} />
        <PrintButton />
      </div>
    </header>
  )
}
