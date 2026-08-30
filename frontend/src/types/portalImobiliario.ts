export interface ConfiguracaoPortal {
  id?: string;
  portal: 'GRUPO_OLX';
  ativa: boolean;
  token_feed_prefixo?: string;
  nome_contato: string;
  email_contato: string;
  telefone_contato: string;
  exibicao_endereco: 'BAIRRO' | 'LOGRADOURO' | 'COMPLETO';
  atualizado_em?: string;
}

export interface CredencialFeedPortal {
  token: string;
  url_feed: string;
  url_webhook: string;
}

export interface PublicacaoPortal {
  imovel_id: string;
  titulo: string;
  tipo: string;
  finalidade: string;
  status: string;
  ativa: boolean;
  tipo_publicacao: 'STANDARD' | 'PREMIUM' | 'SUPER_PREMIUM';
  erros: string[];
}

export interface DiagnosticoFeedPortal {
  valido: boolean;
  total_selecionado: number;
  total_valido: number;
  erros_gerais: string[];
  publicacoes: PublicacaoPortal[];
}
