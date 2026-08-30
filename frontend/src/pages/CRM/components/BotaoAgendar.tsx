import { useState } from 'react';
import { CalendarPlus } from 'lucide-react';
import { toast } from 'sonner';
import { Button } from '@/components/ui/button';
import { agendamentosService } from '@/services/agendamentosService';
import type { AgendamentoInput } from '@/types/agendamento';
import { obterMensagemErro } from '@/lib/http/erro-api';
import { AgendamentoDialog } from './AgendamentoDialog';

interface BotaoAgendarProps {
  clienteID?: string;
  imovelID?: string;
  contexto: string;
}

export function BotaoAgendar({ clienteID, imovelID, contexto }: BotaoAgendarProps) {
  const [aberto, setAberto] = useState(false);
  const agora = new Date();
  const fim = new Date(agora.getTime() + 60 * 60_000);

  async function salvar(dados: AgendamentoInput) {
    try {
      await agendamentosService.criar(dados);
      toast.success('Agendamento criado e vinculado com sucesso.');
    } catch (erro: unknown) {
      toast.error(obterMensagemErro(erro, 'Não foi possível criar o agendamento.'));
      throw erro;
    }
  }

  return <>
    <Button type="button" variant="outline" onClick={() => setAberto(true)}><CalendarPlus className="h-4 w-4" />Agendar visita</Button>
    {aberto && <AgendamentoDialog
      aberto={aberto}
      inicioSugerido={agora}
      fimSugerido={fim}
      clienteIDSugerido={clienteID}
      imovelIDSugerido={imovelID}
      contextoSugerido={contexto}
      onOpenChange={setAberto}
      onSalvar={salvar}
    />}
  </>;
}
