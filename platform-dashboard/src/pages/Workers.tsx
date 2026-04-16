import { useEffect, useMemo, useState } from 'react'
import { getWorkers } from '../api/platform'
import { Search, MoreVertical, CreditCard, MapPin, Download } from 'lucide-react'
import { useLocalization } from '../context/LocalizationContext'

function formatWorkerZone(zone: string, unknownZoneLabel: string) {
  const rawZone = String(zone || '').trim()
  if (!rawZone) return unknownZoneLabel

  const parts = rawZone.split(',').map((part) => part.trim()).filter(Boolean)
  if (parts.length < 2) return rawZone

  const [primaryZone, ...secondaryParts] = parts
  const secondaryZone = secondaryParts.join(', ')
  const normalize = (value: string) =>
    value.toLowerCase().replace(/[()\-\u2014]/g, ' ').replace(/\s+/g, ' ').trim()

  const normalizedPrimary = normalize(primaryZone)
  const normalizedSecondary = normalize(secondaryZone)
  const primaryLooksComplete = /[()\-]/.test(primaryZone) || primaryZone.includes('\u2014')

  if (!normalizedSecondary) return primaryZone
  if (normalizedPrimary === normalizedSecondary) return primaryZone
  if (normalizedPrimary.includes(normalizedSecondary)) return primaryZone
  if (primaryLooksComplete) return primaryZone

  return rawZone
}

export default function Workers() {
  const { t } = useLocalization()
  const [workers, setWorkers] = useState<any[]>([])
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<'all' | 'live' | 'offline'>('all')
  const unknownZoneLabel = t('pages.workers.unknownZone')

  useEffect(() => {
    async function load() {
      try {
        const response = await getWorkers()
        setWorkers(response.data.workers ?? [])
      } catch (error) {
        console.error('Failed to load workers', error)
      }
    }

    load()
    const interval = setInterval(load, 10000) // Poll every 10 seconds
    return () => clearInterval(interval)
  }, [])

  const filteredWorkers = useMemo(() => {
    return workers.filter((worker) => {
      const zoneLabel = formatWorkerZone(worker.zone, unknownZoneLabel)
      const matchesSearch =
        worker.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        worker.worker_id.toString().includes(searchQuery) ||
        zoneLabel.toLowerCase().includes(searchQuery.toLowerCase())

      const isLive = !!worker.is_online
      const matchesStatus =
        statusFilter === 'all' ||
        (statusFilter === 'live' && isLive) ||
        (statusFilter === 'offline' && !isLive)

      return matchesSearch && matchesStatus
    })
  }, [workers, searchQuery, statusFilter])

  const exportToCSV = () => {
    const headers = [
      t('pages.workers.headerWorkerID'),
      t('pages.workers.headerName'),
      t('pages.workers.headerPhone'),
      t('pages.workers.headerZone'),
      t('pages.workers.headerStatus'),
    ]
    const rows = filteredWorkers.map((worker) => {
      const zoneLabel = formatWorkerZone(worker.zone, unknownZoneLabel)
      return [
        worker.worker_id,
        worker.name,
        worker.phone,
        zoneLabel,
        zoneLabel !== unknownZoneLabel ? t('pages.workers.live') : t('pages.workers.offline'),
      ]
    })

    const csvContent = [headers, ...rows].map((entry) => entry.join(',')).join('\n')
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const link = document.createElement('a')
    const url = URL.createObjectURL(blob)
    link.setAttribute('href', url)
    link.setAttribute('download', `${t('pages.workers.exportFileName')}${new Date().toISOString().split('T')[0]}.csv`)
    link.style.visibility = 'hidden'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }

  return (
    <div className="space-y-10">
      <div className="flex items-end justify-between">
        <div>
          <h1 className="text-2xl font-black tracking-tight text-slate-900 dark:text-white">{t('pages.workers.title')}</h1>
          <p className="mt-1 text-sm text-slate-500">
            {t('pages.workers.description')} • {t('pages.workers.lastUpdated')}: {new Date().toLocaleTimeString()}
          </p>
        </div>
        <div className="flex gap-3">
          <button
            onClick={exportToCSV}
            className="flex items-center gap-2 h-9 px-4 rounded-md border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-xs font-bold font-['Outfit'] hover:bg-slate-50 dark:hover:bg-slate-800 transition-none"
          >
            <Download className="h-3.5 w-3.5" />
            {t('pages.workers.exportCSV')}
          </button>
          <button
            onClick={() => window.location.reload()}
            className="flex items-center gap-2 h-9 px-4 rounded-md border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 text-xs font-bold font-['Outfit'] hover:bg-slate-50 dark:hover:bg-slate-800 transition-none"
          >
            <Search className="h-3.5 w-3.5" />
            {t('pages.workers.refresh')}
          </button>
        </div>
      </div>

      <div className="enterprise-panel overflow-hidden">
        <div className="border-b border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/20 p-4 flex items-center justify-between">
          <div className="relative group w-72">
            <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3 text-slate-400">
              <Search className="h-3.5 w-3.5" />
            </div>
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder={t('pages.workers.searchPlaceholder')}
              className="w-full rounded border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 py-1.5 pl-9 pr-3 text-[11px] text-slate-900 dark:text-white outline-none focus:border-orange-500 transition-none"
            />
          </div>
          <div className="flex items-center gap-2">
            <span className="text-[10px] font-black uppercase text-slate-400 mr-2">{t('pages.workers.filterStatus')}:</span>
            {(['all', 'live', 'offline'] as const).map((status) => (
              <button
                key={status}
                onClick={() => setStatusFilter(status)}
                className={`px-3 py-1.5 rounded border text-[10px] font-bold transition-none uppercase ${
                  statusFilter === status
                    ? 'bg-orange-600 border-orange-600 text-white shadow-sm'
                    : 'border-slate-200 dark:border-slate-700 text-slate-500 hover:text-slate-900 dark:hover:text-white'
                }`}
              >
                {status === 'all' ? t('pages.workers.allStatus') : status === 'live' ? t('pages.workers.live') : t('pages.workers.offline')}
              </button>
            ))}
          </div>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full border-collapse text-left">
            <thead>
              <tr className="border-b border-slate-100 dark:border-slate-800">
                <th className="px-8 py-4 text-[10px] font-black uppercase tracking-widest text-slate-400">{t('pages.workers.headerWorker')}</th>
                <th className="px-8 py-4 text-[10px] font-black uppercase tracking-widest text-slate-400">{t('pages.workers.headerZoneAssignment')}</th>
                <th className="px-8 py-4 text-[10px] font-black uppercase tracking-widest text-slate-400">{t('pages.workers.headerPolicyStatus')}</th>
                <th className="px-8 py-4 text-[10px] font-black uppercase tracking-widest text-slate-400">{t('pages.workers.headerActivity')}</th>
                <th className="px-8 py-4 text-[10px] font-black uppercase tracking-widest text-slate-400">{t('pages.workers.headerActions')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
              {filteredWorkers.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-8 py-20 text-center text-slate-400 text-xs italic">
                    {t('pages.workers.noData')}
                  </td>
                </tr>
              ) : (
                filteredWorkers.map((worker) => {
                  const zoneLabel = formatWorkerZone(worker.zone, unknownZoneLabel)
                  const isLive = zoneLabel !== unknownZoneLabel

                  return (
                    <tr key={worker.worker_id} className="hover:bg-slate-50 dark:hover:bg-slate-800/30">
                      <td className="px-8 py-5">
                        <div className="flex items-center gap-3">
                          <div className="h-9 w-9 rounded bg-slate-100 dark:bg-slate-800 flex items-center justify-center text-[10px] font-black text-slate-500 uppercase">
                            {worker.name.split(' ').map((namePart: string) => namePart[0]).join('')}
                          </div>
                          <div>
                            <div className="text-xs font-bold text-slate-900 dark:text-white">{worker.name}</div>
                            <div className="text-[10px] text-slate-400 font-medium">#{String(worker.worker_id).padStart(6, '0')}</div>
                          </div>
                        </div>
                      </td>
                      <td className="px-8 py-5">
                        <div className="flex items-center gap-1.5">
                          <MapPin className="h-3 w-3 text-slate-400" />
                          <span className="text-xs font-medium text-slate-600 dark:text-slate-300">{zoneLabel}</span>
                        </div>
                      </td>
                      <td className="px-8 py-5">
                        <div
                          className={`flex items-center gap-2 px-2 py-1 rounded w-fit border ${
                            worker.is_online
                              ? 'bg-emerald-50 dark:bg-emerald-500/10 border-emerald-100 dark:border-emerald-500/20'
                              : 'bg-slate-100 dark:bg-slate-800 border-slate-200 dark:border-slate-700 opacity-60'
                          }`}
                        >
                          <div className={`h-1 w-1 rounded-full ${worker.is_online ? 'bg-emerald-500' : 'bg-slate-400'}`}></div>
                          <span
                            className={`text-[9px] font-black uppercase tracking-tighter ${
                              worker.is_online ? 'text-emerald-600 dark:text-emerald-400' : 'text-slate-500'
                            }`}
                          >
                            {worker.is_online ? t('pages.workers.activeCoverage') : t('pages.workers.inactive')}
                          </span>
                        </div>
                      </td>
                      <td className="px-8 py-5">
                        <div className={`text-xs font-bold ${worker.is_online ? 'text-emerald-600 dark:text-emerald-400' : 'text-slate-500'}`}>
                          {worker.is_online ? t('pages.workers.liveOnShift') : t('pages.workers.offline')}
                        </div>
                        <div className="text-[9px] text-slate-400 mt-0.5">
                          {worker.is_online ? (
                             <>{t('pages.workers.contact')}: {worker.phone}</>
                          ) : (
                             <>
                               {worker.last_active_at && worker.last_active_at !== '0001-01-01T00:00:00Z' 
                                 ? `${t('pages.workers.lastSeen')} ${new Date(worker.last_active_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}` 
                                 : t('pages.workers.neverSeen')}
                             </>
                          )}
                        </div>
                      </td>
                      <td className="px-8 py-5">
                        <div className="flex items-center gap-3">
                          <button className="p-1 text-slate-400 hover:text-orange-500 transition-colors">
                            <CreditCard className="h-4 w-4" />
                          </button>
                          <button className="p-1 text-slate-400 hover:text-slate-900 dark:hover:text-white transition-colors">
                            <MoreVertical className="h-4 w-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  )
                })
              )}
            </tbody>
          </table>
        </div>

        <div className="p-6 border-t border-slate-100 dark:border-slate-800 bg-slate-50/30 dark:bg-slate-800/10 flex items-center justify-between">
          <div className="text-[10px] font-bold text-slate-400 uppercase tracking-widest">
            {t('pages.workers.showingNodes').replace('{filtered}', String(filteredWorkers.length)).replace('{total}', String(workers.length))}
          </div>
          <div className="flex gap-2">
            <button
              disabled
              className="px-3 py-1 rounded border border-slate-200 dark:border-slate-800 text-[10px] font-bold text-slate-300 dark:text-slate-700 cursor-not-allowed"
            >
              {t('pages.workers.prev')}
            </button>
            <button
              disabled
              className="px-3 py-1 rounded border border-slate-200 dark:border-slate-800 text-[10px] font-bold text-slate-300 dark:text-slate-700 cursor-not-allowed"
            >
              {t('pages.workers.next')}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
