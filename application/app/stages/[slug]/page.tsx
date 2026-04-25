import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { getAllStages, getStageBySlug, getStageFiles } from '@/lib/stages'
import { MarkdownRenderer } from '@/components/markdown-renderer'
import { TableOfContents } from '@/components/table-of-contents'
import { StageNav } from '@/components/stage-nav'
import { ReadingProgress } from '@/components/reading-progress'
import { ScrollToTop } from '@/components/scroll-to-top'
import { StageShortcuts } from '@/components/stage-shortcuts'
import { StageHeader } from '@/components/stage-header'
import { SplitLayout } from '@/components/split-layout'
import { CodePane } from '@/components/code-pane'

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

  const fileGroups = getStageFiles(stage.dirName)

  // Docs content node — built once, shared between split desktop and mobile stack
  const docsNode = (
    <div className="px-6 md:px-10 py-16 md:py-20">
      <StageHeader stage={stage} allSlugsInGroup={allSlugsInGroup} />

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
  )

  return (
    <>
      <ReadingProgress />
      <ScrollToTop />
      <StageShortcuts slug={stage.slug} />

      {fileGroups.length > 0 ? (
        // Split layout: docs + code side-by-side on lg+, stacked on mobile
        <SplitLayout
          docs={docsNode}
          code={<CodePane fileGroups={fileGroups} />}
        />
      ) : (
        // No source files — full-width docs only (e.g. concept-only stages)
        <div className="max-w-6xl mx-auto px-6 md:px-12 py-16 md:py-24">
          <StageHeader stage={stage} allSlugsInGroup={allSlugsInGroup} />

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
      )}
    </>
  )
}
