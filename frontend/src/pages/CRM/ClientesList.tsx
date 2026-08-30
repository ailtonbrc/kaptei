import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Plus, Search, Filter, Phone, MoreVertical, Calendar, XCircle, Mail, MessageSquare, User, Briefcase, Handshake } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { clientesService } from '../../services/clientesService';
import { Button } from '../../components';
import { toast } from 'sonner';
import type { Cliente } from '../../types/cliente';
import { STATUS_FUNIL } from '../../types/cliente';
import { usePaginacao } from '../../hooks/usePaginacao';

// Configurações visuais do Funil Premium
const FUNIL_COLORS: Record<string, { bg: string, headerBg: string, border: string, text: string, icon: LucideIcon }> = {
  'NOVO':        { bg: 'bg-sky-50',     headerBg: 'bg-sky-100/50',   border: 'border-sky-200',   text: 'text-sky-700',   icon: User },
  'ATENDIMENTO': { bg: 'bg-amber-50',   headerBg: 'bg-amber-100/50', border: 'border-amber-200', text: 'text-amber-700', icon: MessageSquare },
  'VISITA':      { bg: 'bg-fuchsia-50', headerBg: 'bg-fuchsia-100/50',border: 'border-fuchsia-200',text: 'text-fuchsia-700', icon: Calendar },
  'PROPOSTA':    { bg: 'bg-indigo-50',  headerBg: 'bg-indigo-100/50', border: 'border-indigo-200', text: 'text-indigo-700', icon: Briefcase },
  'FECHADO':     { bg: 'bg-emerald-50', headerBg: 'bg-emerald-100/50',border: 'border-emerald-200',text: 'text-emerald-700', icon: Handshake },
  'PERDIDO':     { bg: 'bg-rose-50',    headerBg: 'bg-rose-100/50',   border: 'border-rose-200',   text: 'text-rose-700',   icon: XCircle },
};

export const ClientesList: React.FC = () => {
  const navigate = useNavigate();
  const {
    items: clientes,
    setItems: setClientes,
    loading,
    pagina,
    total,
    searchTerm,
    setSearchTerm,
    carregarItems: carregarClientes,
  } = usePaginacao<Cliente>({
    fetchData: async (pag, busca) => await clientesService.listar({ pagina: pag, busca })
  });

  const handleDragStart = (e: React.DragEvent, id: string) => {
    e.dataTransfer.setData('clienteId', id);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
  };

  const handleDrop = async (e: React.DragEvent, novoStatus: string) => {
    e.preventDefault();
    const id = e.dataTransfer.getData('clienteId');
    if (!id) return;

    const cliente = clientes.find(c => c.id === id);
    if (!cliente || cliente.status_funil === novoStatus) return;

    // Optimistic update
    setClientes(prev => prev.map(c => 
      c.id === id ? { ...c, status_funil: novoStatus } : c
    ));

    try {
      await clientesService.atualizar(id, { ...cliente, status_funil: novoStatus });
    } catch (error) {
      toast.error('Não foi possível mover a oportunidade.');
      console.error('Erro ao atualizar status', error);
      carregarClientes(); // Revert
    }
  };

  const filteredClientes = clientes.filter(c => 
    c.nome.toLowerCase().includes(searchTerm.toLowerCase()) || 
    (c.email && c.email.toLowerCase().includes(searchTerm.toLowerCase()))
  );

  const getInitials = (name: string) => {
    return name.split(' ').slice(0, 2).map(n => n[0]).join('').toUpperCase();
  };


  return (
    <div className="h-full flex flex-col p-4 sm:p-6 overflow-hidden bg-slate-50/30">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
        <div>
          <h1 className="text-2xl font-extrabold text-slate-900 tracking-tight">Pipeline de Vendas</h1>
          <p className="text-sm text-slate-500 mt-1">Gerencie oportunidades e arraste os cards para avançar.</p>
        </div>
        <Button 
          variant="outlined"
          icon={<Plus className="w-5 h-5" />}
          onClick={() => navigate('/app/crm/novo')}
        >
          Nova Oportunidade
        </Button>
      </div>

      {/* Filtros Premium */}
      <div className="flex items-center gap-3 mb-6">
        <div className="relative flex-1 max-w-md group">
          <Search className="absolute left-3.5 top-2.5 text-slate-400 w-5 h-5 group-focus-within:text-blue-600 transition-colors" />
          <input 
            type="text" 
            placeholder="Buscar leads por nome, e-mail..." 
            value={searchTerm}
			onChange={(e) => setSearchTerm(e.target.value)}
			onKeyDown={(e) => { if (e.key === 'Enter') void carregarClientes(1); }}
            className="w-full pl-11 pr-4 py-2.5 border-2 border-slate-200 rounded-xl focus:border-blue-600 focus:ring-0 bg-white shadow-sm transition-all"
          />
        </div>
        <button className="p-3 text-slate-600 bg-white border-2 border-slate-200 rounded-xl hover:bg-slate-50 hover:border-slate-300 transition-colors shadow-sm">
          <Filter className="w-5 h-5" />
        </button>
      </div>

      {/* Kanban Board */}
      <div className="flex-1 overflow-x-auto pb-4 custom-scrollbar">
        <div className="flex gap-4 min-w-max h-full items-start">
          {STATUS_FUNIL.map(coluna => {
            const style = FUNIL_COLORS[coluna.value] || FUNIL_COLORS['NOVO'];
            const clientesColuna = filteredClientes.filter(c => c.status_funil === coluna.value);
            const Icon = style.icon;

            return (
              <div 
                key={coluna.value} 
                className={`w-[320px] flex flex-col bg-slate-50/50 rounded-2xl border border-slate-200 shadow-sm h-full max-h-full transition-colors`}
                onDragOver={handleDragOver}
                onDrop={(e) => handleDrop(e, coluna.value)}
              >
                {/* Column Header */}
                <div className={`p-4 ${style.headerBg} border-b ${style.border} rounded-t-2xl flex justify-between items-center backdrop-blur-sm`}>
                  <div className="flex items-center gap-2">
                    <Icon className={`w-5 h-5 ${style.text}`} />
                    <h3 className={`font-bold text-sm uppercase tracking-wider ${style.text}`}>{coluna.label}</h3>
                  </div>
                  <span className={`bg-white/80 ${style.text} text-xs font-bold px-2.5 py-1 rounded-full shadow-sm`}>
                    {clientesColuna.length}
                  </span>
                </div>
                
                {/* Column Body / Cards */}
                <div className="p-3 flex-1 overflow-y-auto space-y-3 custom-scrollbar">
                  {loading ? (
                    <div className="animate-pulse space-y-3">
                      {[1,2,3].map(i => (
                        <div key={i} className="bg-white/60 h-28 rounded-xl border border-white/20"></div>
                      ))}
                    </div>
                  ) : clientesColuna.map(cliente => (
                    <div 
                      key={cliente.id}
                      draggable
                      onDragStart={(e) => handleDragStart(e, cliente.id!)}
                      onClick={() => navigate(`/app/crm/${cliente.id}`)}
                      className={`group ${style.bg} p-4 rounded-xl border ${style.border} shadow-sm hover:shadow-md cursor-grab active:cursor-grabbing hover:border-slate-300 transition-all flex flex-col gap-3 relative overflow-hidden`}
                    >
                      {/* Left Accent Bar */}
                      <div className={`absolute left-0 top-0 bottom-0 w-1 ${style.bg.replace('50', '400')} opacity-0 group-hover:opacity-100 transition-opacity`} />
                      
                      <div className="flex justify-between items-start">
                        <div className="flex items-center gap-3">
                          <div className={`w-10 h-10 rounded-full flex items-center justify-center font-bold text-sm bg-white/60 ${style.text} border ${style.border}`}>
                            {getInitials(cliente.nome)}
                          </div>
                          <div>
                            <h4 className={`font-bold text-sm line-clamp-1 text-slate-800`}>{cliente.nome}</h4>
                            {cliente.interesse_tipo ? (
                              <p className={`text-[11px] font-semibold ${style.text} uppercase tracking-wider mt-0.5 opacity-80`}>
                                Interesse: {cliente.interesse_tipo}
                              </p>
                            ) : (
                              <p className={`text-[11px] font-semibold ${style.text} uppercase tracking-wider mt-0.5 opacity-80`}>
                                Lead
                              </p>
                            )}
                          </div>
                        </div>
                        <button className="text-slate-400 hover:text-slate-700 transition-colors p-1 -mr-1">
                          <MoreVertical className="w-4 h-4" />
                        </button>
                      </div>
                      
                      <div className="flex flex-col gap-1.5 mt-1">
                        {cliente.telefone && (
                          <div className="flex items-center text-xs text-slate-600 gap-2">
                            <Phone className="w-3.5 h-3.5 text-slate-400" />
                            <span className="font-medium">{cliente.telefone}</span>
                          </div>
                        )}
                        {cliente.email && (
                          <div className="flex items-center text-xs text-slate-600 gap-2">
                            <Mail className="w-3.5 h-3.5 text-slate-400" />
                            <span className="truncate">{cliente.email}</span>
                          </div>
                        )}
                      </div>
                      
                      {/* Renderizar Tags */}
                      {cliente.tags && cliente.tags.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-1">
                          {cliente.tags.slice(0, 3).map((tag, idx) => (
                            <span key={idx} className="px-1.5 py-0.5 text-[10px] bg-blue-50 text-blue-600 rounded font-medium border border-blue-100">
                              #{tag}
                            </span>
                          ))}
                          {cliente.tags.length > 3 && (
                            <span className="px-1.5 py-0.5 text-[10px] bg-slate-50 text-slate-500 rounded font-medium border border-slate-200">
                              +{cliente.tags.length - 3}
                            </span>
                          )}
                        </div>
                      )}
                      
                      {cliente.origem && (
                        <div className="pt-3 mt-1 border-t border-slate-100 flex items-center justify-between">
                          <span className="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-semibold bg-slate-100 text-slate-500 border border-slate-200">
                            {cliente.origem}
                          </span>
                        </div>
                      )}
                    </div>
                  ))}
                  
                  {!loading && clientesColuna.length === 0 && (
                    <div className="h-24 flex items-center justify-center border-2 border-dashed border-slate-300/50 rounded-xl text-sm font-medium text-slate-400/80 bg-white/30 backdrop-blur-sm">
                      Mova um card para cá
                    </div>
                  )}
                </div>
              </div>
            )
          })}
        </div>
      </div>
	  {clientes.length < total && <div className="mt-4 text-center"><button type="button" onClick={() => void carregarClientes(pagina + 1, true)} disabled={loading} className="rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-50 disabled:opacity-60">Carregar mais oportunidades ({clientes.length} de {total})</button></div>}
      
      <style>{`
        .custom-scrollbar::-webkit-scrollbar {
          width: 6px;
          height: 6px;
        }
        .custom-scrollbar::-webkit-scrollbar-track {
          background: transparent;
        }
        .custom-scrollbar::-webkit-scrollbar-thumb {
          background-color: #cbd5e1;
          border-radius: 20px;
        }
        .custom-scrollbar::-webkit-scrollbar-thumb:hover {
          background-color: #94a3b8;
        }
      `}</style>
    </div>
  );
};
