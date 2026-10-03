import type { ApiRecord } from '../../api/client'
import { rgwS3KeyCreateInput } from './rgwS3KeyCreate'
import { rgwS3KeyDeleteInput } from './rgwS3KeyDelete'

export function rgwS3KeyRotateInput(values: Record<string, unknown>, row?: ApiRecord) {
  rgwS3KeyDeleteInput(values, row)
  return rgwS3KeyCreateInput(values, row)
}
