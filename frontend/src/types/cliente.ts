// Tipagens do módulo de CRM de Clientes/Leads

export interface Cliente {
  id?: string;
  conta_id?: string;
  nome: string;
  cpf?: string;
  data_nascimento?: string; // YYYY-MM-DD
  estado_civil?: string;
  email?: string;
  telefone?: string;
  status_funil: string;
  origem?: string;
  interesse_tipo?: string;
  notas?: string;
  tags?: string[];
  preferencias?: ClientePreferencias;
  financeiro?: ClienteFinanceiro;
  origem_utm?: ClienteOrigemUTM;
  corretor_id?: string;
  temperatura?: string; // Quente, Morno, Frio
  proxima_acao?: string;
  criado_em?: string;
  atualizado_em?: string;
}

export interface ClientePreferencias {
  tipo_imovel?: string[];
  finalidade?: string;
  orcamento_min?: number | null;
  orcamento_max?: number | null;
  bairros?: string[];
  quartos_min?: number | null;
}

export interface ClienteFinanceiro {
  renda_mensal?: number | null;
  precisa_financiamento?: string; // Sim, Nao, JaAprovado
  possui_fgts?: boolean;
  forma_pagamento?: string; // AVista, Financiamento, Consorcio, Parcelado
}

export interface ClienteOrigemUTM {
  canal?: string;
  campanha?: string;
  imovel_origem_id?: string;
}

export interface Interacao {
  id?: string;
  conta_id?: string;
  cliente_id: string;
  corretor_id?: string;
  tipo: string; // LIGACAO, MENSAGEM, VISITA, PROPOSTA, ANOTACAO
  descricao: string;
  data_hora?: string;
  criado_em?: string;
}

// Etapas do funil de vendas
export const STATUS_FUNIL = [
  { value: 'NOVO',         label: 'Novo Contato'    },
  { value: 'ATENDIMENTO',  label: 'Em Atendimento'  },
  { value: 'VISITA',       label: 'Visita Agendada' },
  { value: 'PROPOSTA',     label: 'Proposta'        },
  { value: 'FECHADO',      label: 'Fechado/Ganho'   },
  { value: 'PERDIDO',      label: 'Perdido'         },
];

// Origens possíveis de um lead
export const ORIGEM_LEAD = [
  { value: 'SITE',       label: 'Site Próprio'       },
  { value: 'PORTAL',     label: 'Portal Imobiliário' },
  { value: 'WHATSAPP',   label: 'WhatsApp'           },
  { value: 'INDICACAO',  label: 'Indicação'          },
  { value: 'SOCIAL',     label: 'Redes Sociais'      },
  { value: 'OUTROS',     label: 'Outros'             },
];
