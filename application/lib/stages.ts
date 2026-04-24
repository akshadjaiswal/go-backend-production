import fs from 'fs'
import path from 'path'
import type { Stage, StageMeta, TocHeading, Group } from '@/types/stage'

// Content is read from the parent repo at build time.
// application/ → go-backend-production/ (one level up)
const CONTENT_ROOT = path.join(process.cwd(), '..')

// Stage dirs follow a consistent naming convention: stage-NN-topic
// e.g. stage-01-basics, stage-10-deployment
const STAGE_REGEX = /^stage-(\d+)-(.+)$/

function getAllStageDirs(): string[] {
  const entries = fs.readdirSync(CONTENT_ROOT, { withFileTypes: true })
  return entries
    .filter((e) => e.isDirectory() && STAGE_REGEX.test(e.name))
    .map((e) => e.name)
    .sort((a, b) => {
      const aNum = parseInt(a.match(STAGE_REGEX)![1])
      const bNum = parseInt(b.match(STAGE_REGEX)![1])
      return aNum - bNum
    })
}

// Parse stage directory name into structured metadata
function parseDirName(dirName: string): { number: string; title: string; group: 1 | 2 | 3 } {
  const match = dirName.match(STAGE_REGEX)
  if (!match) return { number: '00', title: dirName, group: 1 }

  const num = match[1].padStart(2, '0')
  const n = parseInt(match[1])

  // Humanize the topic slug: "jwt-auth" → "JWT Auth", "basics" → "Basics"
  const rawTitle = match[2]
    .split('-')
    .map((word) => {
      // Acronyms that should be upper-cased
      if (['jwt', 'api', 'sql', 'db', 'http'].includes(word)) return word.toUpperCase()
      return word.charAt(0).toUpperCase() + word.slice(1)
    })
    .join(' ')

  // Group assignment:
  // Foundations  (1-3):  HTTP, routing, middleware
  // Data & Auth  (4-6):  database, auth, validation
  // Production   (7-10): config, logging, testing, deployment
  const group: 1 | 2 | 3 = n <= 3 ? 1 : n <= 6 ? 2 : 3

  return { number: num, title: rawTitle, group }
}

// Slug is the directory name with the "stage-" prefix stripped
// stage-01-basics → 01-basics
export function dirNameToSlug(dirName: string): string {
  return dirName.replace(/^stage-/, '')
}

function slugify(text: string): string {
  return text
    .toLowerCase()
    .replace(/<[^>]*>/g, '')
    .replace(/[^a-z0-9\s-]/g, '')
    .trim()
    .replace(/\s+/g, '-')
    .replace(/-+/g, '-')
}

function computeReadTime(content: string): number {
  const words = content.trim().split(/\s+/).length
  return Math.max(1, Math.ceil(words / 200))
}

function extractHeadings(markdown: string): TocHeading[] {
  const headings: TocHeading[] = []
  const lines = markdown.split('\n')
  const slugCount = new Map<string, number>()

  for (const line of lines) {
    const match = line.match(/^(#{2,4})\s+(.+?)\s*$/)
    if (!match) continue

    const level = match[1].length
    const text = match[2]
      .replace(/\*\*/g, '')
      .replace(/`/g, '')
      .trim()

    const base = slugify(text)
    const count = slugCount.get(base) ?? 0
    slugCount.set(base, count + 1)
    const slug = count === 0 ? base : `${base}-${count}`

    headings.push({ text, slug, level })
  }

  return headings
}

function getGroupLabel(group: 1 | 2 | 3): string {
  const labels: Record<number, string> = {
    1: 'Foundations',
    2: 'Data & Auth',
    3: 'Production',
  }
  return labels[group]
}

export function getStageBySlug(slug: string): Stage | null {
  const dirs = getAllStageDirs()
  const dirName = dirs.find((d) => dirNameToSlug(d) === slug)

  if (!dirName) return null

  const readmePath = path.join(CONTENT_ROOT, dirName, 'README.md')
  const content = fs.readFileSync(readmePath, 'utf-8')
  const { number, title, group } = parseDirName(dirName)
  const headings = extractHeadings(content)

  return {
    slug,
    dirName,
    title,
    number,
    group,
    groupLabel: getGroupLabel(group),
    content,
    headings,
    readTime: computeReadTime(content),
  }
}

export function getAllStages(): StageMeta[] {
  const dirs = getAllStageDirs()
  return dirs.map((dirName) => {
    const { number, title, group } = parseDirName(dirName)
    const readmePath = path.join(CONTENT_ROOT, dirName, 'README.md')
    let readTime = 1
    try {
      const content = fs.readFileSync(readmePath, 'utf-8')
      readTime = computeReadTime(content)
    } catch {}
    return {
      slug: dirNameToSlug(dirName),
      dirName,
      title,
      number,
      group,
      groupLabel: getGroupLabel(group),
      readTime,
    }
  })
}

export function getGroups(): Group[] {
  const stages = getAllStages()

  const groupDescriptions: Record<number, string> = {
    1: 'HTTP servers, routing with Chi, and custom middleware — the building blocks of every Go backend.',
    2: 'PostgreSQL with sqlx, JWT authentication, and input validation — making data safe and access secure.',
    3: 'Config management, structured logging with slog, integration testing, and Docker deployment.',
  }

  const grouped = new Map<number, StageMeta[]>()
  for (const stage of stages) {
    const existing = grouped.get(stage.group) || []
    existing.push(stage)
    grouped.set(stage.group, existing)
  }

  return ([1, 2, 3] as const).map((num) => ({
    number: num,
    label: getGroupLabel(num),
    description: groupDescriptions[num],
    stages: grouped.get(num) || [],
  }))
}

export function getSearchIndex() {
  return getAllStages().map((s) => ({
    slug: s.slug,
    title: s.title,
    number: s.number,
    groupLabel: s.groupLabel,
  }))
}
