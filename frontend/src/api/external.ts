import { asArray, jsonInit, request, type ApiRecord, type ApiRequestInit } from './client'
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
  // Unavailable endpoints/capabilities are failures, not successful empty lists.
  // Configuration absence is represented explicitly by the backend row.
  const payload = await request<ListEnvelope<ApiRecord> | ApiRecord>(`${path}${suffix}`, jsonInit('GET', {
    cluster_id: clusterId,
    ...body
  }))
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

export async function queryMetric(clusterId: number, input: MetricQueryInput, init?: ApiRequestInit) {
  const query = new URLSearchParams({ metric_id: input.metricId })
  if (input.time) {
    query.set('time', input.time)
  }
  return readMetric(`/metric/query?${query}`, 'vector', clusterId, init)
}

export async function queryMetricRange(clusterId: number, input: MetricRangeInput, init?: ApiRequestInit) {
  const query = new URLSearchParams({
    metric_id: input.metricId,
    start: input.start,
    end: input.end,
    step: input.step
  })
  return readMetric(`/metric/range?${query}`, 'matrix', clusterId, init)
}

async function readMetric(path: string, resultType: 'vector' | 'matrix', clusterId?: number, init?: ApiRequestInit) {
  if (!clusterId) {
    throw new Error('请先选择集群')
  }
  const payload = await request<MetricResponse>(path, jsonInit('GET', { cluster_id: clusterId }, init))
  if (!payload || typeof payload !== 'object' || Array.isArray(payload) || payload.result_type !== resultType || !Array.isArray(payload.series) || payload.series.some(item => !item || typeof item !== 'object' || Array.isArray(item))) {
    throw new Error('指标响应格式异常，不能据此认定没有指标数据')
  }
  return payload
}
