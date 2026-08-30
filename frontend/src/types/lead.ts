export interface Lead {
  id?: string;
  conta_id?: string;
  usuario_id?: string | null;
  imovel_id?: string | null;
  
  nome: string;
  email?: string;
  telefone?: string;
  origem?: string;
  mensagem?: string;
  
  status: 'NOVO' | 'EM_ATENDIMENTO' | 'QUALIFICADO' | 'DESCARTADO';
  motivo_descarte?: string;
  
  criado_em?: string;
  atualizado_em?: string;
}
