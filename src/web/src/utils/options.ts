/**
 * Parse a newline-delimited list of options into an array.
 * - Trim whitespace from each line
 * - Drop blank lines
 * - De-duplicate while preserving order
 * - Fallback to defaults if input is empty/invalid
 */
export function parseOptionList(value: string | null | undefined, defaults: string[]): string[] {
  if (!value || value.trim() === '') {
    return [...defaults]
  }

  const lines = value.split('\n')
    .map(line => line.trim())
    .filter(line => line.length > 0)

  // De-duplicate using a Set, preserving first occurrence order
  const seen = new Set<string>()
  const unique: string[] = []
  for (const line of lines) {
    if (!seen.has(line)) {
      seen.add(line)
      unique.push(line)
    }
  }

  return unique.length > 0 ? unique : [...defaults]
}

/**
 * Convert an array of options back to a newline-delimited string.
 */
export function formatOptionList(options: string[]): string {
  return options.join('\n')
}

/**
 * Parse a newline-delimited "Value=Label" list (the shape of the
 * CoinCategoryLabelsPtPT / CoinEraLabelsPtPT settings) into a map keyed by
 * the stored value. Mirrors the API's ParseSettingLabelMap: only the first
 * "=" separates value from label, blank or malformed lines are skipped, and a
 * later line for the same value wins.
 */
export function parseOptionLabelMap(value: string | null | undefined): Record<string, string> {
  const labels: Record<string, string> = {}
  for (const line of (value ?? '').split('\n')) {
    const separator = line.indexOf('=')
    if (separator < 0) continue
    const key = line.slice(0, separator).trim()
    const label = line.slice(separator + 1).trim()
    if (key && label) labels[key] = label
  }
  return labels
}

/**
 * Convert per-option labels back to "Value=Label" lines, keeping only options
 * that are still configured and have a non-blank label.
 */
export function formatOptionLabelMap(options: string[], labels: Record<string, string>): string {
  return options
    .filter(option => (labels[option] ?? '').trim() !== '')
    .map(option => `${option}=${(labels[option] ?? '').trim()}`)
    .join('\n')
}
