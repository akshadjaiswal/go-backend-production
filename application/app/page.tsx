import Link from 'next/link'
import { getGroups } from '@/lib/stages'
import { StageCompletionBadge } from '@/components/stage-completion-badge'

export default function Home() {
  const groups = getGroups()

  return (
    <div className="max-w-4xl mx-auto px-6 md:px-12 py-10 md:py-16">
      {/* Hero */}
      <header className="mb-14">
        <p className="font-mono text-xs tracking-widest uppercase mb-4">
          A Complete Learning Resource
        </p>
        <h1 className="font-heading text-6xl sm:text-7xl md:text-8xl lg:text-9xl font-black tracking-tighter leading-none">
          GO
          <br />
          BACKEND
        </h1>
        <div className="h-2 bg-foreground dark:bg-[#FAFAFA] mt-8 mb-6" />
        <p className="font-body text-lg md:text-xl leading-relaxed max-w-2xl">
          10 stages across 3 groups. From spinning up an HTTP server with the standard library
          to deploying a production-ready API with Docker. Learn Go the way real backends are built.
        </p>
      </header>

      {/* Groups */}
      {groups.map((group) => (
        <section key={group.number} className="mb-14 pl-4 border-l-4 border-accent">
          <div className="flex items-baseline gap-4 mb-2">
            <span className="font-mono text-xs tracking-widest uppercase text-accent">
              Group {String(group.number).padStart(2, '0')}
            </span>
            <span className="font-mono text-xs text-muted-foreground dark:text-[#A3A3A3]">
              {group.stages.length} stages
            </span>
          </div>
          <h2 className="font-heading text-3xl md:text-4xl font-bold tracking-tight mb-2">
            {group.label}
          </h2>
          <p className="font-body text-muted-foreground dark:text-[#A3A3A3] mb-6">
            {group.description}
          </p>
          <div className="h-1 bg-foreground dark:bg-[#FAFAFA] mb-6" />

          <div className="grid grid-cols-1 md:grid-cols-2">
            {group.stages.map((stage) => (
              <Link
                key={stage.slug}
                href={`/stages/${stage.slug}`}
                className="relative group block border border-foreground dark:border-[#2A2A2A] p-5 -mt-px -ml-px hover:bg-foreground hover:text-background dark:hover:bg-[#FAFAFA] dark:hover:text-[#0A0A0A] transition-colors duration-100"
              >
                <StageCompletionBadge slug={stage.slug} />
                <span className="font-mono text-[10px] tracking-widest uppercase text-muted-foreground dark:text-[#A3A3A3] group-hover:text-background/60 dark:group-hover:text-[#0A0A0A]/60">
                  Stage {stage.number}
                </span>
                <h3 className="font-heading text-base font-semibold mt-1 leading-snug">
                  {stage.title}
                </h3>
                <span className="block font-mono text-5xl font-black leading-none mt-3 text-foreground/5 dark:text-[#FAFAFA]/5 group-hover:text-background/10 dark:group-hover:text-[#0A0A0A]/10 select-none" aria-hidden="true">
                  {stage.number}
                </span>
              </Link>
            ))}
          </div>
        </section>
      ))}
    </div>
  )
}
