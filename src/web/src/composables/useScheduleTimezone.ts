import { t } from '@/i18n'

/** Returns the browser's IANA timezone (e.g. "America/Chicago"). */
export function getBrowserTimezone(): string {
  return Intl.DateTimeFormat().resolvedOptions().timeZone
}

// Intl.supportedValuesOf is missing on older Safari; fall back to the browser
// zone alone so a saved value is still selectable.
export function listTimezones(): string[] {
  const intl = Intl as typeof Intl & { supportedValuesOf?: (key: 'timeZone') => string[] }
  const zones = intl.supportedValuesOf?.('timeZone') ?? []
  return zones.length ? zones : [getBrowserTimezone()].filter(Boolean)
}

export function timezoneOptionsWith(saved?: string): string[] {
  const zones = listTimezones()
  return saved && !zones.includes(saved) ? [saved, ...zones] : zones
}

export function scheduleZoneLabel(zone?: string): string {
  return zone?.trim() || t('admin.schedules.serverTime')
}

function formatMinutes(minutes: number): string {
  if (minutes % 60 === 0) {
    return t('admin.schedules.everyHours', { count: minutes / 60 })
  }
  return t('admin.schedules.everyMinutes', { count: minutes })
}

export interface ScheduleDescription {
  startTime?: string
  defaultStartTime: string
  zone?: string
  intervalMinutes?: string | number
  intervalDays?: string | number
}

// Mirrors the API's dailySchedule rules: sub-day intervals restart from the
// start time each day; day-or-longer intervals run at the start time.
export function describeSchedule(options: ScheduleDescription): string {
  const time = options.startTime?.trim() || options.defaultStartTime
  const zone = scheduleZoneLabel(options.zone)

  let days = Number(options.intervalDays)
  const minutes = Number(options.intervalMinutes)
  if (!(days > 0) && minutes > 0) {
    if (minutes < 1440) {
      return t('admin.schedules.runsInterval', { time, zone, interval: formatMinutes(Math.floor(minutes)) })
    }
    days = Math.floor(minutes / 1440)
  }
  days = days > 0 ? Math.floor(days) : 1
  return days === 1
    ? t('admin.schedules.runsDaily', { time, zone })
    : t('admin.schedules.runsEveryDays', { days, time, zone })
}
