export interface DominioSite {
  id: string;
  hostname: string;
  status: 'PENDENTE' | 'ATIVO' | 'FALHOU';
  registro_txt_nome: string;
  registro_txt_valor: string;
  verificado_em?: string;
  ultima_verificacao_em?: string;
  ultimo_erro?: string;
  criado_em: string;
  atualizado_em: string;
}

