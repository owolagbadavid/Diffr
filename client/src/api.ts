import type {
  User,
  Repository,
  PullRequest,
  Strategy,
  PRFilesResponse,
  PaginatedResponse,
  CodeSearchResult,
} from "./types"

const BASE_URL = "";

class FetchError extends Error {
  status: number
  constructor(message: string, status: number) {
    super(message)
    this.status = status
  }
}

async function request<T>(url: string): Promise<T> {
  const res = await fetch(BASE_URL+url, { credentials: "include" })
  if (!res.ok) {
    let message = res.statusText
    try {
      const body = await res.json()
      if (body.error) message = body.error
    } catch { /* empty */ }
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
  strategy = "default"
): Promise<PRFilesResponse> {
  return request<PRFilesResponse>(
    `/api/repos/${owner}/${repo}/pulls/${number}/files?strategy=${strategy}`
  )
}

export async function fetchStrategies(): Promise<Strategy[]> {
  return request<Strategy[]>("/api/strategies")
}

export async function fetchRawFile(contentsURL: string): Promise<string> {
  const res = await fetch(BASE_URL+`/api/raw?url=${encodeURIComponent(contentsURL)}`, {
    credentials: "include",
  })
  if (!res.ok) throw new FetchError("Failed to load file", res.status)
  return res.text()
}

export async function searchCode(
  owner: string,
  repo: string,
  query: string
): Promise<CodeSearchResult> {
  return request<CodeSearchResult>(
    `/api/repos/${owner}/${repo}/search?q=${encodeURIComponent(query)}`
  )
}

export async function exchangeCode(code: string): Promise<void> {
  const res = await fetch(BASE_URL+`/auth/callback?code=${encodeURIComponent(code)}`, {
    method: "POST",
    credentials: "include",
  })
  if (!res.ok) {
    let message = res.statusText
    try {
      const body = await res.json()
      if (body.error) message = body.error
    } catch { /* empty */ }
    throw new FetchError(message, res.status)
  }
}

export async function logout(): Promise<void> {
  await fetch(BASE_URL+"/auth/logout", {
    method: "POST",
    credentials: "include",
  })
}

export { FetchError }
