import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const APP_ROOT = path.resolve(__dirname, '..')
const CONTENT_ROOT = path.resolve(APP_ROOT, '..')
const OUTPUT_DIR = path.join(APP_ROOT, 'public', 'stage-images')

// Clear and recreate output dir
if (fs.existsSync(OUTPUT_DIR)) {
  fs.rmSync(OUTPUT_DIR, { recursive: true })
}
fs.mkdirSync(OUTPUT_DIR, { recursive: true })

const STAGE_REGEX = /^stage-(\d+)-(.+)$/

function dirNameToSlug(dirName) {
  return dirName.replace(/^stage-/, '')
}

function copyImagesRecursive(srcDir, destDir) {
  const entries = fs.readdirSync(srcDir, { withFileTypes: true })
  for (const entry of entries) {
    const srcPath = path.join(srcDir, entry.name)
    if (entry.isDirectory() && entry.name !== 'node_modules' && entry.name !== '.git') {
      copyImagesRecursive(srcPath, destDir)
    } else if (/\.(png|jpg|jpeg|gif|svg|webp)$/i.test(entry.name)) {
      fs.mkdirSync(destDir, { recursive: true })
      fs.copyFileSync(srcPath, path.join(destDir, entry.name))
      console.log(`  Copied: ${entry.name}`)
    }
  }
}

const dirs = fs.readdirSync(CONTENT_ROOT, { withFileTypes: true })
  .filter((d) => d.isDirectory() && STAGE_REGEX.test(d.name))
  .map((d) => d.name)
  .sort()

for (const dir of dirs) {
  const slug = dirNameToSlug(dir)
  const stagePath = path.join(CONTENT_ROOT, dir)
  const destPath = path.join(OUTPUT_DIR, slug)

  const entries = fs.readdirSync(stagePath, { withFileTypes: true })
  const hasImages = entries.some(
    (e) =>
      /\.(png|jpg|jpeg|gif|svg|webp)$/i.test(e.name) ||
      (e.isDirectory() && e.name === 'images')
  )

  if (hasImages) {
    console.log(`Processing: ${dir} -> ${slug}`)
    copyImagesRecursive(stagePath, destPath)
  }
}

console.log('Done copying stage images.')
