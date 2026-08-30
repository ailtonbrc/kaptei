export interface ConfiguracaoSitePublico {
  logo_url?: string;
  cor_primaria?: string;
  titulo?: string;
  subtitulo?: string;
  descricao?: string;
  telefone?: string;
  whatsapp?: string;
  email?: string;
  cidade?: string;
  creci?: string;
}

export interface SitePublico {
  slug: string;
  nome: string;
  publicado: boolean;
  configuracao: ConfiguracaoSitePublico;
}

export interface FotoImovelPublico {
  id: string;
  url: string;
  url_thumbnail?: string;
  ordem: number;
  is_capa: boolean;
}

export interface ImovelPublico {
  id: string;
  slug: string;
  titulo: string;
  tipo: string;
  finalidade: string;
  valor_venda?: number;
  valor_locacao?: number;
  valor_condominio?: number;
  valor_iptu?: number;
  area_total?: number;
  area_util?: number;
  quartos: number;
  suites: number;
  banheiros: number;
  vagas: number;
  bairro?: string;
  cidade?: string;
  estado?: string;
  descricao?: string;
  titulo_seo?: string;
  descricao_seo?: string;
  destaque: boolean;
  fotos: FotoImovelPublico[];
}

export interface FiltrosCatalogo {
  tipo?: string;
  finalidade?: string;
  cidade?: string;
  bairro?: string;
  valor_min?: number;
  valor_max?: number;
  quartos_min?: number;
  pagina?: number;
  limite?: number;
}

export interface CapturaLeadSite {
  nome: string;
  email?: string;
  telefone?: string;
  mensagem?: string;
  imovel_slug?: string;
  pagina_origem?: string;
  utm_source?: string;
  utm_medium?: string;
  utm_campaign?: string;
  consentimento_lgpd: boolean;
  chave_idempotencia: string;
  website?: string;
}

export type TipoEventoConversao =
  | 'SITE_VISUALIZADO'
  | 'IMOVEL_VISUALIZADO'
  | 'FORMULARIO_INICIADO'
  | 'LEAD_ENVIADO'
  | 'WHATSAPP_CLICADO'
  | 'TELEFONE_CLICADO';

export interface AtribuicaoConversao {
  utm_source?: string;
  utm_medium?: string;
  utm_campaign?: string;
}

export interface EventoConversaoSite extends AtribuicaoConversao {
  chave_evento: string;
  sessao_id: string;
  tipo: TipoEventoConversao;
  imovel_slug?: string;
}
