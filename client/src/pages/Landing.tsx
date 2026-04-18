import { useState } from "react"
import { useNavigate } from "react-router-dom"

export default function Landing() {
  const navigate = useNavigate()
  const [input, setInput] = useState("")
  const [error, setError] = useState("")

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const val = input.trim()
    const parts = val.split("/")
    if (parts.length !== 2 || !parts[0] || !parts[1]) {
      setError("Enter a valid owner/repo (e.g. facebook/react)")
      return
    }
    navigate(`/${parts[0]}/${parts[1]}`)
  }

  return (
    <div className="max-w-md mx-auto mt-16 text-center">
      <h2 className="text-2xl font-semibold mb-2">
        Customize how you review pull requests
      </h2>
      <p className="text-[var(--muted)] mb-6">
        Sign in with GitHub to browse your repos, or enter any public repo below.
      </p>
      <a
        href="/auth/login"
        className="inline-block px-6 py-2.5 bg-[#238636] text-white rounded-lg font-medium hover:opacity-85 transition-opacity"
      >
        Sign in with GitHub
      </a>
      <div className="flex items-center gap-3 my-5 text-[var(--muted)] text-[13px]">
        <div className="flex-1 h-px bg-[var(--border)]" />
        <span>or</span>
        <div className="flex-1 h-px bg-[var(--border)]" />
      </div>
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
      {error && (
        <p className="mt-3 text-sm text-[var(--red)]">{error}</p>
      )}
    </div>
  )
}
