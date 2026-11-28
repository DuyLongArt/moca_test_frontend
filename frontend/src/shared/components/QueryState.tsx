import { useQueryClient } from '@tanstack/react-query'
import { Loader2, RefreshCw } from 'lucide-react'
import { useState, type ReactNode } from 'react'
import { getApiErrorMessage } from '../utils/apiError'

type Props = {
  isLoading: boolean
  error: Error | null
  children: ReactNode
  empty?: ReactNode
  isEmpty?: boolean
}

export function QueryState({ isLoading, error, children, empty, isEmpty }: Props) {
  const queryClient = useQueryClient()
  const [retrying, setRetrying] = useState(false)

  const retry = async () => {
    setRetrying(true)
    try {
      await queryClient.refetchQueries({ type: 'active' })
    } finally {
      setRetrying(false)
    }
  }

  if (isLoading) {
    return (
      <div
        role="status"
        aria-live="polite"
        className="flex items-center justify-center gap-2 py-12 text-slate-500"
      >
        <Loader2 size={20} className="animate-spin" />
        Đang tải…
      </div>
    )
  }

  if (error) {
    return (
      <div
        role="alert"
        className="rounded-2xl border border-red-200 bg-red-50 p-4 text-sm text-red-800"
      >
        <p>Không tải được dữ liệu: {getApiErrorMessage(error, 'Vui lòng thử lại sau.')}</p>
        <button
          type="button"
          onClick={() => void retry()}
          disabled={retrying}
          aria-busy={retrying}
          className="mt-3 inline-flex h-9 items-center gap-2 rounded-lg border border-red-300 bg-white px-3 text-sm font-medium text-red-800 transition-colors hover:bg-red-100 disabled:opacity-60"
        >
          <RefreshCw size={15} className={retrying ? 'animate-spin' : undefined} />
          {retrying ? 'Đang thử lại…' : 'Thử lại'}
        </button>
      </div>
    )
  }

  if (isEmpty && empty) return <>{empty}</>

  return <>{children}</>
}
