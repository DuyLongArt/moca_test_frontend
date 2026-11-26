import { Brain, CalendarDays, ClipboardList, UserRound } from 'lucide-react'
import { Link } from 'react-router-dom'
import { MocaPatientLayout } from '../../shared/components/layout/MocaPatientLayout'
import { useHealth } from '../../shared/hooks/useHealth'

const MENU = [
  {
    to: '/patient/test',
    label: 'Làm test MoCA',
    desc: 'Bài kiểm tra 9 phần',
    icon: Brain,
    color: 'bg-primary text-on-primary',
  },
  {
    to: '/patient/results',
    label: 'Kết quả',
    desc: 'Điểm tạm & chính thức',
    icon: ClipboardList,
    color: 'bg-secondary text-on-primary',
  },
  {
    to: '/patient/appointments',
    label: 'Lịch khám',
    desc: 'Lịch hẹn sắp tới',
    icon: CalendarDays,
    color: 'bg-surface-container-high text-on-surface',
  },
  {
    to: '/patient/clinicians',
    label: 'Đổi bác sĩ',
    desc: 'Bác sĩ phụ trách',
    icon: UserRound,
    color: 'bg-surface-container-high text-on-surface',
  },
]

export function PatientHomePage() {
  const showDevHealth = import.meta.env.DEV
  const { data: health, isError } = useHealth({ enabled: showDevHealth })

  return (
    <MocaPatientLayout fitViewport>
      <section className="shrink-0 space-y-1">
        <p className="text-xs font-medium uppercase tracking-wider text-secondary">
          Tổng quan
        </p>
        <h2 className="text-xl font-semibold text-on-surface md:text-2xl">
          Chào mừng trở lại
        </h2>
        {showDevHealth && !isError && health && (
          <p className="text-xs text-secondary">
            API · {health.service} — {health.status}
          </p>
        )}
        {showDevHealth && isError && (
          <p className="text-xs text-error">Không kết nối được API</p>
        )}
      </section>

      <div className="mt-4 grid grid-cols-2 gap-3">
        {MENU.map((item) => (
          <Link
            key={item.to}
            to={item.to}
            className="group flex flex-col items-start gap-2 rounded-xl border border-outline-variant bg-surface-container-lowest p-3 text-left transition-all hover:border-primary/40 active:scale-[0.98] sm:flex-row sm:gap-3 sm:p-4"
          >
            <div
              className={`flex h-12 w-12 shrink-0 items-center justify-center rounded-lg ${item.color} sm:h-15 sm:w-15`}
            >
              <item.icon size={20} />
            </div>
            <div>
              <h3 className="text-sm font-semibold text-on-surface sm:text-base">
                {item.label}
              </h3>
              <p className="mt-0.5 text-[11px] text-on-surface-variant sm:text-xs">
                {item.desc}
              </p>
            </div>
          </Link>
        ))}
      </div>
    </MocaPatientLayout>
  )
}
