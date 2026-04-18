export interface HunkLine {
  type: "add" | "del" | "ctx"
  oldLn: number | null
  newLn: number | null
  text: string
}

export interface Hunk {
  header: string
  oldStart: number
  oldEnd: number
  newStart: number
  newEnd: number
  lines: HunkLine[]
}

export function parseHunks(patch: string): Hunk[] {
  const lines = patch.split("\n")
  const hunks: Hunk[] = []
  let current: Hunk | null = null
  let oldLine = 0
  let newLine = 0

  for (const line of lines) {
    const m = line.match(/^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@(.*)/)
    if (m) {
      const oldStart = parseInt(m[1], 10)
      const oldCount = m[2] !== undefined ? parseInt(m[2], 10) : 1
      const newStart = parseInt(m[3], 10)
      const newCount = m[4] !== undefined ? parseInt(m[4], 10) : 1
      current = {
        header: line,
        oldStart,
        oldEnd: oldStart + oldCount - 1,
        newStart,
        newEnd: newStart + newCount - 1,
        lines: [],
      }
      hunks.push(current)
      oldLine = oldStart
      newLine = newStart
      continue
    }
    if (!current) continue

    if (line.startsWith("+")) {
      current.lines.push({ type: "add", oldLn: null, newLn: newLine, text: line })
      newLine++
    } else if (line.startsWith("-")) {
      current.lines.push({ type: "del", oldLn: oldLine, newLn: null, text: line })
      oldLine++
    } else {
      current.lines.push({ type: "ctx", oldLn: oldLine, newLn: newLine, text: line })
      oldLine++
      newLine++
    }
  }
  return hunks
}

export function timeAgo(dateStr: string): string {
  if (!dateStr) return ""
  const diff = Math.floor((Date.now() - new Date(dateStr).getTime()) / 1000)
  if (diff < 60) return "just now"
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  return `${Math.floor(diff / 86400)}d ago`
}
