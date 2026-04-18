import { useNavigate } from "react-router-dom"

export default function NotFound() {
  const navigate = useNavigate()

  return (
    <div className="text-center mt-16">
      <h2 className="text-xl font-semibold mb-2">Page not found</h2>
      <p className="text-[var(--muted)] mb-4">
        The page you're looking for doesn't exist.
      </p>
      <button
        onClick={() => navigate("/")}
        className="px-4 py-2 bg-[var(--accent)] text-white rounded-lg text-sm font-medium hover:opacity-85 transition-opacity"
      >
        Go home
      </button>
    </div>
  )
}
