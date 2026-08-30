export type StatusAgendamento = 'AGENDADO' | 'CONFIRMADO' | 'CONCLUIDO' | 'CANCELADO' | 'NAO_COMPARECEU';
export type TipoAgendamento = 'VISITA' | 'LIGACAO' | 'REUNIAO_ONLINE' | 'TAREFA';

export interface Agendamento {
  id: string;
  usuario_id: string;
  cliente_id?: string | null;
  imovel_id?: string | null;
  titulo: string;
  descricao: string;
  data_hora_inicio: string;
  data_hora_fim: string;
  status: StatusAgendamento;
  tipo: TipoAgendamento;
  cliente_nome?: string;
  imovel_titulo?: string;
}

export type AgendamentoInput = Pick<
  Agendamento,
  'titulo' | 'descricao' | 'data_hora_inicio' | 'data_hora_fim' | 'status' | 'tipo'
> & {
  cliente_id?: string | null;
  imovel_id?: string | null;
};
