import { useEffect, useState } from "react"
import { useNavigate, useSearchParams } from "react-router-dom"
import type { Repository } from "../types"
import { fetchRepos, FetchError } from "../api"
import { timeAgo } from "../utils"
import Pagination from "../components/Pagination"

export default function Repos() {
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const page = Number(searchParams.get("page")) || 1

  const [repos, setRepos] = useState<Repository[]>([])
  const [hasNext, setHasNext] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [filter, setFilter] = useState("")

  useEffect(() => {
    setLoading(true)
    setError("")
    fetchRepos(page)
      .then((res) => {
        setRepos(res.data)
        setHasNext(res.has_next)
      })
      .catch((err) => {
        if (err instanceof FetchError) setError(err.message)
        else setError("Failed to load repositories")
      })
      .finally(() => setLoading(false))
  }, [page])

  const filtered = filter
    ? repos.filter(
        (r) =>
          r.full_name.toLowerCase().includes(filter.toLowerCase()) ||
          (r.description ?? "").toLowerCase().includes(filter.toLowerCase()) ||
          (r.language ?? "").toLowerCase().includes(filter.toLowerCase())
      )
    : repos

  function handlePageChange(p: number) {
    setSearchParams({ page: String(p) })
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-semibold">Your repositories</h2>
      </div>
      <div className="mb-3">
        <input
          type="text"
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          placeholder="Filter repositories..."
          autoComplete="off"
          spellCheck="false"
          className="w-full px-3 py-2 bg-[var(--surface)] border border-[var(--border)] rounded-lg text-[var(--text)] text-sm outline-none focus:border-[var(--accent)] transition-colors"
        />
      </div>

      {loading && <div className="text-center text-[var(--muted)] py-8">Loading...</div>}

      {error && (
        <div className="bg-[var(--red)]/10 border border-[var(--red)]/30 text-[var(--red)] px-4 py-3 rounded-lg text-sm mt-4">
          {error}
        </div>
      )}

      {!loading && !error && filtered.length === 0 && (
        <p className="text-[var(--muted)] p-4">No repositories found.</p>
      )}

      {!loading &&
        !error &&
        filtered.map((repo) => (
          <div
            key={repo.full_name}
            onClick={() => {
              const [owner, name] = repo.full_name.split("/")
              navigate(`/${owner}/${name}`)
            }}
            className="flex items-center gap-3 px-4 py-3 bg-[var(--surface)] border border-[var(--border)] rounded-lg mb-2 cursor-pointer hover:border-[var(--accent)] transition-colors"
          >
            <img
              className="w-7 h-7 rounded-full shrink-0"
              src={repo.owner_avatar}
              alt=""
            />
            <div className="flex-1 min-w-0 flex flex-col gap-0.5">
              <span className="font-medium text-[15px] text-[var(--accent)]">
                {repo.full_name}
              </span>
              {repo.private && (
                <span className="inline-block text-[11px] px-1.5 rounded-full bg-[var(--border)] text-[var(--muted)] w-fit">
                  Private
                </span>
              )}
              <span className="text-[13px] text-[var(--muted)] truncate">
                {repo.description || ""}
              </span>
            </div>
            <div className="flex gap-3 text-[13px] text-[var(--muted)] whitespace-nowrap shrink-0">
              {repo.language && (
                <span className="text-[var(--text)] font-medium">{repo.language}</span>
              )}
              {repo.stars > 0 && (
                <span className="text-[var(--yellow)]">{"\u2605"} {repo.stars}</span>
              )}
              <span>{timeAgo(repo.updated_at)}</span>
            </div>
          </div>
        ))}

      {!loading && !error && (
        <Pagination page={page} hasNext={hasNext} onPageChange={handlePageChange} />
      )}
    </div>
  )
}
