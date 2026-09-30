import { useContext, useEffect, useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { ApiError } from "@/lib/api";
import { AuthContext } from "@/store/authStore";
import { useListasStore } from "@/stores/listasStore";
import { mapearErroApiParaCampo } from "./listas.utils";
import { LISTAS_ERRO_GENERICO } from "./listas.constants";
import { novaListaSchema } from "./novaLista.schema";
import type { ListaOrigem, ListaTipo } from "@/types/listas";

export interface ConfiguracaoDesafio {
  nome: string;
  descricao: string | null;
  origem: ListaOrigem;
}

export function useNovaListaForm(
  onClose: () => void,
  onVerJogos: (config: ConfiguracaoDesafio) => void,
) {
  const navigate = useNavigate();
  const token = useContext(AuthContext)?.sessao?.access_token;
  const { listaEmEdicao, criarLista, atualizarLista } = useListasStore();
  const [tipo, setTipo] = useState<ListaTipo>(listaEmEdicao?.tipo ?? "fila");
  const [nome, setNome] = useState(listaEmEdicao?.nome ?? "");
  const [descricao, setDescricao] = useState(listaEmEdicao?.descricao ?? "");
  const [origemTipo, setOrigemTipo] = useState<
    "franquia" | "plataforma" | "genero"
  >(listaEmEdicao?.origem?.tipo ?? "franquia");
  const [origem, setOrigem] = useState<ListaOrigem | null>(
    listaEmEdicao?.origem ?? null,
  );
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    setTipo(listaEmEdicao?.tipo ?? "fila");
    setNome(listaEmEdicao?.nome ?? "");
    setDescricao(listaEmEdicao?.descricao ?? "");
    setOrigemTipo(listaEmEdicao?.origem?.tipo ?? "franquia");
    setOrigem(listaEmEdicao?.origem ?? null);
    setErrors({});
  }, [listaEmEdicao]);

  const escolherOrigem = (valor: ListaOrigem | null) => setOrigem(valor);
  const trocarOrigemTipo = (valor: typeof origemTipo) => {
    setOrigemTipo(valor);
    setOrigem(null);
  };
  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    setErrors({});
    const parsed = novaListaSchema.safeParse({
      tipo,
      nome,
      descricao: descricao || undefined,
      origemTipo: tipo === "desafio" ? origemTipo : undefined,
      origemId: origem?.igdb_id ?? undefined,
      origemNome: origem?.nome,
    });
    if (!parsed.success) {
      setErrors(
        Object.fromEntries(
          parsed.error.issues.map((issue) => [
            String(issue.path[0]),
            issue.message,
          ]),
        ),
      );
      return;
    }
    setIsSubmitting(true);
    try {
      if (listaEmEdicao) {
        await atualizarLista(listaEmEdicao.id, {
          nome: parsed.data.nome,
          descricao: parsed.data.descricao ?? null,
        });
        onClose();
        return;
      }
      if (tipo === "desafio" && origem) {
        onVerJogos({
          nome: parsed.data.nome,
          descricao: parsed.data.descricao ?? null,
          origem,
        });
        return;
      }
      const criada = await criarLista({
        tipo,
        nome: parsed.data.nome,
        descricao: parsed.data.descricao ?? null,
      });
      onClose();
      navigate(`/listas/${criada.id}`);
    } catch (err: unknown) {
      const resposta =
        err instanceof ApiError
          ? mapearErroApiParaCampo(err.codigo)
          : { campo: "form", mensagem: LISTAS_ERRO_GENERICO };
      setErrors((anterior) => ({
        ...anterior,
        [resposta.campo]: resposta.mensagem,
      }));
    } finally {
      setIsSubmitting(false);
    }
  };

  return {
    token,
    tipo,
    setTipo,
    nome,
    setNome,
    descricao,
    setDescricao,
    origemTipo,
    trocarOrigemTipo,
    origem,
    escolherOrigem,
    errors,
    isSubmitting,
    handleSubmit,
    isEdicao: Boolean(listaEmEdicao),
  };
}
