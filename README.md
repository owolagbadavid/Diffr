# deniro

Customizable PR review tool. Browse your GitHub repos and review pull requests with different file organization strategies.

## Quick start

```bash
go build ./cmd/deniro/
./deniro
# open http://localhost:3000
```

## GitHub OAuth (optional)

Create a [GitHub OAuth App](https://github.com/settings/developers) with callback URL `http://localhost:3000/auth/callback`, then:

```bash
export GITHUB_CLIENT_ID=...
export GITHUB_CLIENT_SECRET=...
./deniro
```

Without OAuth, you can still browse public repos via the manual input, or set `GITHUB_TOKEN` for private repo access.

## Review strategies

| Strategy | Description |
|---|---|
| `by-size` | Groups files into small, medium, and large buckets |
| `largest-first` | Sorts files by total changes, largest first |
| `by-directory` | Groups files by parent directory |

Adding a new strategy is one file in `internal/strategy/` with an `init()` that calls `Register()`.

## Project structure

```
cmd/deniro/          CLI entrypoint (HTTP server)
internal/
  api/               Handlers, router, GitHub OAuth
  github/            GitHub REST API client
  model/             Domain types (FileDiff, PullRequest, Repository)
  strategy/          Pluggable review strategies
web/                 Embedded frontend (HTML + CSS + JS)
```

## Flags

```
-port           HTTP port (default 3000)
-token          Fallback GitHub token (default $GITHUB_TOKEN)
-client-id      GitHub OAuth client ID (default $GITHUB_CLIENT_ID)
-client-secret  GitHub OAuth client secret (default $GITHUB_CLIENT_SECRET)
-base-url       Public base URL (default http://localhost:<port>)
```
