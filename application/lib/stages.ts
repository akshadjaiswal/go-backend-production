import fs from 'fs'
import path from 'path'
import type { Stage, StageMeta, TocHeading, Group, StageFile, StageFileGroup } from '@/types/stage'

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

// ── Code Explorer ───────────────────────────────────────────────────────────

const EXT_TO_LANG: Record<string, string> = {
  '.go':   'go',
  '.sql':  'sql',
  '.http': 'http',
  '.yml':  'yaml',
  '.yaml': 'yaml',
  '.json': 'json',
  '.sh':   'bash',
  '.toml': 'toml',
  '.mod':  'go',
}

const NAME_TO_LANG: Record<string, string> = {
  'Dockerfile':          'dockerfile',
  '.dockerignore':       'plaintext',
  '.env.example':        'ini',
  '.env.test.example':   'ini',
}

function detectLang(filename: string): string {
  if (NAME_TO_LANG[filename]) return NAME_TO_LANG[filename]
  const ext = path.extname(filename)
  return EXT_TO_LANG[ext] ?? 'text'
}

function isAllowed(filename: string): boolean {
  // Explicit names
  if (NAME_TO_LANG[filename]) return true
  // Extension allowlist
  const ext = path.extname(filename)
  const allowedExts = new Set(['.go', '.sql', '.http', '.yml', '.yaml'])
  return allowedExts.has(ext)
}

const DIR_ORDER = [
  '.', 'handlers', 'middleware', 'routes', 'models',
  'db', 'config', 'logger', 'migrations', 'validator', 'testhelpers',
]

function collectFiles(stageDirPath: string): StageFile[] {
  const results: StageFile[] = []

  function walk(dirPath: string, relDir: string) {
    let entries: fs.Dirent[]
    try {
      entries = fs.readdirSync(dirPath, { withFileTypes: true })
    } catch {
      return
    }

    for (const entry of entries) {
      if (entry.isDirectory()) {
        // Only recurse one level — no stage has deeper nesting
        // Skip hidden dirs, but allow normal subdirs
        if (relDir === '.' && !entry.name.startsWith('.')) {
          walk(path.join(dirPath, entry.name), entry.name)
        }
      } else if (isAllowed(entry.name)) {
        const filePath = relDir === '.' ? entry.name : `${relDir}/${entry.name}`
        let content = ''
        try {
          content = fs.readFileSync(path.join(dirPath, entry.name), 'utf-8')
        } catch {
          continue
        }
        results.push({
          path: filePath,
          filename: entry.name,
          dir: relDir,
          content,
          lang: detectLang(entry.name),
        })
      }
    }
  }

  walk(stageDirPath, '.')
  return results
}

export function getStageFiles(dirName: string): StageFileGroup[] {
  const stageDirPath = path.join(CONTENT_ROOT, dirName)
  const files = collectFiles(stageDirPath)

  if (files.length === 0) return []

  // Group by directory
  const map = new Map<string, StageFile[]>()
  for (const f of files) {
    const arr = map.get(f.dir) ?? []
    arr.push(f)
    map.set(f.dir, arr)
  }

  // Sort files within each group
  for (const [dir, arr] of map) {
    arr.sort((a, b) => {
      // Pin main.go first in root
      if (dir === '.') {
        if (a.filename === 'main.go') return -1
        if (b.filename === 'main.go') return 1
      }
      // Non-test files before _test.go files
      const aTest = a.filename.endsWith('_test.go')
      const bTest = b.filename.endsWith('_test.go')
      if (aTest !== bTest) return aTest ? 1 : -1
      return a.filename.localeCompare(b.filename)
    })
  }

  // Sort directory groups by DIR_ORDER
  const dirs = [...map.keys()].sort((a, b) => {
    const ai = DIR_ORDER.indexOf(a)
    const bi = DIR_ORDER.indexOf(b)
    if (ai !== -1 && bi !== -1) return ai - bi
    if (ai !== -1) return -1
    if (bi !== -1) return 1
    return a.localeCompare(b)
  })

  return dirs.map((dir) => ({ dir, files: map.get(dir)! }))
}
