import { api } from './api';

export interface MembroEquipe {
  id: string;
  nome: string;
  email: string;
  papel: 'GESTOR' | 'CORRETOR_EQUIPE';
  status: string;
  url_avatar?: string;
}

export interface ConviteEquipe {
  id: string;
  email: string;
  papel: 'CORRETOR_EQUIPE';
  expira_em: string;
  criado_em: string;
}

export interface ResumoEquipe {
  membros: MembroEquipe[];
  convites: ConviteEquipe[];
}

export const equipeService = {
  async listar(): Promise<ResumoEquipe> {
    const resposta = await api.get<ResumoEquipe>('/v1/equipe');
    return resposta.data;
  },
  async convidar(email: string): Promise<void> {
    await api.post('/v1/equipe/convites', { email });
  },
  async cancelarConvite(id: string): Promise<void> {
    await api.delete(`/v1/equipe/convites/${encodeURIComponent(id)}`);
  },
  async atualizarStatus(id: string, status: 'ATIVO' | 'INATIVO'): Promise<void> {
    await api.patch(`/v1/equipe/${encodeURIComponent(id)}/status`, { status });
  },
  async aceitarConvite(token: string, nome: string, senha: string): Promise<void> {
    await api.post('/auth/aceitar-convite', { token, nome, senha });
  },
};
