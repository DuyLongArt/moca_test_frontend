import type { ReactNode } from 'react'

type Props = {
  children: ReactNode
  className?: string
}

export function GlassCard({ children, className = '' }: Props) {
  return (
    <div
      className={[
        'rounded-3xl border border-white/60 bg-white/70 p-6 shadow-sm backdrop-blur-md',
        'transition hover:shadow-md',
        className,
      ].join(' ')}
    >
      {children}
    </div>
  )
}
