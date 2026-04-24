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
  lineCount: number
}

interface CodePaneClientProps {
  files: HighlightedFile[]
  dirs: string[]
}

export function CodePaneClient({ files, dirs }: CodePaneClientProps) {
  const [activeDir, setActiveDir] = useState<string>(dirs[0] ?? '.')
  const [selectedPath, setSelectedPath] = useState<string>(files[0]?.path ?? '')
  const [treeOpen, setTreeOpen] = useState(true)

  const dirFiles = files.filter((f) => f.dir === activeDir)
  const selectedFile =
    files.find((f) => f.path === selectedPath && f.dir === activeDir) ?? dirFiles[0]

  function switchDir(dir: string) {
    setActiveDir(dir)
    const first = files.find((f) => f.dir === dir)
    if (first) setSelectedPath(first.path)
  }

  return (
    // code-pane class triggers always-dark shiki override in globals.css
    <div className="code-pane flex flex-col h-full bg-[#1e1e1e] overflow-hidden">

      {/* Tab bar — directory tabs + tree toggle */}
      <div className="flex items-stretch bg-[#252526] border-b border-[#3E3E42] overflow-x-auto shrink-0">
        <div className="flex overflow-x-auto flex-1">
          {dirs.map((dir) => (
            <button
              key={dir}
              onClick={() => switchDir(dir)}
              className={[
                'shrink-0 px-4 py-2 font-mono text-xs border-r border-[#3E3E42] transition-colors duration-100',
                activeDir === dir
                  ? 'bg-[#1e1e1e] text-white border-t-2 border-t-accent'
                  : 'bg-[#2D2D2D] text-[#858585] hover:bg-[#1e1e1e] hover:text-white border-t-2 border-t-transparent',
              ].join(' ')}
            >
              {dir === '.' ? 'root' : dir}
            </button>
          ))}
        </div>

        {/* Tree toggle button */}
        <button
          onClick={() => setTreeOpen((o) => !o)}
          title={treeOpen ? 'Collapse file tree' : 'Expand file tree'}
          className="shrink-0 px-3 py-2 text-[#858585] hover:text-white hover:bg-[#2D2D2D] transition-colors duration-100 font-mono text-xs border-l border-[#3E3E42]"
        >
          {treeOpen ? '‹' : '›'}
        </button>
      </div>

      {/* Body: file tree + code area */}
      <div className="flex flex-1 min-h-0 overflow-hidden">

        {/* File tree (collapsible) */}
        <div
          className={[
            'shrink-0 border-r border-[#3E3E42] overflow-y-auto transition-all duration-150 bg-[#1e1e1e]',
            treeOpen ? 'w-36' : 'w-0 overflow-hidden',
          ].join(' ')}
        >
          <ul className="py-1">
            {dirFiles.map((file) => (
              <li key={file.path}>
                <button
                  onClick={() => setSelectedPath(file.path)}
                  title={file.filename}
                  className={[
                    'w-full text-left px-3 py-1.5 font-mono text-xs truncate border-l-2 transition-colors duration-100',
                    selectedFile?.path === file.path
                      ? 'bg-[#094771] text-white border-l-accent'
                      : 'text-[#CCCCCC] hover:bg-[#2A2D2E] border-l-transparent',
                  ].join(' ')}
                >
                  {file.filename}
                </button>
              </li>
            ))}
          </ul>
        </div>

        {/* Code area — single scrolling container so line numbers move with code */}
        {selectedFile && (
          <div className="flex-1 min-w-0 overflow-hidden relative">
            {/* Copy button */}
            <div className="absolute top-2 right-2 z-10">
              <CopyButton text={selectedFile.rawContent} />
            </div>

            {/* Single scroll container — line numbers via CSS counter, always in sync with wrapping */}
            <div className="h-full overflow-y-auto bg-[#1e1e1e] py-5 pr-5">
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
        )}
      </div>

      {/* Status bar */}
      {selectedFile && (
        <div className="shrink-0 h-5 flex items-center justify-between px-3 bg-accent text-white font-mono text-[10px] tracking-wide">
          <span className="truncate">{selectedFile.path}</span>
          <span className="shrink-0 ml-4 uppercase">
            {selectedFile.lang} · {selectedFile.lineCount} lines
          </span>
        </div>
      )}
    </div>
  )
}
