import { useCallback, useEffect, useState } from 'react'
import { describeLoadError } from './load-error'

type Page<T> = { items: T[]; next_cursor: string | null }

/**
 * Cursor-paginated list: first page on mount, `loadMore` appends, `reload` restarts.
 * `fetchPage` must be referentially stable (define it at module level).
 */
export function usePagedList<T>(fetchPage: (cursor?: string) => Promise<Page<T>>, what: string) {
  const [items, setItems] = useState<T[]>([])
  const [nextCursor, setNextCursor] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const fetchInto = useCallback(
    (cursor?: string) =>
      fetchPage(cursor)
        .then((page) => {
          setItems((prev) => (cursor === undefined ? page.items : [...prev, ...page.items]))
          setNextCursor(page.next_cursor ?? null)
        })
        .catch((err: unknown) => setError(describeLoadError(err, what)))
        .finally(() => {
          setLoading(false)
          setLoadingMore(false)
        }),
    [fetchPage, what],
  )

  useEffect(() => {
    void fetchInto()
  }, [fetchInto])

  const reload = () => {
    setLoading(true)
    setError(null)
    return fetchInto()
  }

  const loadMore = () => {
    if (nextCursor === null) return Promise.resolve()
    setLoadingMore(true)
    setError(null)
    return fetchInto(nextCursor)
  }

  return { items, loading, loadingMore, error, hasMore: nextCursor !== null, reload, loadMore }
}
