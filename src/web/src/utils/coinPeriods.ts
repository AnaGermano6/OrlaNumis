// A coin's period, as uCoin divides a country's coins ("República Portuguesa
// 1986–2001"). The server works it out (services.CoinPeriodService):
// Portugal's from the year the coin's type began, other countries' from the
// uCoin import.

const PERIOD_YEARS = /(-?\d{1,4})\s*[–-]\s*(-?\d{1,4})\s*$/

/** The years a period's name gives ("República Portuguesa 1986–2001" → 1986, 2001). */
export function periodYears(period: string): { from: number; to: number } | undefined {
  const match = PERIOD_YEARS.exec(period.trim())
  return match ? { from: Number(match[1]), to: Number(match[2]) } : undefined
}
