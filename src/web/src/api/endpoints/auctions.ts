// calendar endpoints (auction events). Split out of the former monolithic client.ts;
// re-exported from '@/api/client' so existing imports keep working.
import { api } from '@/api/http'
import type { AuctionLot, CalendarEventDetail } from '@/types'

// Calendar / Auction Events
export const getCalendar = (start?: string, end?: string) => {
  const params: Record<string, string> = {}
  if (start) params.start = start
  if (end) params.end = end
  return api.get('/calendar', { params })
}

export const getCalendarEvent = (id: number) => api.get<{ event: CalendarEventDetail; lots: AuctionLot[] }>(`/calendar/events/${id}`)

export const createCalendarEvent = (data: { title: string; auctionHouse?: string; startDate?: string; endDate?: string; url?: string; notes?: string }) => api.post('/calendar/events', data)

export const updateCalendarEvent = (id: number, data: Record<string, unknown>) => api.put(`/calendar/events/${id}`, data)

export const deleteCalendarEvent = (id: number) => api.delete(`/calendar/events/${id}`)
