import { BarChart3, Home, UserRound } from 'lucide-react'
import { NavLink } from 'react-router-dom'

const NAV = [
  { to: '/patient', label: 'Trang chủ', icon: Home, end: true },
  { to: '/patient/clinicians', label: 'Bác sĩ', icon: UserRound },
  { to: '/patient/results', label: 'Kết quả', icon: BarChart3 },
] as const

export function MocaBottomNav() {
  return (
    <nav
      aria-label="Điều hướng chính"
      className="fixed bottom-0 z-50 flex w-full shrink-0 items-center justify-around border-t border-outline-variant bg-surface px-2 py-1.5 pb-[calc(0.375rem+env(safe-area-inset-bottom,0px))] md:hidden"
    >
      {NAV.map((item) => {
        const Icon = item.icon
        return (
        <NavLink
          key={item.to}
          to={item.to}
          end={'end' in item ? item.end : undefined}
          className={({ isActive }) =>
            [
              'flex flex-col items-center justify-center rounded-xl px-4 py-1 transition-transform active:scale-90',
              isActive
                ? 'bg-secondary-container text-on-secondary-container'
                : 'text-on-surface-variant hover:bg-surface-container-low',
            ].join(' ')
          }
        >
          <Icon size={22} />
          <span className="text-xs">{item.label}</span>
        </NavLink>
        )
      })}
    </nav>
  )
}
