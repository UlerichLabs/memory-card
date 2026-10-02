import { useEffect } from 'react'

export function useDocumentTitle(titulo?: string) {
  useEffect(() => {
    document.title = titulo ? `${titulo} · Memory Card` : 'Memory Card'
    return () => { document.title = 'Memory Card' }
  }, [titulo])
}
