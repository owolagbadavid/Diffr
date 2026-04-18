interface Props {
  page: number
  hasNext: boolean
  onPageChange: (page: number) => void
}

export default function Pagination({ page, hasNext, onPageChange }: Props) {
  if (page === 1 && !hasNext) return null

  return (
    <div className="flex items-center justify-center gap-4 py-4">
      <button
        disabled={page <= 1}
        onClick={() => onPageChange(page - 1)}
        className="px-4 py-2 text-sm font-medium rounded-lg border border-[var(--border)] bg-[var(--surface)] text-[var(--text)] hover:border-[var(--accent)] disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
      >
        Previous
      </button>
      <span className="text-sm text-[var(--muted)]">Page {page}</span>
      <button
        disabled={!hasNext}
        onClick={() => onPageChange(page + 1)}
        className="px-4 py-2 text-sm font-medium rounded-lg border border-[var(--border)] bg-[var(--surface)] text-[var(--text)] hover:border-[var(--accent)] disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
      >
        Next
      </button>
    </div>
  )
}
