import type { ListaPaginada } from './paginacao';

export type TipoSolicitacaoTitular =
  | 'CONFIRMACAO' | 'ACESSO' | 'CORRECAO' | 'ANONIMIZACAO' | 'BLOQUEIO'
  | 'EXCLUSAO' | 'PORTABILIDADE' | 'REVOGACAO' | 'INFORMACAO_COMPARTILHAMENTO';

export interface NovaSolicitacaoTitular {
  tipo: TipoSolicitacaoTitular;
  nome: string;
  email: string;
  telefone: string;
  detalhes: string;
}

export interface EventoSolicitacaoTitular {
  id: string;
  tipo: string;
  descricao: string;
  usuario_id?: string;
  criado_em: string;
}

export interface SolicitacaoTitular {
  id: string;
  conta_id: string;
  protocolo: string;
  tipo: TipoSolicitacaoTitular;
  nome?: string;
  email?: string;
  telefone?: string;
  detalhes?: string;
  status: 'RECEBIDA' | 'VALIDANDO_IDENTIDADE' | 'EM_ANALISE' | 'APROVADA' | 'REJEITADA' | 'CONCLUIDA';
  prazo_resposta_em: string;
  identidade_verificada_em?: string;
  metodo_verificacao?: string;
  decisao?: 'APROVADA' | 'REJEITADA';
  fundamento_legal?: string;
  observacao_decisao?: string;
  decidida_em?: string;
  executada_em?: string;
  criado_em: string;
  atualizado_em: string;
  eventos?: EventoSolicitacaoTitular[];
}

export interface PoliticaRetencao {
  ativa: boolean;
  dias_leads_descartados: number;
  dias_clientes_perdidos: number;
  tamanho_lote: number;
  fundamento_legal: string;
  ultima_execucao_em?: string;
  atualizado_em?: string;
}

export interface RelatorioRetencao {
  leads_elegiveis: number;
  clientes_elegiveis: number;
  bloqueios_vigentes: number;
}

export interface ResultadoRetencao {
  leads_anonimizados: number;
  clientes_anonimizados: number;
  relatorio_restante: RelatorioRetencao;
}

export interface BloqueioRetencao {
  id: string;
  tipo_recurso: 'LEAD' | 'CLIENTE';
  recurso_id: string;
  motivo: string;
  valido_ate?: string;
  criado_em: string;
}

export interface NovoBloqueioRetencao {
  tipo_recurso: 'LEAD' | 'CLIENTE';
  recurso_id: string;
  motivo: string;
  valido_ate?: string;
}

export type PaginaSolicitacoesTitular = ListaPaginada<SolicitacaoTitular>;

