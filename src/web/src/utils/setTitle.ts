// A set's name as its card shows it: a joint issue's ("Emissão comum de
// 2007: 50th anniversary of the Treaty of Rome") says its year apart, so its
// name is what comes after it. Pass the year of a joint issue only: an
// annual series' name ("Portugal 2010: série anual") is kept whole.
export function setDisplayTitle(name: string, year?: number | null): string {
  if (!year) return name
  const match = new RegExp(`^.*\\b${year}\\s*[:–—-]\\s*(.+)$`).exec(name.trim())
  return match?.[1]?.trim() || name
}
