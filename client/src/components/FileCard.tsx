import { useState } from "react"
import type { FileDiff } from "../types"
import DiffView from "./DiffView"

interface Props {
  file: FileDiff
  fileCache: React.RefObject<Record<string, string[]>>
  onViewFile: (contentsURL: string, filename: string) => void
}

export default function FileCard({ file, fileCache, onViewFile }: Props) {
  const [collapsed, setCollapsed] = useState(true)

  const statusClass: Record<string, string> = {
    added: "bg-[var(--green)]/15 text-[var(--green)]",
    removed: "bg-[var(--red)]/15 text-[var(--red)]",
    modified: "bg-[var(--yellow)]/15 text-[var(--yellow)]",
    renamed: "bg-[var(--accent)]/15 text-[var(--accent)]",
  }

  return (
    <div className="border border-[var(--border)] rounded-lg mb-2 overflow-hidden">
      <div
        className="flex items-center gap-3 px-3.5 py-2.5 bg-[var(--surface)] cursor-pointer select-none hover:bg-[#1c2129]"
        onClick={(e) => {
          if ((e.target as HTMLElement).closest(".file-view-btn")) return
          setCollapsed(!collapsed)
        }}
      >
        <span
          className={`text-[var(--muted)] text-[10px] shrink-0 transition-transform ${
            !collapsed ? "rotate-90" : ""
          }`}
        >
          &#9654;
        </span>
        <span
          className={`text-[11px] font-semibold uppercase px-1.5 py-0.5 rounded min-w-[4.5rem] text-center shrink-0 ${
            statusClass[file.status] ?? ""
          }`}
        >
          {file.status}
        </span>
        <span className="flex-1 font-mono text-[13px] truncate">
          {file.filename}
        </span>
        <span className="font-mono text-xs text-[var(--green)] min-w-[2.5rem] text-right shrink-0">
          +{file.additions}
        </span>
        <span className="font-mono text-xs text-[var(--red)] min-w-[2.5rem] text-right shrink-0">
          -{file.deletions}
        </span>
        {file.contents_url && (
          <button
            className="file-view-btn text-xs text-[var(--accent)] shrink-0 px-2 py-0.5 border border-[var(--border)] rounded hover:bg-[var(--accent)]/10 hover:border-[var(--accent)] transition-colors"
            onClick={() => onViewFile(file.contents_url, file.filename)}
          >
            View full
          </button>
        )}
      </div>
      {!collapsed && (
        <div className="overflow-x-auto border-t border-[var(--border)]">
          <DiffView
            patch={file.patch}
            contentsURL={file.contents_url}
            status={file.status}
            fileCache={fileCache}
          />
        </div>
      )}
    </div>
  )
}
