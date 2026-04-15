import { useEffect, useState } from 'react'
import { getForecast, type ForecastRow } from '../api/insurer'
import { PageShell, Panel } from './OperationsShared'
import { useLocalization } from '../context/LocalizationContext'

export default function Forecast() {
  const { t } = useLocalization()
  const [rows, setRows] = useState<ForecastRow[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setLoading(true)
    getForecast()
      .then((data) => setRows(Array.isArray(data) ? data : []))
      .catch((err) => setError(err?.message ?? 'Failed to load forecast'))
      .finally(() => setLoading(false))
  }, [])

  return (
    <PageShell
      eyebrow={t('pages.forecast.eyebrow')}
      title={t('pages.forecast.title')}
      description={t('pages.forecast.description')}
    >
      <Panel title={t('pages.forecast.upcomingRisk')}>
        {error ? <p className="text-sm text-rose-600">{error}</p> : null}
        {loading ? <p className="text-sm text-slate-500">Loading forecast...</p> : null}
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {rows.map((row) => (
            <div key={`${row.zone}-${row.date}`} className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
              <p className="text-xs uppercase tracking-[0.22em] text-slate-500">{row.date}</p>
              <p className="mt-2 text-lg font-bold text-slate-950">
                {row.zone}, {row.city}
              </p>
              <p className="mt-3 text-sm text-slate-600">{t('pages.forecast.disruptionProbability')}</p>
              <p className="text-3xl font-black text-slate-950">{Math.round((row.probability ?? 0) * 100)}%</p>
            </div>
          ))}
          {rows.length === 0 && !error && !loading ? (
            <p className="text-sm text-slate-500">{t('pages.forecast.noData')}</p>
          ) : null}
        </div>
      </Panel>
    </PageShell>
  )
}
