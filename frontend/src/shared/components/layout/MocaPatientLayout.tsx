import type { ReactNode } from 'react'
import type { BreadcrumbCrumb } from '../../lib/breadcrumbs'
import { MocaAppHeader } from './MocaAppHeader'
import { MocaBottomNav } from './MocaBottomNav'
import { MocaBreadcrumb } from './MocaBreadcrumb'
import '../../../styles/stitch-elderly.css'

type Props = {
  children: ReactNode
  title?: string
  breadcrumbs?: BreadcrumbCrumb[]
  /** Keep page fixed; disable main scroll (home, empty states) */
  fitViewport?: boolean
}

export function MocaPatientLayout({
  children,
  title,
  breadcrumbs,
  fitViewport = false,
}: Props) {
  return (
    <div className="app-shell elderly-layout bg-surface text-on-surface">
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-3 focus:z-50 focus:rounded-xl focus:bg-white focus:px-4 focus:py-2 focus:text-sm focus:font-semibold focus:text-slate-900"
      >
        Đi đến nội dung chính
      </a>
      <MocaAppHeader title={title} />
      <main
        id="main-content"
        tabIndex={-1}
        className={[
          'app-shell__main px-[var(--stitch-margin-mobile)] py-4',
          'pb-[calc(4.5rem+env(safe-area-inset-bottom,0px))] md:pb-4',
          fitViewport ? 'flex flex-col overflow-hidden' : '',
        ].join(' ')}
      >
        <div className="mx-auto flex w-full max-w-7xl min-h-0 flex-1 flex-col">
          <MocaBreadcrumb items={breadcrumbs} className="mb-3 shrink-0" />
          {children}
        </div>
      </main>
      <MocaBottomNav />
    </div>
  )
}
