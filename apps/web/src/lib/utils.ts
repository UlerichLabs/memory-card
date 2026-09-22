export { cn } from 'cn'

export function formatarCapaIGDB(url?: string): string {
  if (!url) return ''
  const comHttps = url.startsWith('//') ? `https:${url}` : url
  return comHttps.replace('/t_thumb/', '/t_cover_big/')
}
