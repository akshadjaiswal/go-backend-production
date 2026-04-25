import { codeToHtml } from 'shiki'
import type { StageFileGroup } from '@/types/stage'
import { CodePaneClient, type HighlightedFile } from './code-pane-client'

interface CodePaneProps {
  fileGroups: StageFileGroup[]
}

export async function CodePane({ fileGroups }: CodePaneProps) {
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
        const escaped = file.content
          .replace(/&/g, '&amp;')
          .replace(/</g, '&lt;')
          .replace(/>/g, '&gt;')
        lightHtml = `<pre style="font-family:monospace;font-size:13px;line-height:1.75">${escaped}</pre>`
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
        lineCount: file.content.split('\n').length,
      })
    }
  }

  const dirs = fileGroups.map((g) => g.dir)

  return <CodePaneClient files={highlighted} dirs={dirs} />
}
