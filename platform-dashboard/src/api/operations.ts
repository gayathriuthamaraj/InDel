import { coreClient } from './client'

export const getPayoutReconciliation = (params: { from: string; to: string }) =>
  coreClient.get('/api/v1/internal/payouts/reconciliation', { params })
