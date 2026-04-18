import { useEffect, useRef, useState } from "react"
import { useNavigate, useParams } from "react-router-dom"
import type { PRFilesResponse, Strategy } from "../types"
import { fetchPRFiles, fetchStrategies, FetchError } from "../api"
import FileCard from "../components/FileCard"
import FileViewer from "../components/FileViewer"

export default function PRDetail() {
  const { owner, repo, number } = useParams<{
    owner: string
    repo: string
    number: string
  }>()
  const navigate = useNavigate()
  const prNumber = Number(number)

  const [data, setData] = useState<PRFilesResponse | null>(null)
  const [strategies, setStrategies] = useState<Strategy[]>([])
  const [strategy, setStrategy] = useState("by-size")
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [notFound, setNotFound] = useState(false)

  const [viewerFile, setViewerFile] = useState<{
    url: string
    name: string
  } | null>(null)

  const fileCache = useRef<Record<string, string[]>>({})

  useEffect(() => {
    fetchStrategies()
      .then(setStrategies)
      .catch(() => {})
  }, [])

  useEffect(() => {
    if (!owner || !repo || !prNumber) return
    setLoading(true)
    setError("")
    setNotFound(false)
    fetchPRFiles(owner, repo, prNumber, strategy)
      .then(setData)
      .catch((err) => {
        if (err instanceof FetchError && err.message.includes("404")) {
          setNotFound(true)
        } else if (err instanceof FetchError) {
          setError(err.message)
        } else {
          setError("Failed to load PR files")
        }
      })
      .finally(() => setLoading(false))
  }, [owner, repo, prNumber, strategy])

  if (notFound) {
    return (
      <div className="text-center mt-16">
        <h2 className="text-xl font-semibold mb-2">Pull request not found</h2>
        <p className="text-[var(--muted)] mb-4">
          PR #{number} doesn't exist in{" "}
          <span className="font-mono text-[var(--accent)]">
            {owner}/{repo}
          </span>
          .
        </p>
        <button
          onClick={() => navigate(`/${owner}/${repo}`)}
          className="px-4 py-2 bg-[var(--accent)] text-white rounded-lg text-sm font-medium hover:opacity-85 transition-opacity"
        >
          Back to PRs
        </button>
      </div>
    )
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold">
          #{number} — {data ? `${data.total_files} files` : ""}
        </h2>
      </div>

      <div className="flex items-center gap-2 mb-5">
        <label className="text-sm text-[var(--muted)]">Strategy:</label>
        <select
          value={strategy}
          onChange={(e) => setStrategy(e.target.value)}
          className="px-3 py-2 bg-[var(--surface)] border border-[var(--border)] rounded-lg text-[var(--text)] text-sm outline-none cursor-pointer focus:border-[var(--accent)]"
        >
          {strategies.length > 0
            ? strategies.map((s) => (
                <option key={s.name} value={s.name}>
                  {s.name} — {s.description}
                </option>
              ))
            : <option value="by-size">by-size</option>
          }
        </select>
      </div>

      {loading && (
        <div className="text-center text-[var(--muted)] py-8">Loading...</div>
      )}

      {error && (
        <div className="bg-[var(--red)]/10 border border-[var(--red)]/30 text-[var(--red)] px-4 py-3 rounded-lg text-sm mt-4">
          {error}
        </div>
      )}

      {!loading &&
        !error &&
        data?.groups.map((group) => (
          <div key={group.name} className="mb-6">
            <div className="text-[13px] font-semibold text-[var(--muted)] uppercase tracking-wider pb-2 border-b border-[var(--border)] mb-2">
              {group.name} ({group.files.length})
            </div>
            {group.files.map((f) => (
              <FileCard
                key={f.filename}
                file={f}
                fileCache={fileCache}
                onViewFile={(url, name) => setViewerFile({ url, name })}
              />
            ))}
          </div>
        ))}

      {viewerFile && (
        <FileViewer
          contentsURL={viewerFile.url}
          filename={viewerFile.name}
          fileCache={fileCache}
          onClose={() => setViewerFile(null)}
        />
      )}
    </div>
  )
}
