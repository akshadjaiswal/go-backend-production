import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { getAllStages, getStageBySlug } from '@/lib/stages'
import { MarkdownRenderer } from '@/components/markdown-renderer'
import { TableOfContents } from '@/components/table-of-contents'
import { StageNav } from '@/components/stage-nav'
import { ReadingProgress } from '@/components/reading-progress'
import { BookmarkButton } from '@/components/bookmark-button'
import { CompleteButton } from '@/components/complete-button'
import { PrintButton } from '@/components/print-button'
import { ScrollToTop } from '@/components/scroll-to-top'
import { StageShortcuts } from '@/components/stage-shortcuts'

export function generateStaticParams() {
  const stages = getAllStages()
  return stages.map((s) => ({ slug: s.slug }))
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ slug: string }>
}): Promise<Metadata> {
  const { slug } = await params
  const stage = getStageBySlug(slug)
  if (!stage) return { title: 'Stage Not Found' }
  return {
    title: stage.title,
    description: `Stage ${stage.number}: ${stage.title}`,
  }
}

export default async function StagePage({
  params,
}: {
  params: Promise<{ slug: string }>
}) {
  const { slug } = await params
  const stage = getStageBySlug(slug)

  if (!stage) {
    notFound()
  }

  const allStages = getAllStages()
  const currentIndex = allStages.findIndex((s) => s.slug === stage.slug)
  const prev = currentIndex > 0 ? allStages[currentIndex - 1] : null
  const next = currentIndex < allStages.length - 1 ? allStages[currentIndex + 1] : null
  const allSlugsInGroup = allStages
    .filter((s) => s.group === stage.group)
    .map((s) => s.slug)

  return (
    <>
      <ReadingProgress />
      <ScrollToTop />
      <StageShortcuts slug={stage.slug} />
      <div className="max-w-6xl mx-auto px-6 md:px-12 py-16 md:py-24">
        <header className="mb-12">
          <div className="flex flex-col gap-0.5 mb-3">
            <span className="font-mono text-xs tracking-widest uppercase text-accent">
              {stage.groupLabel} · Stage {stage.number}
            </span>
            <span className="font-mono text-[10px] tracking-widest text-muted-foreground dark:text-[#A3A3A3]">
              {stage.readTime} min read
            </span>
          </div>
          <h1 className="font-heading text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-black tracking-tight leading-tight">
            {stage.title}
          </h1>
          <div className="h-2 bg-foreground dark:bg-[#FAFAFA] mt-8" />
          <div className="mt-4 flex items-center gap-3 flex-wrap">
            <BookmarkButton slug={stage.slug} />
            <CompleteButton slug={stage.slug} groupLabel={stage.groupLabel} allSlugsInGroup={allSlugsInGroup} />
            <PrintButton />
          </div>
        </header>

        <div className="flex gap-12">
          <article className="flex-1 min-w-0">
            <MarkdownRenderer content={stage.content} stageSlug={stage.slug} />
          </article>

          {stage.headings.length > 0 && (
            <aside className="hidden xl:block w-56 shrink-0">
              <TableOfContents headings={stage.headings} />
            </aside>
          )}
        </div>

        <div className="h-1 bg-foreground dark:bg-[#FAFAFA] mt-16 mb-8" />
        <StageNav prev={prev} next={next} />
      </div>
    </>
  )
}
