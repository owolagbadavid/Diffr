import { useState } from "react"
import { useNavigate } from "react-router-dom"

export default function Explore() {
  const navigate = useNavigate()
  const [input, setInput] = useState("")
  const [error, setError] = useState("")

  const recent: string[] = JSON.parse(
    localStorage.getItem("deniro_recent") || "[]"
  )

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const val = input.trim()
    const parts = val.split("/")
    if (parts.length !== 2 || !parts[0] || !parts[1]) {
      setError("Enter a valid owner/repo (e.g. facebook/react)")
      return
    }
    saveRecent(val)
    navigate(`/${parts[0]}/${parts[1]}`)
  }

  function saveRecent(repo: string) {
    const list: string[] = JSON.parse(
      localStorage.getItem("deniro_recent") || "[]"
    )
    const updated = [repo, ...list.filter((r) => r !== repo)].slice(0, 10)
    localStorage.setItem("deniro_recent", JSON.stringify(updated))
  }

  return (
    <div className="max-w-lg mx-auto mt-12">
      <h2 className="text-lg font-semibold mb-1">Explore repositories</h2>
      <p className="text-[var(--muted)] text-sm mb-6">
        Enter any public GitHub repository to browse its pull requests.
      </p>
      <form onSubmit={handleSubmit} className="flex gap-2">
        <input
          type="text"
          value={input}
          onChange={(e) => {
            setInput(e.target.value)
            setError("")
          }}
          placeholder="owner/repo"
          autoComplete="off"
          spellCheck="false"
          className="flex-1 px-3.5 py-2.5 bg-[var(--surface)] border border-[var(--border)] rounded-lg text-[var(--text)] font-mono text-[15px] outline-none focus:border-[var(--accent)] transition-colors"
        />
        <button
          type="submit"
          className="px-5 py-2.5 bg-[var(--accent)] text-white rounded-lg text-sm font-medium hover:opacity-85 transition-opacity"
        >
          Go
        </button>
      </form>
      {error && <p className="mt-3 text-sm text-[var(--red)]">{error}</p>}

      {recent.length > 0 && (
        <div className="mt-8">
          <h3 className="text-sm font-medium text-[var(--muted)] mb-3">Recent</h3>
          <div className="flex flex-col gap-1.5">
            {recent.map((r) => (
              <button
                key={r}
                onClick={() => {
                  const [o, n] = r.split("/")
                  navigate(`/${o}/${n}`)
                }}
                className="text-left px-3 py-2 bg-[var(--surface)] border border-[var(--border)] rounded-lg text-[var(--accent)] text-sm font-mono hover:border-[var(--accent)] transition-colors"
              >
                {r}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
