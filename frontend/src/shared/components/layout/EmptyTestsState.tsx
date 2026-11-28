import { ClipboardList } from 'lucide-react'
import { Link } from 'react-router-dom'

export function EmptyTestsState() {
  return (
    <div className="flex flex-col items-center justify-center py-6 text-center">
      <div className="mb-4 flex h-24 w-24 items-center justify-center rounded-xl border border-outline-variant bg-surface-container-lowest">
        <ClipboardList size={40} className="text-primary" aria-hidden="true" />
      </div>
      <h2 className="text-lg font-semibold text-on-surface">Chưa có bài kiểm tra</h2>
      <p className="mt-1 max-w-xs text-sm text-on-surface-variant">
        Làm bài MoCA để xem kết quả tại đây.
      </p>
      <Link
        to="/patient/test"
        className="mt-4 flex h-10 items-center justify-center gap-2 rounded-lg bg-primary px-6 text-sm font-medium text-on-primary active:scale-95"
      >
        Bắt đầu kiểm tra
      </Link>
    </div>
  )
}
