'use client'

import { useState } from 'react'
import { CopyButton } from './copy-button'

export interface HighlightedFile {
  path: string
  filename: string
  dir: string
  lang: string
  rawContent: string
  lightHtml: string
  darkHtml: string
}

interface CodeExplorerClientProps {
  files: HighlightedFile[]
  dirs: string[]
}

export function CodeExplorerClient({ files, dirs }: CodeExplorerClientProps) {
  const [activeDir, setActiveDir] = useState<string>(dirs[0] ?? '.')
  const [selectedPath, setSelectedPath] = useState<string>(files[0]?.path ?? '')

  const dirFiles = files.filter((f) => f.dir === activeDir)

  // If selected file is not in the active dir, fall back to first file of active dir
  const selectedFile =
    files.find((f) => f.path === selectedPath && f.dir === activeDir) ??
    dirFiles[0]

  function switchDir(dir: string) {
    setActiveDir(dir)
    const first = files.find((f) => f.dir === dir)
    if (first) setSelectedPath(first.path)
  }

  return (
    <section className="mt-16 no-print" aria-label="Source Files">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <h2 className="font-heading text-2xl font-bold tracking-tight">Source Files</h2>
        <span className="font-mono text-xs tracking-widest uppercase text-muted-foreground dark:text-[#A3A3A3]">
          {files.length} file{files.length !== 1 ? 's' : ''}
        </span>
      </div>

      <div className="border border-foreground dark:border-[#2A2A2A]">
        {/* Directory tabs — only shown when more than one dir */}
        {dirs.length > 1 && (
          <div className="flex overflow-x-auto border-b border-foreground dark:border-[#2A2A2A]">
            {dirs.map((dir) => (
              <button
                key={dir}
                onClick={() => switchDir(dir)}
                className={[
                  'shrink-0 px-4 py-2 font-mono text-xs tracking-widest uppercase border-r border-foreground dark:border-[#2A2A2A] transition-colors duration-100',
                  activeDir === dir
                    ? 'bg-foreground text-background dark:bg-[#FAFAFA] dark:text-[#0A0A0A]'
                    : 'text-muted-foreground dark:text-[#A3A3A3] hover:bg-foreground hover:text-background dark:hover:bg-[#FAFAFA] dark:hover:text-[#0A0A0A]',
                ].join(' ')}
              >
                {dir === '.' ? 'root' : dir}
              </button>
            ))}
          </div>
        )}

        <div className="flex flex-col sm:flex-row">
          {/* File list — left panel */}
          <div className="sm:w-48 shrink-0 border-b sm:border-b-0 sm:border-r border-foreground dark:border-[#2A2A2A]">
            <ul>
              {dirFiles.map((file) => (
                <li key={file.path}>
                  <button
                    onClick={() => setSelectedPath(file.path)}
                    title={file.filename}
                    className={[
                      'w-full text-left px-3 py-2 font-mono text-xs border-b border-foreground dark:border-[#2A2A2A] last:border-b-0 transition-colors duration-100 truncate',
                      selectedFile?.path === file.path
                        ? 'bg-accent text-white'
                        : 'hover:bg-foreground hover:text-background dark:hover:bg-[#FAFAFA] dark:hover:text-[#0A0A0A]',
                    ].join(' ')}
                  >
                    {file.filename}
                  </button>
                </li>
              ))}
            </ul>
          </div>

          {/* Code viewer — right panel */}
          {selectedFile && (
            <div className="flex-1 min-w-0">
              {/* Breadcrumb + lang */}
              <div className="flex items-center justify-between px-4 py-2 border-b border-foreground dark:border-[#2A2A2A] bg-[#F5F5F5] dark:bg-[#111111]">
                <span className="font-mono text-[11px] text-muted-foreground dark:text-[#A3A3A3] truncate">
                  {selectedFile.path}
                </span>
                <span className="font-mono text-[10px] tracking-widest uppercase text-muted-foreground dark:text-[#A3A3A3] ml-3 shrink-0">
                  {selectedFile.lang}
                </span>
              </div>

              {/* Code + copy button */}
              <div className="relative">
                <CopyButton text={selectedFile.rawContent} />
                <div className="overflow-x-auto max-h-[600px] overflow-y-auto bg-[#fafafa] dark:bg-[#1e1e1e] px-5 pb-5 pt-10 shiki-wrapper">
                  <div
                    className="shiki-light"
                    dangerouslySetInnerHTML={{ __html: selectedFile.lightHtml }}
                  />
                  <div
                    className="shiki-dark"
                    dangerouslySetInnerHTML={{ __html: selectedFile.darkHtml }}
                  />
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </section>
  )
}
