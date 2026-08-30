import { Check, CheckCheck, Clock3, TriangleAlert } from 'lucide-react';
import type { MensagemWhatsApp } from '../../../types/whatsapp';

function IndicadorStatus({ mensagem }: { mensagem: MensagemWhatsApp }) {
  if (mensagem.status === 'FALHOU') return <TriangleAlert className="h-3.5 w-3.5 text-rose-500" aria-label="Falha no envio" />;
  if (mensagem.status === 'PENDENTE') return <Clock3 className="h-3.5 w-3.5" aria-label="Envio pendente" />;
  if (mensagem.status === 'LIDA') return <CheckCheck className="h-3.5 w-3.5 text-blue-600" aria-label="Mensagem lida" />;
  if (mensagem.status === 'ENTREGUE') return <CheckCheck className="h-3.5 w-3.5" aria-label="Mensagem entregue" />;
  return <Check className="h-3.5 w-3.5" aria-label="Mensagem enviada" />;
}

export function HistoricoMensagens({ mensagens, carregando }: { mensagens: MensagemWhatsApp[]; carregando: boolean }) {
  if (carregando) return <p className="m-auto text-sm text-slate-500">Carregando mensagens...</p>;
  if (mensagens.length === 0) return <p className="m-auto text-sm text-slate-500">Nenhuma mensagem nesta conversa.</p>;

  return (
    <div className="flex flex-col-reverse gap-3 p-4 sm:p-6">
      {mensagens.map((mensagem) => {
        const saida = mensagem.direcao === 'SAIDA';
        return (
          <article key={mensagem.id} className={`max-w-[85%] rounded-2xl px-4 py-2.5 shadow-sm ${saida ? 'ml-auto rounded-br-sm bg-blue-600 text-white' : 'mr-auto rounded-bl-sm border border-slate-200 bg-white text-slate-800'}`}>
            <p className="whitespace-pre-wrap break-words text-sm">{mensagem.conteudo || `[${mensagem.tipo}]`}</p>
            <div className={`mt-1 flex items-center justify-end gap-1 text-[10px] ${saida ? 'text-blue-100' : 'text-slate-400'}`}>
              <time>{new Date(mensagem.ocorrida_em).toLocaleString('pt-BR', { hour: '2-digit', minute: '2-digit', day: '2-digit', month: '2-digit' })}</time>
              {saida && <IndicadorStatus mensagem={mensagem} />}
            </div>
            {mensagem.erro_detalhe && <p className="mt-1 text-xs text-rose-100">{mensagem.erro_detalhe}</p>}
          </article>
        );
      })}
    </div>
  );
}
