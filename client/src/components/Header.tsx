import { Link, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { logout } from "../api";

export default function Header() {
  const { user, clearUser } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();

  const crumbs = buildCrumbs(location.pathname);

  return (
    <header className="flex items-center justify-between px-6 py-3 border-b border-[var(--border)] bg-[var(--surface)] sticky top-0 z-50">
      <div className="flex items-center gap-4">
        <Link to="/" className="text-lg font-semibold tracking-tight hover:opacity-80">
          Diffr
        </Link>
        <nav className="flex items-center gap-1.5 text-[13px]">
          {crumbs.map((c, i) => (
            <span key={i} className="flex items-center gap-1.5">
              {i > 0 && <span className="text-[var(--muted)]">/</span>}
              {c.to ? (
                <Link to={c.to} className="text-[var(--accent)] hover:underline">
                  {c.label}
                </Link>
              ) : (
                <span className="text-[var(--text)] font-medium">{c.label}</span>
              )}
            </span>
          ))}
        </nav>
      </div>
      <div className="flex items-center gap-2 text-[13px]">
        <Link
          to="/explore"
          className={`px-3 py-1.5 rounded-lg transition-colors ${
            location.pathname === "/explore"
              ? "bg-[var(--accent)]/10 text-[var(--accent)]"
              : "text-[var(--muted)] hover:text-[var(--text)]"
          }`}
        >
          Explore
        </Link>
        {user?.logged_in ? (
          <>
            <Link to="/repos" className="text-[var(--muted)] hover:text-[var(--text)] px-2">
              My Repos
            </Link>
            <img className="w-6 h-6 rounded-full" src={user.avatar_url} alt="" />
            <span className="font-medium">{user.login}</span>
            <button
              onClick={() => {
                logout().finally(() => {
                  clearUser();
                  navigate("/");
                });
              }}
              className="text-[var(--accent)] hover:underline"
            >
              Logout
            </button>
          </>
        ) : (
          <a
            href="/auth/login"
            className="px-3 py-1.5 bg-[var(--accent)] text-white rounded-lg text-[13px] font-medium hover:opacity-85"
          >
            Sign in
          </a>
        )}
      </div>
    </header>
  );
}

interface Crumb {
  label: string;
  to?: string;
}

function buildCrumbs(pathname: string): Crumb[] {
  const parts = pathname.split("/").filter(Boolean);

  if (parts[0] === "repos") return [{ label: "Repos" }];
  if (parts[0] === "explore") return [{ label: "Explore" }];

  if (parts.length >= 2) {
    const owner = parts[0];
    const repo = parts[1];
    const crumbs: Crumb[] = [{ label: `${owner}/${repo}`, to: `/${owner}/${repo}` }];
    if (parts.length === 3) {
      crumbs[0].to = `/${owner}/${repo}`;
      crumbs.push({ label: `#${parts[2]}` });
    } else {
      delete crumbs[0].to;
    }
    return crumbs;
  }

  return [];
}
