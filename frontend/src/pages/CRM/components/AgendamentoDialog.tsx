import { useState, type FormEvent } from 'react';
import { Loader2, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import type { Agendamento, AgendamentoInput, StatusAgendamento, TipoAgendamento } from '@/types/agendamento';

interface AgendamentoDialogProps {
  aberto: boolean;
  agendamento?: Agendamento;
  inicioSugerido: Date;
  fimSugerido: Date;
	clienteIDSugerido?: string;
	imovelIDSugerido?: string;
	contextoSugerido?: string;
  onOpenChange: (aberto: boolean) => void;
  onSalvar: (dados: AgendamentoInput, id?: string) => Promise<void>;
	onExcluir?: (id: string) => Promise<void>;
}

const paraDataLocal = (data: Date) => {
  const deslocamento = data.getTimezoneOffset() * 60_000;
  return new Date(data.getTime() - deslocamento).toISOString().slice(0, 16);
};

export function AgendamentoDialog({
  aberto,
  agendamento,
  inicioSugerido,
  fimSugerido,
	clienteIDSugerido,
	imovelIDSugerido,
	contextoSugerido,
  onOpenChange,
  onSalvar,
  onExcluir,
}: AgendamentoDialogProps) {
  const [titulo, setTitulo] = useState(agendamento?.titulo ?? '');
  const [descricao, setDescricao] = useState(agendamento?.descricao ?? '');
  const [inicio, setInicio] = useState(paraDataLocal(agendamento ? new Date(agendamento.data_hora_inicio) : inicioSugerido));
  const [fim, setFim] = useState(paraDataLocal(agendamento ? new Date(agendamento.data_hora_fim) : fimSugerido));
  const [tipo, setTipo] = useState<TipoAgendamento>(agendamento?.tipo ?? 'VISITA');
  const [status, setStatus] = useState<StatusAgendamento>(agendamento?.status ?? 'AGENDADO');
  const [salvando, setSalvando] = useState(false);
  const [confirmarExclusao, setConfirmarExclusao] = useState(false);
  const [erro, setErro] = useState('');

  const handleSubmit = async (evento: FormEvent) => {
    evento.preventDefault();
    const dataInicio = new Date(inicio);
    const dataFim = new Date(fim);
    if (!titulo.trim() || Number.isNaN(dataInicio.getTime()) || Number.isNaN(dataFim.getTime()) || dataFim <= dataInicio) {
      setErro('Informe um título e um período válido.');
      return;
    }

    setSalvando(true);
    setErro('');
    try {
      await onSalvar(
        {
          titulo: titulo.trim(),
          descricao: descricao.trim(),
          data_hora_inicio: dataInicio.toISOString(),
          data_hora_fim: dataFim.toISOString(),
          tipo,
		  status,
		  cliente_id: agendamento?.cliente_id ?? clienteIDSugerido ?? null,
		  imovel_id: agendamento?.imovel_id ?? imovelIDSugerido ?? null,
        },
        agendamento?.id,
      );
      onOpenChange(false);
    } catch {
      setErro('Não foi possível salvar o agendamento.');
    } finally {
      setSalvando(false);
    }
  };

  const handleExcluir = async () => {
	if (!agendamento || !onExcluir) return;
    if (!confirmarExclusao) {
      setConfirmarExclusao(true);
      return;
    }
    setSalvando(true);
    try {
      await onExcluir(agendamento.id);
      onOpenChange(false);
    } catch {
      setErro('Não foi possível excluir o agendamento.');
    } finally {
      setSalvando(false);
    }
  };

  return (
    <Dialog open={aberto} onOpenChange={onOpenChange}>
      <DialogContent>
        <form onSubmit={handleSubmit} className="space-y-5">
          <DialogHeader>
            <DialogTitle>{agendamento ? 'Editar agendamento' : 'Novo agendamento'}</DialogTitle>
		  <DialogDescription>Registre visitas, reuniões, ligações e tarefas comerciais.</DialogDescription>
		  {contextoSugerido && <p className="mt-2 rounded-lg bg-blue-50 px-3 py-2 text-sm font-medium text-blue-800">Vinculado a: {contextoSugerido}</p>}
          </DialogHeader>

          <div className="space-y-2">
            <Label htmlFor="agendamento-titulo">Título</Label>
            <Input id="agendamento-titulo" value={titulo} onChange={(e) => setTitulo(e.target.value)} maxLength={255} required />
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="agendamento-inicio">Início</Label>
              <Input id="agendamento-inicio" type="datetime-local" value={inicio} onChange={(e) => setInicio(e.target.value)} required />
            </div>
            <div className="space-y-2">
              <Label htmlFor="agendamento-fim">Fim</Label>
              <Input id="agendamento-fim" type="datetime-local" value={fim} onChange={(e) => setFim(e.target.value)} required />
            </div>
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="agendamento-tipo">Tipo</Label>
              <select id="agendamento-tipo" className="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm" value={tipo} onChange={(e) => setTipo(e.target.value as TipoAgendamento)}>
                <option value="VISITA">Visita</option>
                <option value="LIGACAO">Ligação</option>
                <option value="REUNIAO_ONLINE">Reunião online</option>
                <option value="TAREFA">Tarefa</option>
              </select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="agendamento-status">Status</Label>
              <select id="agendamento-status" className="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm" value={status} onChange={(e) => setStatus(e.target.value as StatusAgendamento)}>
                <option value="AGENDADO">Agendado</option>
                <option value="CONFIRMADO">Confirmado</option>
                <option value="CONCLUIDO">Concluído</option>
                <option value="CANCELADO">Cancelado</option>
                <option value="NAO_COMPARECEU">Não compareceu</option>
              </select>
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="agendamento-descricao">Descrição</Label>
            <Textarea id="agendamento-descricao" value={descricao} onChange={(e) => setDescricao(e.target.value)} maxLength={2000} />
          </div>

          {erro && <p className="text-sm text-red-600">{erro}</p>}

          <DialogFooter className="sm:justify-between">
		  {agendamento && onExcluir ? (
              <Button type="button" variant="destructive" onClick={handleExcluir} disabled={salvando}>
                <Trash2 className="h-4 w-4" />
                {confirmarExclusao ? 'Confirmar exclusão' : 'Excluir'}
              </Button>
            ) : <span />}
            <div className="flex justify-end gap-2">
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={salvando}>Cancelar</Button>
              <Button type="submit" disabled={salvando}>
                {salvando && <Loader2 className="h-4 w-4 animate-spin" />}
                Salvar
              </Button>
            </div>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
