// @ts-check
import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const CONTENT_ROOT = path.join(__dirname, '..', '..')

const STAGE_REGEX = /^stage-(\d+)-(.+)$/

function dirNameToSlug(dirName) {
  return dirName.replace(/^stage-/, '')
}

function parseDirName(dirName) {
  const match = dirName.match(STAGE_REGEX)
  if (!match) return { number: '00', title: dirName, group: 1 }

  const num = match[1].padStart(2, '0')
  const n = parseInt(match[1])

  const rawTitle = match[2]
    .split('-')
    .map((word) => {
      if (['jwt', 'api', 'sql', 'db', 'http'].includes(word)) return word.toUpperCase()
      return word.charAt(0).toUpperCase() + word.slice(1)
    })
    .join(' ')

  const group = n <= 3 ? 1 : n <= 6 ? 2 : 3

  return { number: num, title: rawTitle, group }
}

function getGroupLabel(group) {
  const labels = { 1: 'Foundations', 2: 'Data & Auth', 3: 'Production' }
  return labels[group]
}

function stripMarkdown(text) {
  return text
    .replace(/```[\s\S]*?```/g, '')
    .replace(/`[^`]+`/g, '')
    .replace(/^#{1,6}\s+/gm, '')
    .replace(/!\[.*?\]\(.*?\)/g, '')
    .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
    .replace(/https?:\/\/\S+/g, '')
    .replace(/\*{1,3}([^*]+)\*{1,3}/g, '$1')
    .replace(/_{1,3}([^_]+)_{1,3}/g, '$1')
    .replace(/^>\s*/gm, '')
    .replace(/^---+$/gm, '')
    .replace(/^[-*+]\s+/gm, '')
    .replace(/^\d+\.\s+/gm, '')
    .replace(/\s+/g, ' ')
    .trim()
}

const entries = fs.readdirSync(CONTENT_ROOT, { withFileTypes: true })
const stageDirs = entries
  .filter((e) => e.isDirectory() && STAGE_REGEX.test(e.name))
  .map((e) => e.name)
  .sort((a, b) => {
    const aNum = parseInt(a.match(STAGE_REGEX)[1])
    const bNum = parseInt(b.match(STAGE_REGEX)[1])
    return aNum - bNum
  })

const index = stageDirs.map((dirName) => {
  const { number, title, group } = parseDirName(dirName)
  const readmePath = path.join(CONTENT_ROOT, dirName, 'README.md')
  let content = ''
  try {
    const raw = fs.readFileSync(readmePath, 'utf-8')
    content = stripMarkdown(raw)
  } catch {
    // no README — content stays empty
  }
  return {
    slug: dirNameToSlug(dirName),
    title,
    number,
    groupLabel: getGroupLabel(group),
    content,
  }
})

const outPath = path.join(__dirname, '..', 'public', 'search-index.json')
fs.writeFileSync(outPath, JSON.stringify(index, null, 2))
console.log(`✓ search-index.json written (${index.length} entries)`)
