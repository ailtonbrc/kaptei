import type { ListaPaginada } from './paginacao';

export interface ConversaWhatsApp {
  id: string;
  lead_id?: string;
  numero_contato: string;
  nome_contato?: string;
  consentimento_marketing: boolean;
  janela_atendimento_ate?: string;
  ultima_mensagem_em: string;
}

export interface MensagemWhatsApp {
  id: string;
  conversa_id: string;
  identificador_externo?: string;
  direcao: 'ENTRADA' | 'SAIDA';
  tipo: string;
  conteudo: string;
  status: 'RECEBIDA' | 'PENDENTE' | 'ENVIADA' | 'ENTREGUE' | 'LIDA' | 'FALHOU';
  ocorrida_em: string;
  erro_detalhe?: string;
}

export type ListaConversasWhatsApp = ListaPaginada<ConversaWhatsApp>;
export type ListaMensagensWhatsApp = ListaPaginada<MensagemWhatsApp>;

export interface EnvioTemplateWhatsApp {
  nome: string;
  idioma: string;
  parametros: string[];
}
