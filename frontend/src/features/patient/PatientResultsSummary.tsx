import { useRef, useState } from 'react'
import { AlertTriangle, CheckCircle2, FileDown } from 'lucide-react'
import type { TestSessionSummary } from '../../shared/types/session'
import { formatDateVi } from '../../shared/utils/format'

type Props = {
  session: TestSessionSummary
  patientName?: string
}

export function PatientResultsSummary({
  session,
  patientName = 'Bệnh nhân',
}: Props) {
  const score = session.finalScore ?? session.provisionalScore ?? 0
  const isNormal = score >= 26
  const isPending = session.status === 'PENDING_REVIEW'
  const progressWidth = Math.min(100, Math.max(0, (score / 30) * 100))

  const rootRef = useRef<HTMLDivElement>(null)
  const [exporting, setExporting] = useState(false)
  const [exportError, setExportError] = useState<string | null>(null)

  async function handleExportPdf() {
    const el = rootRef.current
    if (!el || exporting) return
    setExporting(true)
    setExportError(null)

    // html2canvas-pro renders the glass cards' backdrop-filter: blur(12px),
    // which makes the whole PDF look soft. Disable it for the capture.
    const noBackdrop = document.createElement('style')
    noBackdrop.textContent =
      '* { backdrop-filter: none !important; -webkit-backdrop-filter: none !important; }'
    document.head.appendChild(noBackdrop)

    try {
      const [{ jsPDF }, { default: html2canvas }] = await Promise.all([
        import('jspdf'),
        import('html2canvas-pro'),
      ])

      // Target ~240 DPI at A4 print width so text stays crisp.
      const targetWidthPx = 1800
      const scale = Math.max(2, Math.ceil(targetWidthPx / Math.max(el.offsetWidth, 1)))
      const canvas = await html2canvas(el, {
        scale,
        useCORS: true,
        backgroundColor: '#ffffff',
      })
      const imgData = canvas.toDataURL('image/png')
      const pdf = new jsPDF({ unit: 'mm', format: 'a4', orientation: 'portrait' })
      const pageWidth = pdf.internal.pageSize.getWidth()
      const pageHeight = pdf.internal.pageSize.getHeight()
      const margin = 10
      const imgWidth = pageWidth - margin * 2
      const imgHeight = (canvas.height * imgWidth) / canvas.width
      const printableHeight = pageHeight - margin * 2

      let heightLeft = imgHeight
      let position = margin
      pdf.addImage(imgData, 'PNG', margin, position, imgWidth, imgHeight)
      heightLeft -= printableHeight
      while (heightLeft > 0) {
        position = margin - (imgHeight - heightLeft)
        pdf.addPage()
        pdf.addImage(imgData, 'PNG', margin, position, imgWidth, imgHeight)
        heightLeft -= printableHeight
      }
      pdf.save(`moca-ket-qua-${session.id}.pdf`)
    } catch {
      setExportError('Không xuất được PDF. Vui lòng thử lại.')
    } finally {
      document.head.removeChild(noBackdrop)
      setExporting(false)
    }
  }

  return (
    <div ref={rootRef} className="space-y-4 stitch-fade-in">
      {/* Hero score */}
      <section className="grid grid-cols-1 gap-3 md:grid-cols-3">
        <div className="stitch-glass-card relative overflow-hidden rounded-xl p-4 md:col-span-2">
          <p className="text-xs font-medium uppercase tracking-wider text-on-surface-variant">
            Tổng điểm
          </p>
          <div className="mt-0.5 flex items-baseline gap-1">
            <span className="text-5xl font-bold leading-none text-primary">
              {score}
            </span>
            <span className="text-xl font-medium text-outline">/ 30</span>
          </div>
          {isPending ? (
            <div className="mt-2 inline-flex items-center gap-2 rounded-full bg-amber-100 px-3 py-1 text-xs font-semibold text-amber-800">
              <AlertTriangle size={14} />
              Chờ bác sĩ duyệt
            </div>
          ) : (
            <div className="mt-2 inline-flex items-center gap-2 rounded-full bg-secondary-container px-3 py-1 text-xs font-semibold text-on-secondary-container">
              <CheckCircle2 size={14} />
              {session.classification ?? (isNormal ? 'Nhận thức bình thường' : 'Dưới ngưỡng bình thường')}
            </div>
          )}
          <div className="mt-4 space-y-1.5">
            <div className="flex justify-between text-[10px] font-medium">
              <span>Thang điểm</span>
              <span>≥26 Bình thường</span>
            </div>
            <div
              className="flex h-2 w-full overflow-hidden rounded-full bg-surface-container"
              role="progressbar"
              aria-label="Điểm MoCA"
              aria-valuenow={score}
              aria-valuemin={0}
              aria-valuemax={30}
              aria-valuetext={isPending ? `${score}/30, chờ bác sĩ duyệt` : `${score}/30 điểm`}
            >
              <div
                className="h-full bg-primary"
                style={{ width: `${progressWidth}%` }}
              />
            </div>
          </div>
        </div>

        <div className="stitch-glass-card rounded-xl p-4">
          <h3 className="text-sm font-semibold text-on-surface">Thông tin phiên</h3>
          <div className="mt-2 space-y-2 text-sm">
            <div>
              <p className="text-[10px] text-outline">Bệnh nhân</p>
              <p className="font-semibold">{patientName}</p>
            </div>
            <div>
              <p className="text-[10px] text-outline">Ngày làm bài</p>
              <p className="font-semibold">{formatDateVi(session.submittedAt)}</p>
            </div>
            <div>
              <p className="text-[10px] text-outline">Mã đề</p>
              <p className="font-semibold">{session.setId}</p>
            </div>
          </div>
        </div>
      </section>

      {/* Actions */}
      {!isPending && (
        <section>
          <button
            type="button"
            onClick={handleExportPdf}
            disabled={exporting}
            className="flex w-full items-center justify-center gap-2 rounded-xl border-2 border-primary py-3 text-sm font-semibold text-primary active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-50"
          >
            <FileDown size={18} />
            {exporting ? 'Đang xuất…' : 'Xuất PDF'}
          </button>
          {exportError && (
            <p role="alert" className="mt-2 text-center text-sm font-medium text-error">
              {exportError}
            </p>
          )}
        </section>
      )}
    </div>
  )
}
