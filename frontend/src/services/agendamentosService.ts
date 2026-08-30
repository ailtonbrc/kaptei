import { api } from './api';
import type { Agendamento, AgendamentoInput } from '@/types/agendamento';

const recurso = '/v1/negocios/agendamentos';

export const agendamentosService = {
  async listar(inicio: Date, fim: Date): Promise<Agendamento[]> {
    const response = await api.get<Agendamento[]>(recurso, {
      params: { inicio: inicio.toISOString(), fim: fim.toISOString() },
    });
    return response.data ?? [];
  },

  async criar(dados: AgendamentoInput): Promise<Agendamento> {
    const response = await api.post<Agendamento>(recurso, dados);
    return response.data;
  },

  async atualizar(id: string, dados: AgendamentoInput): Promise<Agendamento> {
    const response = await api.put<Agendamento>(`${recurso}/${id}`, dados);
    return response.data;
  },

  async excluir(id: string): Promise<void> {
    await api.delete(`${recurso}/${id}`);
  },
};
