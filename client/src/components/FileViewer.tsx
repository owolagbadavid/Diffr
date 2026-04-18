import { useEffect, useState } from "react"
import { fetchRawFile } from "../api"

interface Props {
  contentsURL: string
  filename: string
  fileCache: React.RefObject<Record<string, string[]>>
  onClose: () => void
}

export default function FileViewer({ contentsURL, filename, fileCache, onClose }: Props) {
  const [lines, setLines] = useState<string[] | null>(null)
  const [error, setError] = useState(false)

  useEffect(() => {
    const cache = fileCache.current
    if (cache[contentsURL]) {
      setLines(cache[contentsURL])
      return
    }

    fetchRawFile(contentsURL)
      .then((text) => {
        const l = text.split("\n")
        cache[contentsURL] = l
        setLines(l)
      })
      .catch(() => setError(true))
  }, [contentsURL, fileCache])

  return (
    <div className="fixed top-0 right-0 w-[55%] h-screen bg-[var(--bg)] border-l border-[var(--border)] z-[100] flex flex-col shadow-[-4px_0_24px_rgba(0,0,0,0.5)]">
      <div className="flex items-center justify-between px-4 py-3 bg-[var(--surface)] border-b border-[var(--border)] font-mono text-[13px] font-medium">
        <span className="truncate">{filename}</span>
        <button
          onClick={onClose}
          className="text-[var(--accent)] hover:underline bg-transparent text-[13px]"
        >
          Close
        </button>
      </div>
      <div className="flex-1 overflow-auto">
        {error ? (
          <div className="p-4 text-[var(--muted)] text-[13px] text-center">
            Failed to load file
          </div>
        ) : !lines ? (
          <div className="p-4 text-[var(--muted)] text-[13px] text-center">
            Loading...
          </div>
        ) : (
          <table className="diff-table file-table">
            <tbody>
              {lines.map((line, i) => (
                <tr key={i} className="hover:bg-[var(--surface)]">
                  <td className="diff-ln">{i + 1}</td>
                  <td className="diff-code text-[var(--text)]">{line}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}
