export function apiUrl(path: string): string {
  const base = (import.meta.env.VITE_API_BASE_URL || '').trim();
  if (!base) {
    return path;
  }
  return new URL(path, base).toString();
}
