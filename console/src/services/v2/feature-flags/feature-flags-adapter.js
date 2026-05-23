import { formatDateToDayMonthYearHour } from '@/helpers/convert-date'

export const FeatureFlagsAdapter = {
  transformList(data) {
    const entries = Object.entries(data?.flags ?? {})
    const body = entries.map(([key, flag]) => ({
      id: key,
      key,
      enabled: flag.enabled,
      description: flag.description,
      updatedAt: formatDateToDayMonthYearHour(flag.updated_at)
    }))
    return { body, count: body.length }
  },

  transformItem(key, flag) {
    return { key, enabled: flag.enabled, description: flag.description }
  }
}
