import { useEffect, useState } from 'react'
import { getClaims, type ClaimListRow } from '../api/insurer'
import { PageShell, Panel } from './OperationsShared'
import { useLocalization } from '../context/LocalizationContext'

export default function Claims() {
  const { t } = useLocalization()
  const [rows, setRows] = useState<ClaimListRow[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setLoading(true)
    getClaims({ page: 1, limit: 20 })
      .then((payload) => setRows(payload.data ?? []))
      .catch((err) => setError(err?.message ?? 'Failed to load claims'))
      .finally(() => setLoading(false))
  }, [])

  return (
    <PageShell
      eyebrow={t('pages.claims.eyebrow')}
      title={t('pages.claims.title')}
      description={t('pages.claims.description')}
    >
      <Panel title={t('pages.claims.activeStream')} subtitle={t('pages.claims.activeStreamSubtitle')}>
        {error ? <p className="mb-6 text-[10px] font-black uppercase text-rose-600">{error}</p> : null}
        {loading ? <p className="mb-6 text-sm text-slate-500">Loading claims...</p> : null}
        <div className="overflow-x-auto">
          <table className="w-full text-left text-[11px]">
            <thead>
              <tr className="border-b border-slate-200 dark:border-slate-800">
                <th className="pb-4 font-black uppercase tracking-widest text-slate-400">{t('pages.claims.headerID')}</th>
                <th className="pb-4 font-black uppercase tracking-widest text-slate-400">{t('pages.claims.headerWorker')}</th>
                <th className="pb-4 font-black uppercase tracking-widest text-slate-400">{t('pages.claims.headerZone')}</th>
                <th className="pb-4 font-black uppercase tracking-widest text-slate-400">{t('pages.claims.headerValuation')}</th>
                <th className="pb-4 font-black uppercase tracking-widest text-slate-400">{t('pages.claims.headerStatus')}</th>
                <th className="pb-4 text-right font-black uppercase tracking-widest text-slate-400">{t('pages.claims.headerSecurity')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 dark:divide-slate-800/50">
              {rows.map((row) => (
                <tr key={row.claim_id} className="hover:bg-slate-50 dark:hover:bg-slate-800 transition-none">
                  <td className="py-4 font-bold text-slate-900 dark:text-white">#{row.claim_id}</td>
                  <td className="py-4 text-slate-500">{row.worker_id}</td>
                  <td className="py-4 font-medium text-slate-700 dark:text-slate-300">{row.zone_name}</td>
                  <td className="py-4 font-bold text-slate-900 dark:text-white">Rs {Math.round(row.claim_amount ?? 0).toLocaleString()}</td>
                  <td className="py-4">
                    <span
                      className={`rounded px-2 py-0.5 text-[9px] font-black uppercase tracking-widest ${
                        row.status === 'approved' ? 'bg-emerald-100 text-emerald-700' : 'bg-amber-100 text-amber-700'
                      }`}
                    >
                      {row.status}
                    </span>
                  </td>
                  <td className="py-4 text-right">
                    <span
                      className={`rounded px-2 py-0.5 text-[9px] font-black uppercase tracking-widest leading-none ${
                        row.fraud_verdict === 'safe' ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'
                      }`}
                    >
                      {row.fraud_verdict === 'safe' ? t('pages.claims.safe') : t('pages.claims.flagged')}
                    </span>
                  </td>
                </tr>
              ))}
              {rows.length === 0 && !error && !loading ? (
                <tr>
                  <td className="py-12 text-center italic text-slate-400" colSpan={6}>
                    {t('pages.claims.noData')}
                  </td>
                </tr>
              ) : null}
            </tbody>
          </table>
        </div>
      </Panel>
    </PageShell>
  )
}
