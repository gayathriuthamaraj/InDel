import type { ReactNode } from 'react'
import { NavLink } from 'react-router-dom'
import Navbar from './Navbar'
import { useLocalization, type Language, type TranslationKey } from '../../context/LocalizationContext'

type NavSection = 'main' | 'analysis' | 'claims'

type NavItem = {
  to: string
  labelKey: TranslationKey
  icon: string
  section: NavSection
}

const insurerNav: NavItem[] = [
  // Main Dashboard
  { to: '/', labelKey: 'sidebar.overview', icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6', section: 'main' },
  { to: '/plan-status', labelKey: 'sidebar.planStatus', icon: 'M12 8c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2zm0 10c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z', section: 'main' },
  // Analysis & Monitoring
  { to: '/loss-ratio', labelKey: 'sidebar.lossRatio', icon: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z', section: 'analysis' },
  { to: '/forecast', labelKey: 'sidebar.forecast', icon: 'M13 7h8m0 0v8m0-8l-8 8-4-4-6 6', section: 'analysis' },
  // Claims Management
  { to: '/claims', labelKey: 'sidebar.claimsMenu', icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4', section: 'claims' },
  { to: '/fraud-queue', labelKey: 'sidebar.fraudQueue', icon: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z', section: 'claims' },
]

function navClass(isActive: boolean) {
  return [
    'group flex items-center gap-3 px-4 py-2.5 text-sm font-medium border-l-2 transition-none',
    isActive
      ? 'bg-[var(--brand-soft)] dark:bg-slate-800 border-[var(--brand-primary)] text-slate-900 dark:text-white'
      : 'border-transparent text-slate-500 hover:text-slate-900 dark:hover:text-white hover:bg-slate-50 dark:hover:bg-slate-800/50',
  ].join(' ')
}

export default function Sidebar({ children }: { children: ReactNode }) {
  const { language, setLanguage, t } = useLocalization()
  const mainItems = insurerNav.filter(item => item.section === 'main')
  const analysisItems = insurerNav.filter(item => item.section === 'analysis')
  const claimsItems = insurerNav.filter(item => item.section === 'claims')

  const renderNavSection = (label: string, items: NavItem[]) => (
    <section>
      <p className="mb-2 px-6 text-[9px] font-black uppercase tracking-[0.3em] text-slate-400">{label}</p>
      <nav className="flex flex-col">
            {items.map((item) => (
          <NavLink key={item.to} to={item.to} end={item.to === '/'} className={({ isActive }) => navClass(isActive)}>
            <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d={item.icon} />
            </svg>
            {t(item.labelKey)}
          </NavLink>
        ))}
      </nav>
    </section>
  )

  return (
    <div className="flex min-h-screen bg-[var(--bg-main)]">
      {/* Enterprise Sidebar */}
      <aside className="fixed left-0 top-0 z-50 h-screen w-64 border-r border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900">
        <div className="flex h-full flex-col py-8">
          <div className="mb-10 flex items-center gap-3 px-6">
            <div className="flex h-8 w-8 items-center justify-center rounded bg-[var(--brand-primary)]">
              <span className="text-sm font-black text-white italic">ID</span>
            </div>
            <h1 className="text-lg font-black tracking-tight text-slate-900 dark:text-white">InDel <span className="font-light text-slate-500">Insurer</span></h1>
          </div>

          <div className="flex-1 space-y-8 overflow-y-auto no-scrollbar">
            {renderNavSection(t('sidebar.dashboard'), mainItems)}
            {renderNavSection(t('sidebar.analysis'), analysisItems)}
            {renderNavSection(t('sidebar.claims'), claimsItems)}
          </div>

          <div className="mt-auto px-4 pt-4 border-t border-slate-100 dark:border-slate-800">
             <div className="mb-3">
               <label className="mb-2 block text-[9px] font-black uppercase tracking-[0.3em] text-slate-400">{t('common.language')}</label>
               <select
                value={language}
                onChange={(e) => setLanguage(e.target.value as Language)}
                className="w-full rounded border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-950 px-3 py-2 text-[11px] font-bold text-slate-900 dark:text-white outline-none"
               >
                <option value="en">{t('lang.english')}</option>
                <option value="ta">{t('lang.tamil')}</option>
                <option value="hi">{t('lang.hindi')}</option>
               </select>
             </div>
             <div className="rounded-lg bg-slate-50 dark:bg-slate-800/50 p-4 border border-slate-100 dark:border-slate-800">
                <div className="flex items-center gap-2 mb-1">
                   <div className="h-1.5 w-1.5 rounded-full bg-emerald-500"></div>
                 <p className="text-[9px] font-black uppercase tracking-widest text-emerald-600 dark:text-emerald-400">{t('sidebar.networkConnected')}</p>
                </div>
               <p className="text-[10px] text-slate-500 leading-tight">{t('sidebar.node')}: primary-alpha-12</p>
             </div>
          </div>
        </div>
      </aside>

      {/* Main Surface */}
      <div className="flex-1 pl-64">
        <Navbar />
        <main className="min-h-screen px-12 py-12 max-w-[1400px] mx-auto">
          {children}
        </main>
      </div>
    </div>
  )
}
