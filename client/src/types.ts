export interface User {
  logged_in: boolean
  login?: string
  name?: string
  avatar_url?: string
}

export interface Repository {
  name: string
  full_name: string
  description: string
  private: boolean
  language: string
  stars: number
  open_issues: number
  updated_at: string
  owner_login: string
  owner_avatar: string
}

export interface PullRequest {
  number: number
  title: string
  state: string
  user: string
  avatar_url: string
  branch: string
  created_at: string
  updated_at: string
  draft: boolean
}

export interface FileDiff {
  filename: string
  status: string
  additions: number
  deletions: number
  patch: string
  blob_url: string
  contents_url: string
}

export interface FileGroup {
  name: string
  files: FileDiff[]
}

export interface Strategy {
  name: string
  description: string
}

export interface PaginatedResponse<T> {
  data: T[]
  page: number
  has_next: boolean
  has_prev: boolean
}

export interface PRFilesResponse {
  owner: string
  repo: string
  number: number
  strategy: string
  total_files: number
  groups: FileGroup[]
}

export interface ApiError {
  error: string
  status: number
}
