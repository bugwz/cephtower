import { asArray, isApiError, jsonInit, request, type ApiRecord } from './client'
import type { ListEnvelope } from './types'

export interface MetricQueryInput {
  metricId: string
  time?: string
}

export interface MetricRangeInput {
  metricId: string
  start: string
  end: string
  step: string
}

export interface MetricResponse {
  result_type: string
  series: ApiRecord[]
  meta?: ApiRecord
}

export async function readExternalList(path: string, clusterId: number, body: ApiRecord = {}, query?: URLSearchParams) {
  const suffix = query?.toString() ? `?${query.toString()}` : ''
  let payload: ListEnvelope<ApiRecord> | ApiRecord
  try {
    payload = await request<ListEnvelope<ApiRecord> | ApiRecord>(`${path}${suffix}`, jsonInit('GET', {
      cluster_id: clusterId,
      ...body
    }))
  } catch (err) {
    if (isApiError(err, 501, 'capability_unavailable') || (isApiError(err, 501) && !err.code)) {
      return { items: [], meta: undefined }
    }
    throw err
  }
  if ('items' in payload && Array.isArray(payload.items)) {
    return {
      items: payload.items,
      meta: payload.meta
    }
  }
  return {
    items: asArray(payload),
    meta: undefined
  }
}

export async function queryMetric(clusterId: number, input: MetricQueryInput) {
  const query = new URLSearchParams({ metric_id: input.metricId })
  if (input.time) {
    query.set('time', input.time)
  }
  return readMetric(`/metric/query?${query}`, clusterId)
}

export async function queryMetricRange(clusterId: number, input: MetricRangeInput) {
  const query = new URLSearchParams({
    metric_id: input.metricId,
    start: input.start,
    end: input.end,
    step: input.step
  })
  return readMetric(`/metric/range?${query}`, clusterId)
}

async function readMetric(path: string, clusterId?: number) {
  if (!clusterId) {
    throw new Error('请先选择集群')
  }
  const payload = await request<MetricResponse>(path, jsonInit('GET', { cluster_id: clusterId }))
  return {
    ...payload,
    series: Array.isArray(payload.series) ? payload.series.filter((item): item is ApiRecord => typeof item === 'object' && item !== null && !Array.isArray(item)) : []
  }
}
