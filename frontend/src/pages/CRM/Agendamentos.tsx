import { useMemo, useState } from 'react';
import { Calendar, dateFnsLocalizer, Views, type SlotInfo, type View } from 'react-big-calendar';
import { addDays, endOfDay, endOfMonth, endOfWeek, format, getDay, parse, startOfDay, startOfMonth, startOfWeek } from 'date-fns';
import { ptBR } from 'date-fns/locale';
import { CalendarDays, Loader2, Plus } from 'lucide-react';
import 'react-big-calendar/lib/css/react-big-calendar.css';
import { Button } from '@/components/ui/button';
import { useAgendamentos } from '@/hooks/useAgendamentos';
import type { Agendamento } from '@/types/agendamento';
import { AgendamentoDialog } from './components/AgendamentoDialog';

const localizer = dateFnsLocalizer({
  format,
  parse,
  startOfWeek: (data: Date) => startOfWeek(data, { weekStartsOn: 0 }),
  getDay,
  locales: { 'pt-BR': ptBR },
});

interface EventoAgenda {
  id: string;
  title: string;
  start: Date;
  end: Date;
  resource: Agendamento;
}

function calcularPeriodo(data: Date, visualizacao: View) {
  if (visualizacao === Views.DAY) return { inicio: startOfDay(data), fim: endOfDay(data) };
  if (visualizacao === Views.WEEK || visualizacao === Views.WORK_WEEK) {
    return { inicio: startOfWeek(data), fim: endOfWeek(data) };
  }
  if (visualizacao === Views.AGENDA) return { inicio: startOfDay(data), fim: endOfDay(addDays(data, 30)) };
  return { inicio: startOfWeek(startOfMonth(data)), fim: endOfWeek(endOfMonth(data)) };
}

export function Agendamentos() {
  const [visualizacao, setVisualizacao] = useState<View>(Views.MONTH);
  const [data, setData] = useState(new Date());
  const [dialogAberto, setDialogAberto] = useState(false);
  const [selecionado, setSelecionado] = useState<Agendamento>();
  const [intervaloSugerido, setIntervaloSugerido] = useState<{ inicio: Date; fim: Date }>();
  const [chaveDialog, setChaveDialog] = useState(0);

  const periodo = useMemo(() => calcularPeriodo(data, visualizacao), [data, visualizacao]);
  const { agendamentos, carregando, salvar, excluir } = useAgendamentos(periodo.inicio, periodo.fim);
  const eventos = useMemo<EventoAgenda[]>(() => agendamentos.map((agendamento) => ({
    id: agendamento.id,
    title: `${agendamento.titulo}${agendamento.cliente_nome ? ` â€” ${agendamento.cliente_nome}` : ''}`,
    start: new Date(agendamento.data_hora_inicio),
    end: new Date(agendamento.data_hora_fim),
    resource: agendamento,
  })), [agendamentos]);

  const abrirNovo = (inicio = new Date(), fim = new Date(Date.now() + 60 * 60_000)) => {
    setSelecionado(undefined);
    setIntervaloSugerido({ inicio, fim });
    setChaveDialog((atual) => atual + 1);
    setDialogAberto(true);
  };

  const abrirEdicao = (evento: EventoAgenda) => {
    setSelecionado(evento.resource);
    setIntervaloSugerido(undefined);
    setChaveDialog((atual) => atual + 1);
    setDialogAberto(true);
  };

  const selecionarIntervalo = (slot: SlotInfo) => abrirNovo(slot.start, slot.end);

  return (
    <div className="flex min-h-full flex-col bg-slate-50 p-4 sm:p-6">
      <header className="mb-6 flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <h1 className="flex items-center gap-2 text-2xl font-extrabold tracking-tight text-slate-900">
            <CalendarDays className="h-7 w-7 text-blue-600" /> Agenda comercial
          </h1>
          <p className="mt-1 text-sm text-slate-500">Gerencie visitas, reuniões e próximas ações do relacionamento.</p>
        </div>
        <Button onClick={() => abrirNovo()}><Plus className="h-4 w-4" /> Novo agendamento</Button>
      </header>

      <section className="relative min-h-[650px] flex-1 rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
        {carregando && (
          <div className="absolute inset-0 z-10 flex items-center justify-center rounded-xl bg-white/70 backdrop-blur-sm">
            <Loader2 className="h-9 w-9 animate-spin text-blue-600" />
          </div>
        )}
        <Calendar<EventoAgenda>
          localizer={localizer}
          events={eventos}
          startAccessor="start"
          endAccessor="end"
          culture="pt-BR"
          view={visualizacao}
          date={data}
          onView={setVisualizacao}
          onNavigate={setData}
          onSelectEvent={abrirEdicao}
          onSelectSlot={selecionarIntervalo}
          selectable
          popup
          style={{ height: '100%', minHeight: 610 }}
          messages={{
            next: 'Próximo', previous: 'Anterior', today: 'Hoje', month: 'Mês', week: 'Semana', day: 'Dia',
            agenda: 'Agenda', date: 'Data', time: 'Hora', event: 'Compromisso', noEventsInRange: 'Nenhum compromisso neste período.',
          }}
          eventPropGetter={(evento) => ({
            style: {
              backgroundColor: evento.resource.status === 'CANCELADO' ? '#ef4444'
                : evento.resource.status === 'CONCLUIDO' ? '#10b981'
                  : evento.resource.status === 'CONFIRMADO' ? '#8b5cf6' : '#2563eb',
              borderRadius: 6,
              border: 0,
            },
          })}
        />
      </section>

      {dialogAberto && (
        <AgendamentoDialog
          key={chaveDialog}
          aberto={dialogAberto}
          agendamento={selecionado}
          inicioSugerido={intervaloSugerido?.inicio ?? (selecionado ? new Date(selecionado.data_hora_inicio) : data)}
          fimSugerido={intervaloSugerido?.fim ?? (selecionado ? new Date(selecionado.data_hora_fim) : addDays(data, 1))}
          onOpenChange={setDialogAberto}
          onSalvar={salvar}
          onExcluir={excluir}
        />
      )}
    </div>
  );
}
