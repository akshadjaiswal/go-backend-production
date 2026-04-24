export interface TocHeading {
  text: string
  slug: string
  level: number
}

export interface Stage {
  slug: string
  dirName: string
  title: string
  number: string
  group: 1 | 2 | 3
  groupLabel: string
  content: string
  headings: TocHeading[]
  readTime: number
}

export type StageMeta = Omit<Stage, 'content' | 'headings'>

export interface Group {
  number: 1 | 2 | 3
  label: string
  description: string
  stages: StageMeta[]
}

export interface StageFile {
  path: string      // relative to stage dir: "handlers/auth.go"
  filename: string  // basename: "auth.go"
  dir: string       // parent dir: "handlers" | "." for root-level files
  content: string   // raw file text
  lang: string      // shiki language id: "go" | "sql" | "yaml" | "ini" | etc.
}

export interface StageFileGroup {
  dir: string
  files: StageFile[]
}
