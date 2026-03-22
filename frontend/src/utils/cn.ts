type ClassValue = string | undefined | null | false | Record<string, boolean>

/**
 * Merges class names, filtering out falsy values.
 * Lightweight alternative to clsx for basic use cases.
 */
export function cn(...classes: ClassValue[]): string {
  return classes
    .flatMap((cls) => {
      if (!cls) return []
      if (typeof cls === 'string') return [cls]
      return Object.entries(cls)
        .filter(([, v]) => v)
        .map(([k]) => k)
    })
    .join(' ')
}
