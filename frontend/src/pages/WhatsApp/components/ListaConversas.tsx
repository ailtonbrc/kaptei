import { MessageCircle, Search } from 'lucide-react';
import type { ConversaWhatsApp } from '../../../types/whatsapp';

interface Props {
  conversas: ConversaWhatsApp[];
  conversaAtiva?: string;
  busca: string;
  carregando: boolean;
  aoBuscar: (valor: string) => void;
  aoSelecionar: (conversa: ConversaWhatsApp) => void;
}

export function ListaConversas({ conversas, conversaAtiva, busca, carregando, aoBuscar, aoSelecionar }: Props) {
  return (
    <aside className="flex min-h-0 flex-col border-r border-slate-200 bg-white lg:w-80">
      <div className="border-b border-slate-200 p-4">
        <label className="relative block">
          <span className="sr-only">Buscar conversa</span>
          <Search className="absolute left-3 top-2.5 h-4 w-4 text-slate-400" />
          <input
            value={busca}
            onChange={(evento) => aoBuscar(evento.target.value)}
            placeholder="Buscar nome ou telefone"
            className="h-9 w-full rounded-lg border border-slate-200 bg-slate-50 pl-9 pr-3 text-sm outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
          />
        </label>
      </div>
      <div className="min-h-0 flex-1 overflow-y-auto">
        {carregando && <p className="p-5 text-sm text-slate-500">Carregando conversas...</p>}
        {!carregando && conversas.length === 0 && (
          <div className="p-8 text-center text-slate-500">
            <MessageCircle className="mx-auto mb-3 h-9 w-9 text-slate-300" />
            <p className="text-sm">Nenhuma conversa encontrada.</p>
          </div>
        )}
        {conversas.map((conversa) => (
          <button
            key={conversa.id}
            type="button"
            onClick={() => aoSelecionar(conversa)}
            className={`w-full border-b border-slate-100 p-4 text-left transition-colors hover:bg-slate-50 ${conversaAtiva === conversa.id ? 'bg-blue-50' : ''}`}
          >
            <div className="flex items-start justify-between gap-2">
              <strong className="truncate text-sm text-slate-900">{conversa.nome_contato || conversa.numero_contato}</strong>
              <time className="shrink-0 text-[11px] text-slate-400">
                {new Date(conversa.ultima_mensagem_em).toLocaleDateString('pt-BR')}
              </time>
            </div>
            <p className="mt-1 text-xs text-slate-500">+{conversa.numero_contato}</p>
          </button>
        ))}
      </div>
    </aside>
  );
}
