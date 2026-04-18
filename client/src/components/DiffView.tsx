import { parseHunks } from "../utils"
import GapBar from "./GapBar"

interface Props {
  patch: string
  contentsURL: string
  status: string
  fileCache: React.RefObject<Record<string, string[]>>
}

export default function DiffView({ patch, contentsURL, status, fileCache }: Props) {
  if (!patch) {
    return <div className="p-4 text-[var(--muted)] text-[13px] text-center">Binary file or no diff available</div>
  }

  const hunks = parseHunks(patch)
  const canExpand = contentsURL && status !== "removed"
  let prevNewEnd = 0
  let prevOldEnd = 0

  const elements: React.ReactNode[] = []

  hunks.forEach((hunk, hi) => {
    const gapNewStart = prevNewEnd + 1
    const gapNewEnd = hunk.newStart - 1
    const gapOldStart = prevOldEnd + 1
    const gapCount = gapNewEnd - gapNewStart + 1

    if (gapCount > 0 && canExpand) {
      elements.push(
        <GapBar
          key={`gap-${hi}`}
          newStart={gapNewStart}
          newEnd={gapNewEnd}
          oldStart={gapOldStart}
          contentsURL={contentsURL}
          fileCache={fileCache}
        />
      )
    }

    elements.push(
      <div key={`hh-${hi}`} className="diff-hunk-header">
        {hunk.header}
      </div>
    )

    elements.push(
      <table key={`ht-${hi}`} className="diff-table">
        <tbody>
          {hunk.lines.map((l, li) => (
            <tr
              key={li}
              className={
                l.type === "add" ? "diff-add" : l.type === "del" ? "diff-del" : "diff-ctx"
              }
            >
              <td className="diff-ln">{l.oldLn ?? ""}</td>
              <td className="diff-ln">{l.newLn ?? ""}</td>
              <td className="diff-code">{l.text}</td>
            </tr>
          ))}
        </tbody>
      </table>
    )

    prevNewEnd = hunk.newEnd
    prevOldEnd = hunk.oldEnd
  })

  if (hunks.length > 0 && canExpand) {
    elements.push(
      <GapBar
        key="gap-end"
        newStart={prevNewEnd + 1}
        newEnd={-1}
        oldStart={prevOldEnd + 1}
        contentsURL={contentsURL}
        fileCache={fileCache}
      />
    )
  }

  return <>{elements}</>
}
