import { useCallback, useEffect, useState } from 'react';
import { toast } from 'sonner';
import { agendamentosService } from '@/services/agendamentosService';
import type { Agendamento, AgendamentoInput } from '@/types/agendamento';

export function useAgendamentos(inicio: Date, fim: Date) {
  const [agendamentos, setAgendamentos] = useState<Agendamento[]>([]);
  const [carregando, setCarregando] = useState(true);

  const carregar = useCallback(async () => {
    try {
	  const dados = await agendamentosService.listar(inicio, fim);
	  setAgendamentos(dados);
    } catch {
      toast.error('Não foi possível carregar os agendamentos.');
    } finally {
      setCarregando(false);
    }
  }, [fim, inicio]);

  useEffect(() => {
    // A atualização ocorre somente após a resposta assíncrona da API.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void carregar();
  }, [carregar]);

  const salvar = async (dados: AgendamentoInput, id?: string) => {
	setCarregando(true);
    if (id) {
      await agendamentosService.atualizar(id, dados);
      toast.success('Agendamento atualizado.');
    } else {
      await agendamentosService.criar(dados);
      toast.success('Agendamento criado.');
    }
    await carregar();
  };

  const excluir = async (id: string) => {
	setCarregando(true);
    await agendamentosService.excluir(id);
    toast.success('Agendamento excluído.');
    await carregar();
  };

  return { agendamentos, carregando, salvar, excluir, recarregar: carregar };
}
