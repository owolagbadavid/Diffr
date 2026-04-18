import { useEffect, useState } from "react"
import { useNavigate, useParams, useSearchParams } from "react-router-dom"
import type { PullRequest } from "../types"
import { fetchPRs, FetchError } from "../api"
import { timeAgo } from "../utils"
import Pagination from "../components/Pagination"

export default function PRList() {
  const { owner, repo } = useParams<{ owner: string; repo: string }>()
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const page = Number(searchParams.get("page")) || 1

  const [prs, setPRs] = useState<PullRequest[]>([])
  const [hasNext, setHasNext] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [notFound, setNotFound] = useState(false)

  useEffect(() => {
    if (!owner || !repo) return
    setLoading(true)
    setError("")
    setNotFound(false)
    fetchPRs(owner, repo, page)
      .then((res) => {
        setPRs(res.data)
        setHasNext(res.has_next)
      })
      .catch((err) => {
        if (err instanceof FetchError && err.message.includes("404")) {
          setNotFound(true)
        } else if (err instanceof FetchError) {
          setError(err.message)
        } else {
          setError("Failed to load pull requests")
        }
      })
      .finally(() => setLoading(false))
  }, [owner, repo, page])

  function handlePageChange(p: number) {
    setSearchParams({ page: String(p) })
  }

  if (notFound) {
    return (
      <div className="text-center mt-16">
        <h2 className="text-xl font-semibold mb-2">Repository not found</h2>
        <p className="text-[var(--muted)] mb-4">
          <span className="font-mono text-[var(--accent)]">{owner}/{repo}</span> doesn't exist or isn't accessible.
        </p>
        <button
          onClick={() => navigate("/explore")}
          className="px-4 py-2 bg-[var(--accent)] text-white rounded-lg text-sm font-medium hover:opacity-85 transition-opacity"
        >
          Try another repo
        </button>
      </div>
    )
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold">
          {owner}/{repo} — Pull Requests
        </h2>
      </div>

      {loading && <div className="text-center text-[var(--muted)] py-8">Loading...</div>}

      {error && (
        <div className="bg-[var(--red)]/10 border border-[var(--red)]/30 text-[var(--red)] px-4 py-3 rounded-lg text-sm mt-4">
          {error}
        </div>
      )}

      {!loading && !error && prs.length === 0 && (
        <p className="text-[var(--muted)] p-4">No open pull requests.</p>
      )}

      {!loading &&
        !error &&
        prs.map((pr) => (
          <div
            key={pr.number}
            onClick={() => navigate(`/${owner}/${repo}/${pr.number}`)}
            className="flex items-center gap-3 px-4 py-3.5 bg-[var(--surface)] border border-[var(--border)] rounded-lg mb-2 cursor-pointer hover:border-[var(--accent)] transition-colors"
          >
            <img className="w-7 h-7 rounded-full shrink-0" src={pr.avatar_url} alt="" />
            <span className="text-[var(--muted)] font-mono text-[13px] min-w-[3.5rem] shrink-0">
              #{pr.number}
            </span>
            <div className="flex-1 min-w-0 flex flex-col gap-0.5">
              <span className="font-medium truncate">{pr.title}</span>
              <span className="text-xs font-mono text-[var(--accent)] opacity-70">
                {pr.branch}
              </span>
            </div>
            <div className="flex gap-3 text-[13px] text-[var(--muted)] whitespace-nowrap shrink-0">
              {pr.draft && (
                <span className="text-xs bg-[var(--border)] text-[var(--muted)] px-2 py-0.5 rounded-full">
                  Draft
                </span>
              )}
              <span>{pr.user}</span>
              <span>{timeAgo(pr.updated_at)}</span>
            </div>
          </div>
        ))}

      {!loading && !error && (
        <Pagination page={page} hasNext={hasNext} onPageChange={handlePageChange} />
      )}
    </div>
  )
}
