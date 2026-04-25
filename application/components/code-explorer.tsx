import { codeToHtml } from 'shiki'
import type { StageFileGroup } from '@/types/stage'
import { CodeExplorerClient, type HighlightedFile } from './code-explorer-client'

interface CodeExplorerProps {
  fileGroups: StageFileGroup[]
}

export async function CodeExplorer({ fileGroups }: CodeExplorerProps) {
  if (fileGroups.length === 0) return null

  const highlighted: HighlightedFile[] = []

  for (const group of fileGroups) {
    for (const file of group.files) {
      let lightHtml: string
      let darkHtml: string

      try {
        lightHtml = await codeToHtml(file.content, { lang: file.lang, theme: 'github-light' })
        darkHtml  = await codeToHtml(file.content, { lang: file.lang, theme: 'github-dark' })
      } catch {
        // Fallback: wrap raw content in a pre so it still renders
        const escaped = file.content
          .replace(/&/g, '&amp;')
          .replace(/</g, '&lt;')
          .replace(/>/g, '&gt;')
        lightHtml = `<pre style="font-family:monospace;font-size:13px">${escaped}</pre>`
        darkHtml  = lightHtml
      }

      highlighted.push({
        path: file.path,
        filename: file.filename,
        dir: file.dir,
        lang: file.lang,
        rawContent: file.content,
        lightHtml,
        darkHtml,
      })
    }
  }

  const dirs = fileGroups.map((g) => g.dir)

  return <CodeExplorerClient files={highlighted} dirs={dirs} />
}
