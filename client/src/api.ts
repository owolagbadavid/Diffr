import type {
  User,
  Repository,
  PullRequest,
  Strategy,
  PRFilesResponse,
  PaginatedResponse,
} from "./types"

class FetchError extends Error {
  status: number
  constructor(message: string, status: number) {
    super(message)
    this.status = status
  }
}

async function request<T>(url: string): Promise<T> {
  const res = await fetch(url, { credentials: "include" })
  if (!res.ok) {
    let message = res.statusText
    try {
      const body = await res.json()
      if (body.error) message = body.error
    } catch {}
    throw new FetchError(message, res.status)
  }
  return res.json()
}

export async function fetchUser(): Promise<User> {
  return request<User>("/api/user")
}

export async function fetchRepos(
  page = 1
): Promise<PaginatedResponse<Repository>> {
  return request<PaginatedResponse<Repository>>(
    `/api/user/repos?page=${page}`
  )
}

export async function fetchPRs(
  owner: string,
  repo: string,
  page = 1
): Promise<PaginatedResponse<PullRequest>> {
  return request<PaginatedResponse<PullRequest>>(
    `/api/repos/${owner}/${repo}/pulls?page=${page}`
  )
}

export async function fetchPRFiles(
  owner: string,
  repo: string,
  number: number,
  strategy = "by-size"
): Promise<PRFilesResponse> {
  return request<PRFilesResponse>(
    `/api/repos/${owner}/${repo}/pulls/${number}/files?strategy=${strategy}`
  )
}

export async function fetchStrategies(): Promise<Strategy[]> {
  return request<Strategy[]>("/api/strategies")
}

export async function fetchRawFile(contentsURL: string): Promise<string> {
  const res = await fetch(`/api/raw?url=${encodeURIComponent(contentsURL)}`, {
    credentials: "include",
  })
  if (!res.ok) throw new FetchError("Failed to load file", res.status)
  return res.text()
}

export { FetchError }
