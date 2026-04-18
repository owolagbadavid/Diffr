import { Routes, Route, Navigate } from "react-router-dom"
import { useAuth } from "./context/AuthContext"
import Header from "./components/Header"
import Landing from "./pages/Landing"
import Repos from "./pages/Repos"
import Explore from "./pages/Explore"
import PRList from "./pages/PRList"
import PRDetail from "./pages/PRDetail"
import NotFound from "./pages/NotFound"

export default function App() {
  const { user, loading } = useAuth()

  if (loading) {
    return (
      <div className="min-h-screen bg-[var(--bg)] text-[var(--text)] flex items-center justify-center">
        <span className="text-[var(--muted)]">Loading...</span>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-[var(--bg)] text-[var(--text)]">
      <Header />
      <main className="max-w-[1060px] mx-auto px-8 py-8">
        <Routes>
          <Route
            path="/"
            element={user ? <Navigate to="/repos" replace /> : <Landing />}
          />
          <Route path="/repos" element={<Repos />} />
          <Route path="/explore" element={<Explore />} />
          <Route path="/:owner/:repo" element={<PRList />} />
          <Route path="/:owner/:repo/:number" element={<PRDetail />} />
          <Route path="*" element={<NotFound />} />
        </Routes>
      </main>
    </div>
  )
}
