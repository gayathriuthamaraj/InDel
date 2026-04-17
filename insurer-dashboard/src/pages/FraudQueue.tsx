import { useEffect, useState } from 'react'
import { getFraudQueue, reviewClaim, type FraudQueueRow } from '../api/insurer'
import { PageShell, Panel } from './OperationsShared'
import { useLocalization } from '../context/LocalizationContext'

export default function FraudQueue() {
  const { t } = useLocalization()
  const [rows, setRows] = useState<FraudQueueRow[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const loadQueue = () => {
    setLoading(true)
    getFraudQueue({ page: 1, limit: 20 })
      .then((payload) => setRows(payload.data ?? []))
      .catch((err) => setError(err?.message ?? 'Failed to load fraud queue'))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    loadQueue()
  }, [])

  const handleAction = async (id: number, status: string, verdict: string) => {
    try {
      await reviewClaim(id, status, verdict)
      loadQueue()
    } catch (e: any) {
      setError(e?.message ?? 'Failed to perform action')
    }
  }

  return (
    <PageShell
      eyebrow={t('pages.fraudQueue.eyebrow')}
      title={t('pages.fraudQueue.title')}
      description={t('pages.fraudQueue.description')}
    >
      <div className="grid gap-8">
        <Panel title={t('pages.fraudQueue.securityFlags')} subtitle={t('pages.fraudQueue.securityFlagsSubtitle')}>
          {error ? <p className="mb-6 text-[10px] font-black uppercase text-rose-600">{error}</p> : null}
          {loading ? <p className="mb-6 text-sm text-slate-500">Loading fraud queue...</p> : null}
          <div className="overflow-x-auto">
            <table className="w-full text-left text-[11px]">
              <thead>
                <tr className="border-b border-slate-200 dark:border-slate-800">
                  <th className="pb-4 font-black uppercase tracking-widest text-slate-400">{t('pages.fraudQueue.headerID')}</th>
                  <th className="pb-4 font-black uppercase tracking-widest text-slate-400">{t('pages.fraudQueue.headerSignal')}</th>
                  <th className="pb-4 font-black uppercase tracking-widest text-slate-400 text-right">{t('pages.fraudQueue.headerAnomalyScore')}</th>
                  <th className="pb-4 text-right font-black uppercase tracking-widest text-slate-400">{t('pages.fraudQueue.headerVerification')}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 dark:divide-slate-800/50">
                {rows.map((row) => (
                  <tr key={row.claim_id} className="hover:bg-slate-50 dark:hover:bg-slate-800 transition-none">
                    <td className="py-5 font-bold text-slate-900 dark:text-white">#{row.claim_id}</td>
                    <td className="py-5">
                      <div className="flex flex-wrap gap-2">
                        {row.violations && row.violations.length > 0 ? (
                          row.violations.map((v, i) => (
                            <span key={i} className="rounded bg-slate-100 px-2 py-0.5 text-[9px] font-black uppercase tracking-widest text-slate-500 dark:bg-slate-800">
                              {v}
                            </span>
                          ))
                        ) : (
                          <span className="rounded bg-slate-100 px-2 py-0.5 text-[9px] font-black uppercase tracking-widest text-slate-500 dark:bg-slate-800">
                            {row.status}
                          </span>
                        )}
                      </div>
                    </td>
                    <td className="py-5 text-right font-black text-slate-900 dark:text-white">
                      <div className="flex items-center justify-end gap-3">
                        <span className="text-sm font-black">{(row.fraud_score ?? 0).toFixed(2)}</span>
                        <div className="h-1 w-16 overflow-hidden bg-slate-100 dark:bg-slate-800">
                          <div
                            className={`h-full ${row.fraud_score > 0.7 ? 'bg-rose-500' : 'bg-orange-500'}`}
                            style={{ width: `${Math.min((row.fraud_score ?? 0) * 100, 100)}%` }}
                          />
                        </div>
                      </div>
                    </td>
                    <td className="py-5 text-right text-[9px] font-black uppercase tracking-[0.2em]">
                      {row.fraud_verdict === 'delay' || row.status === 'manual_review' ? (
                        <div className="flex justify-end gap-2">
                          <button
                            onClick={() => handleAction(row.claim_id, 'approved', 'safe')}
                            className="rounded bg-emerald-500/10 px-3 py-1 text-emerald-500 hover:bg-emerald-500 hover:text-white transition-colors"
                          >
                            Approve
                          </button>
                          <button
                            onClick={() => handleAction(row.claim_id, 'rejected', 'fraud')}
                            className="rounded bg-rose-500/10 px-3 py-1 text-rose-500 hover:bg-rose-500 hover:text-white transition-colors"
                          >
                            Reject
                          </button>
                        </div>
                      ) : (
                        <span className={row.fraud_verdict === 'safe' || row.fraud_verdict === 'approve' ? 'text-emerald-600' : 'text-rose-600'}>{row.fraud_verdict}</span>
                      )}
                    </td>
                  </tr>
                ))}
                {rows.length === 0 && !error && !loading ? (
                  <tr>
                    <td className="py-12 text-center italic text-slate-400" colSpan={4}>
                      {t('pages.fraudQueue.noData')}
                    </td>
                  </tr>
                ) : null}
              </tbody>
            </table>
          </div>
        </Panel>
      </div>
    </PageShell>
  )
}
