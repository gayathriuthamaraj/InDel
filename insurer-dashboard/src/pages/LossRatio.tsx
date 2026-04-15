import { useEffect, useState } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Cell } from 'recharts'
import { getLossRatio, type LossRatioRow } from '../api/insurer'
import { PageShell, Panel } from './OperationsShared'
import { useLocalization } from '../context/LocalizationContext'

export default function LossRatio() {
  const { t } = useLocalization()
  const [rows, setRows] = useState<LossRatioRow[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setLoading(true)
    getLossRatio()
      .then((payload) => setRows(Array.isArray(payload) ? payload : []))
      .catch((err) => setError(err?.message ?? 'Failed to load loss ratio'))
      .finally(() => setLoading(false))
  }, [])

  const chartData = rows
    .map((row) => ({
      name: row.zone_name,
      ratio: Math.round(row.loss_ratio * 100),
      premiums: row.premiums,
      claims: row.claims,
    }))
    .sort((a, b) => b.ratio - a.ratio)
    .slice(0, 8)

  return (
    <PageShell
      eyebrow={t('pages.lossRatio.eyebrow')}
      title={t('pages.lossRatio.title')}
      description={t('pages.lossRatio.description')}
    >
      {error ? (
        <div className="mb-6 rounded border border-rose-200 bg-rose-50 p-4 text-xs font-bold uppercase tracking-widest text-rose-700 dark:border-rose-900 dark:bg-rose-950 dark:text-rose-400">
          {error}
        </div>
      ) : null}

      <div className="grid gap-8 xl:grid-cols-[1fr_0.4fr]">
        <Panel title={t('pages.lossRatio.zoneMetrics')} subtitle={t('pages.lossRatio.zoneMetricsSubtitle')}>
          <div className="h-[350px] w-full">
            {loading ? (
              <div className="flex h-full items-center justify-center text-sm text-slate-500">Loading loss ratio data...</div>
            ) : chartData.length > 0 ? (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={chartData}>
                  <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#e2e8f0" dark-stroke="#1e293b" />
                  <XAxis dataKey="name" axisLine={false} tickLine={false} tick={{ fontSize: 10, fill: '#64748b' }} dy={10} />
                  <YAxis axisLine={false} tickLine={false} tick={{ fontSize: 10, fill: '#64748b' }} />
                  <Tooltip
                    cursor={{ fill: 'rgba(0,0,0,0.05)' }}
                    contentStyle={{ borderRadius: '4px', border: '1px solid #e2e8f0', fontSize: '11px' }}
                  />
                  <Bar dataKey="ratio" radius={[2, 2, 0, 0]} barSize={32}>
                    {chartData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.ratio > 80 ? '#ef4444' : '#f59e0b'} />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            ) : (
              <div className="flex h-full items-center justify-center text-sm text-slate-500">No loss ratio data available.</div>
            )}
          </div>
        </Panel>

        <Panel title={t('pages.lossRatio.summaryInsights')} subtitle={t('pages.lossRatio.summaryInsightsSubtitle')}>
          <div className="space-y-4">
            <div className="rounded border border-orange-200 bg-orange-50 p-4 text-xs text-orange-800 dark:border-orange-950 dark:bg-orange-900/10 dark:text-orange-400">
              <p className="mb-1 text-[9px] font-bold uppercase tracking-widest">{t('pages.lossRatio.exposureAlert')}</p>
              <p className="leading-relaxed">{t('pages.lossRatio.exposureAlertDesc')}</p>
            </div>
            <div className="rounded border border-emerald-200 bg-emerald-50 p-4 text-xs text-emerald-800 dark:border-emerald-950 dark:bg-emerald-900/10 dark:text-emerald-400">
              <p className="mb-1 text-[9px] font-bold uppercase tracking-widest">{t('pages.lossRatio.growthOpportunity')}</p>
              <p className="leading-relaxed">{t('pages.lossRatio.growthOpportunityDesc')}</p>
            </div>
          </div>
        </Panel>
      </div>

      <Panel title={t('pages.lossRatio.dataGrid')} className="mt-8">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead>
              <tr className="border-b border-slate-200 dark:border-slate-800">
                <th className="pb-4 font-black uppercase tracking-widest text-slate-400">{t('pages.lossRatio.headerZone')}</th>
                <th className="pb-4 font-black uppercase tracking-widest text-slate-400">{t('pages.lossRatio.headerPremiums')}</th>
                <th className="pb-4 font-black uppercase tracking-widest text-slate-400">{t('pages.lossRatio.headerClaims')}</th>
                <th className="pb-4 font-black uppercase tracking-widest text-slate-400 text-right">{t('pages.lossRatio.headerRatio')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 dark:divide-slate-800/50">
              {rows.map((row) => (
                <tr key={`${row.city}-${row.zone_name}`} className="hover:bg-slate-50 dark:hover:bg-slate-800 transition-none">
                  <td className="py-4">
                    <p className="font-bold text-slate-900 dark:text-white">{row.zone_name}</p>
                    <p className="text-[9px] font-black uppercase tracking-widest text-slate-400">{row.city}</p>
                  </td>
                  <td className="py-4 text-slate-500">Rs {Math.round(row.premiums).toLocaleString()}</td>
                  <td className="py-4 text-slate-500">Rs {Math.round(row.claims).toLocaleString()}</td>
                  <td className="py-4 text-right">
                    <span
                      className={`rounded px-2 py-0.5 text-[10px] font-black uppercase tracking-widest ${
                        row.loss_ratio > 0.8 ? 'bg-rose-500/10 text-rose-600' : 'bg-orange-500/10 text-orange-600'
                      }`}
                    >
                      {Math.round(row.loss_ratio * 100)}%
                    </span>
                  </td>
                </tr>
              ))}
              {rows.length === 0 && !error ? (
                <tr>
                  <td className="py-12 text-center italic text-slate-400" colSpan={4}>
                    No zone metrics currently streaming.
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
