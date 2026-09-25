export { cn } from 'cn'

export function formatarCapaIGDB(url?: string): string {
  if (!url) return ''
  const comHttps = url.startsWith('//') ? `https:${url}` : url
  return comHttps.replace('/t_thumb/', '/t_cover_big/')
}

export function isoParaDataPt(iso?: string | null): string {
  if (!iso) return ''
  const partes = iso.slice(0, 10).split('-')
  if (partes.length === 3) return `${partes[2]}/${partes[1]}/${partes[0]}`
  return iso
}

export function dataPtParaIso(pt: string): string {
  const limpo = pt.trim()
  const partes = limpo.split('/')
  if (partes.length === 3 && partes[0].length === 2 && partes[1].length === 2 && partes[2].length === 4) {
    return `${partes[2]}-${partes[1]}-${partes[0]}`
  }
  return limpo
}

export function aplicarMascaraData(valor: string): string {
  const digits = valor.replace(/\D/g, '').slice(0, 8)
  if (digits.length > 4) {
    return `${digits.slice(0, 2)}/${digits.slice(2, 4)}/${digits.slice(4)}`
  }
  if (digits.length > 2) {
    return `${digits.slice(0, 2)}/${digits.slice(2)}`
  }
  return digits
}
