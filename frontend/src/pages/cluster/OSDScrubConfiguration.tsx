import { Alert } from 'antd'
import { ConfigurationPage } from './ConfigurationPage'

// Reference: OsdPgScrubModalOptions basicOptions and advancedOptions.
export const osdScrubOptions = [
  'osd_scrub_during_recovery', 'osd_scrub_begin_hour', 'osd_scrub_end_hour',
  'osd_scrub_begin_week_day', 'osd_scrub_end_week_day', 'osd_scrub_min_interval',
  'osd_scrub_max_interval', 'osd_deep_scrub_interval', 'osd_scrub_auto_repair',
  'osd_max_scrubs', 'osd_scrub_priority', 'osd_scrub_sleep',
  'osd_scrub_auto_repair_num_errors', 'osd_debug_deep_scrub_sleep', 'osd_deep_scrub_keys',
  'osd_deep_scrub_large_omap_object_key_threshold', 'osd_deep_scrub_large_omap_object_value_sum_threshold',
  'osd_deep_scrub_randomize_ratio', 'osd_deep_scrub_stride', 'osd_deep_scrub_update_digest_min_age',
  'osd_requested_scrub_priority', 'osd_scrub_backoff_ratio', 'osd_scrub_chunk_max',
  'osd_scrub_chunk_min', 'osd_scrub_cost', 'osd_scrub_interval_randomize_ratio',
  'osd_scrub_invalid_stats', 'osd_scrub_load_threshold', 'osd_scrub_max_preemptions',
  'osd_shallow_scrub_chunk_max', 'osd_shallow_scrub_chunk_min'
] as const

export function OSDScrubConfiguration() {
  return <>
    <Alert type="info" message="PG Scrub 基础与高级参数" description="此处管理集群配置，不是立即执行 Scrub。每次设置或移除覆盖前，请确认表单中的作用域；global 会影响所有适用守护进程。仅展示当前集群已采集的参数，缺失参数不代表默认值。" />
    <ConfigurationPage optionNames={osdScrubOptions} />
  </>
}
