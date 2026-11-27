import { useQueryClient } from '@tanstack/react-query'
import { Loader2, RefreshCw } from 'lucide-react'
import type { ReactNode } from 'react'
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
          onClick={() => void queryClient.refetchQueries({ type: 'active' })}
          className="mt-3 inline-flex h-9 items-center gap-2 rounded-lg border border-red-300 bg-white px-3 text-sm font-medium text-red-800 transition-colors hover:bg-red-100"
        >
          <RefreshCw size={15} />
          Thử lại
        </button>
      </div>
    )
  }

  if (isEmpty && empty) return <>{empty}</>

  return <>{children}</>
}
