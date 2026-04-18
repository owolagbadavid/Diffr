import { useState } from "react"
import { fetchRawFile } from "../api"

interface Props {
  newStart: number
  newEnd: number // -1 means "remaining"
  oldStart: number
  contentsURL: string
  fileCache: React.RefObject<Record<string, string[]>>
}

export default function GapBar({ newStart, newEnd, oldStart, contentsURL, fileCache }: Props) {
  const [lines, setLines] = useState<string[] | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(false)

  if (lines) {
    const end = newEnd === -1 ? lines.length : newEnd
    if (newStart > end) return null

    return (
      <table className="diff-table">
        <tbody>
          {Array.from({ length: end - newStart + 1 }, (_, i) => {
            const lineNum = newStart + i
            const oldLn = oldStart + i
            return (
              <tr key={lineNum} className="diff-ctx">
                <td className="diff-ln">{oldLn}</td>
                <td className="diff-ln">{lineNum}</td>
                <td className="diff-code">{lines[lineNum - 1] ?? ""}</td>
              </tr>
            )
          })}
        </tbody>
      </table>
    )
  }

  const label =
    newEnd === -1
      ? `\u2195 Show remaining lines from line ${newStart}`
      : `\u2195 Show ${newEnd - newStart + 1} hidden lines (${newStart}\u2013${newEnd})`

  async function expand() {
    setLoading(true)
    try {
      const cache = fileCache.current
      if (!cache[contentsURL]) {
        const text = await fetchRawFile(contentsURL)
        cache[contentsURL] = text.split("\n")
      }
      setLines(cache[contentsURL])
    } catch {
      setError(true)
    } finally {
      setLoading(false)
    }
  }

  if (error) {
    return <div className="diff-gap-bar text-[var(--red)]">Failed to load file</div>
  }

  return (
    <div className="diff-gap-bar" onClick={expand}>
      {loading ? "Loading..." : label}
    </div>
  )
}
