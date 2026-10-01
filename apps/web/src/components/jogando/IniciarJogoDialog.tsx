import { useEffect, useState } from 'react'
import { Gamepad2 } from 'lucide-react'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { DateInput } from '@/components/jogos/GameForm/DateInput'
import { GameFormAutocomplete } from '@/components/jogos/GameForm/GameFormAutocomplete'
import { formatarCapaIGDB } from '@/lib/utils'
import { hojeIso } from '@/lib/jogandoUtils'
import { useJogandoStore } from '@/stores/jogandoStore'
import type { IGDBJogoSugestao } from '@/lib/services/jogosService'
import { useIniciarJogoForm } from './useIniciarJogoForm'
import type { IniciarJogoFormData } from './iniciarJogo.schema'

export function IniciarJogoDialog() {
  const { isModalOpen, fecharModal, criar } = useJogandoStore()
  const [modoData, setModoData] = useState<'hoje' | 'ontem' | 'outra'>('hoje')
  const form = useIniciarJogoForm(async (data: IniciarJogoFormData) => { await criar(data); fecharModal() })
  const { reset } = form
  const ontem = (() => { const data = new Date(); data.setDate(data.getDate() - 1); return hojeIso(data) })()

  useEffect(() => {
    if (isModalOpen) {
      reset()
      setModoData('hoje')
    }
  }, [isModalOpen, reset])

  function selecionarData(modo: 'hoje' | 'ontem' | 'outra') {
    setModoData(modo)
    if (modo === 'hoje') form.setIniciadoEm(hojeIso())
    if (modo === 'ontem') form.setIniciadoEm(ontem)
  }

  function selecionarSugestao(sugestao: IGDBJogoSugestao) {
    form.escolherSugestao(sugestao.id, formatarCapaIGDB(sugestao.cover?.url) || null, sugestao.name)
  }

  return (
    <Dialog open={isModalOpen} onOpenChange={(aberto) => !aberto && fecharModal()}>
      <DialogContent className="flex !max-h-[calc(100vh-32px)] w-[min(620px,calc(100vw-32px))] flex-col gap-5 overflow-y-auto rounded-2xl border border-[var(--modal-border)] bg-[var(--modal-bg)] px-6 py-6 text-[var(--text-primary)] sm:max-w-none">
        <DialogHeader>
          <DialogTitle className="text-[20px] font-bold">Iniciar jogo</DialogTitle>
          <DialogDescription className="text-[13px] text-[var(--text-secondary)]">Anote quando você começou. Ao terminar, é só clicar em Zerei!</DialogDescription>
        </DialogHeader>
        <form noValidate onSubmit={form.handleSubmit} className="flex flex-col gap-5">
          {form.errors.form && <p className="text-sm text-[var(--danger)]">{form.errors.form}</p>}
          <GameFormAutocomplete nome={form.nome} onChangeNome={form.setNome} onSelectSugestao={selecionarSugestao} error={form.errors.nome} />
          <div className={form.igdbId ? 'flex flex-col gap-5 sm:flex-row sm:items-start' : ''}>
            {form.igdbId && <div className="flex justify-center sm:justify-start"><div className="flex h-32 w-24 items-center justify-center sm:h-40 sm:w-[120px]">{form.igdbCapaUrl ? <img src={form.igdbCapaUrl} alt={`Capa de ${form.nome}`} className="h-full w-full rounded-md border border-[var(--border)] object-cover" /> : <div className="flex h-full w-full items-center justify-center rounded-md border border-dashed border-[var(--border-subtle)] bg-[linear-gradient(150deg,var(--bg-surface-alt),var(--bg-primary))]"><Gamepad2 className="size-8 text-[var(--text-faint)]" /></div>}</div></div>}
            <div className="flex flex-1 flex-col gap-3">
              <span className="text-sm font-medium text-[var(--text-secondary)]">Começou em</span>
              <div className="flex flex-wrap gap-2">
                {(['hoje', 'ontem', 'outra'] as const).map((modo) => <button key={modo} type="button" onClick={() => selecionarData(modo)} className={`rounded-full border px-3 py-1.5 text-xs font-semibold ${modoData === modo ? 'border-[var(--accent)] bg-[var(--accent)] text-[var(--accent-foreground)]' : 'border-[var(--border-subtle)] text-[var(--text-secondary)]'}`}>{modo === 'hoje' ? 'Hoje' : modo === 'ontem' ? 'Ontem' : 'Outra data'}</button>)}
              </div>
              {modoData === 'outra' && <DateInput id="iniciado_em" label="Data de início" value={form.iniciadoEm} onChange={form.setIniciadoEm} error={form.errors.iniciado_em} />}
              <p className="text-xs text-[var(--text-muted)]">Não precisa ser exata, uma data aproximada já ajuda. Não pode ser no futuro.</p>
              {modoData !== 'outra' && form.errors.iniciado_em && <span className="text-xs text-[var(--danger)]">{form.errors.iniciado_em}</span>}
            </div>
          </div>
          <div className="flex justify-end gap-3 border-t border-[var(--border)] pt-4"><Button type="button" variant="outline" onClick={fecharModal} disabled={form.isSubmitting}>Cancelar</Button><Button type="submit" disabled={form.isSubmitting || !form.nome.trim()} className="bg-[var(--accent)] font-bold text-[var(--accent-foreground)]">{form.isSubmitting ? 'Iniciando...' : 'Iniciar jogo'}</Button></div>
        </form>
      </DialogContent>
    </Dialog>
  )
}
