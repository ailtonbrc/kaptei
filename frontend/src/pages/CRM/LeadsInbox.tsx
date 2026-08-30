import React, { useState } from 'react';
import { 
  Inbox, Check, X, Search, Loader2, Clock, 
  MapPin, Phone, Mail, MessageSquare 
} from 'lucide-react';
import { leadsService } from '../../services/leadsService';
import type { Lead } from '../../types/lead';
import { formatDistanceToNow } from 'date-fns';
import { ptBR } from 'date-fns/locale';
import { usePaginacao } from '../../hooks/usePaginacao';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '../../components/ui/dialog';
import { Button } from '../../components/ui/button';
import { Label } from '../../components/ui/label';
import { Input } from '../../components/ui/input';
import { toast } from 'sonner';

export const LeadsInbox: React.FC = () => {
  const {
    items: leads,
    setItems: setLeads,
    loading,
    pagina,
    total,
    searchTerm,
    setSearchTerm,
    filter: filterStatus,
    setFilter: setFilterStatus,
    carregarItems: carregarLeads,
    removeItem,
  } = usePaginacao<Lead>({
    fetchData: async (pag, busca, filtro) => {
      const statusEnvio = filtro === 'TODOS' ? '' : filtro;
      return await leadsService.listar({ pagina: pag, busca, status: statusEnvio });
    },
    initialFilter: 'NOVO'
  });

  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [leadParaDescartar, setLeadParaDescartar] = useState<Lead | null>(null);
  const [motivoDescarte, setMotivoDescarte] = useState('');

  const handleQualificar = async (lead: Lead) => {
    if (!lead.id) return;
    try {
      setActionLoading(lead.id);
      await leadsService.qualificar(lead.id);
      removeItem(lead.id);
      toast.success('Lead qualificado com sucesso!');
    } catch (error) {
      toast.error('Não foi possível qualificar o lead.');
      console.error('Erro ao qualificar', error);
    } finally {
      setActionLoading(null);
    }
  };

  const abrirModalDescarte = (lead: Lead) => {
    setLeadParaDescartar(lead);
    setMotivoDescarte('');
  };

  const handleDescartarConfirmado = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!leadParaDescartar?.id || !motivoDescarte.trim()) return;
    
    try {
      setActionLoading(leadParaDescartar.id);
      await leadsService.descartar(leadParaDescartar.id, motivoDescarte);
      removeItem(leadParaDescartar.id);
      setLeadParaDescartar(null);
      toast.success('Lead descartado com sucesso.');
    } catch (error) {
      toast.error('Não foi possível descartar o lead.');
      console.error('Erro ao descartar', error);
    } finally {
      setActionLoading(null);
    }
  };

  const filteredLeads = leads.filter(l => {
    const matchStatus = l.status === filterStatus || filterStatus === 'TODOS';
    const matchSearch = l.nome.toLowerCase().includes(searchTerm.toLowerCase()) || 
                        (l.email && l.email.toLowerCase().includes(searchTerm.toLowerCase()));
    return matchStatus && matchSearch;
  });

  return (
    <div className="h-full flex flex-col p-4 sm:p-6 overflow-hidden bg-slate-50/50">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
        <div>
          <h1 className="text-2xl font-extrabold text-slate-900 tracking-tight flex items-center gap-2">
            <Inbox className="w-7 h-7 text-blue-600" /> Caixa de Entrada (Leads)
          </h1>
          <p className="text-sm text-slate-500 mt-1">
            Faça a triagem, descarte spam ou qualifique potenciais clientes para o CRM.
          </p>
        </div>
      </div>

      <div className="flex flex-col sm:flex-row items-center gap-4 mb-6">
        <div className="relative flex-1 w-full max-w-md group">
          <Search className="absolute left-3.5 top-2.5 text-slate-400 w-5 h-5 group-focus-within:text-blue-600 transition-colors" />
          <input 
            type="text" 
            placeholder="Buscar leads por nome, e-mail..." 
            value={searchTerm}
			onChange={(e) => setSearchTerm(e.target.value)}
			onKeyDown={(e) => { if (e.key === 'Enter') void carregarLeads(1); }}
            className="w-full pl-11 pr-4 py-2.5 border border-slate-200 rounded-xl focus:border-blue-600 focus:ring-0 bg-white shadow-sm transition-all"
          />
        </div>
        
        <div className="flex items-center gap-2 w-full sm:w-auto overflow-x-auto pb-2 sm:pb-0">
          {['NOVO', 'EM_ATENDIMENTO', 'TODOS'].map(status => (
            <button
              key={status}
			  onClick={() => { setFilterStatus(status); void carregarLeads(1, false, status); }}
              className={`px-4 py-2 rounded-lg font-medium text-sm transition-all whitespace-nowrap ${
                filterStatus === status 
                  ? 'bg-blue-100 text-blue-700 border-2 border-blue-200' 
                  : 'bg-white text-slate-600 border-2 border-slate-200 hover:bg-slate-50'
              }`}
            >
              {status.replace('_', ' ')}
            </button>
          ))}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto custom-scrollbar pr-2">
        {loading ? (
          <div className="flex justify-center items-center h-40">
            <Loader2 className="w-8 h-8 animate-spin text-blue-600" />
          </div>
        ) : filteredLeads.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-slate-500 bg-white rounded-2xl border border-slate-200 shadow-sm p-12">
            <Inbox className="w-16 h-16 text-slate-300 mb-4" />
            <h3 className="text-lg font-semibold text-slate-700">Nenhum lead encontrado</h3>
            <p className="text-sm text-slate-400 mt-1">Sua caixa de entrada está limpa!</p>
          </div>
        ) : (
          <div className="space-y-4">
            {filteredLeads.map(lead => (
              <div 
                key={lead.id} 
                className="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm hover:shadow-md transition-shadow relative overflow-hidden group flex flex-col md:flex-row md:items-center justify-between gap-4"
              >
                <div className="absolute left-0 top-0 bottom-0 w-1 bg-blue-500 opacity-0 group-hover:opacity-100 transition-opacity" />
                
                <div className="flex-1">
                  <div className="flex items-center gap-3 mb-2">
                    <span className={`px-2.5 py-1 text-[10px] font-bold uppercase rounded-md tracking-wider ${
                      lead.status === 'NOVO' ? 'bg-sky-100 text-sky-700' : 'bg-amber-100 text-amber-700'
                    }`}>
                      {lead.status.replace('_', ' ')}
                    </span>
                    <span className="flex items-center gap-1.5 text-xs font-medium text-slate-400">
                      <Clock className="w-3.5 h-3.5" />
                      {lead.criado_em && formatDistanceToNow(new Date(lead.criado_em), { addSuffix: true, locale: ptBR })}
                    </span>
                    {lead.origem && (
                      <span className="flex items-center gap-1.5 text-xs font-medium text-slate-400 bg-slate-100 px-2 py-0.5 rounded-full">
                        <MapPin className="w-3 h-3" />
                        {lead.origem}
                      </span>
                    )}
                  </div>
                  
                  <h3 className="text-lg font-bold text-slate-900 mb-2">{lead.nome}</h3>
                  
                  <div className="flex flex-wrap items-center gap-x-6 gap-y-2 mb-3">
                    {lead.telefone && (
                      <div className="flex items-center gap-2 text-sm text-slate-600">
                        <Phone className="w-4 h-4 text-slate-400" />
                        {lead.telefone}
                      </div>
                    )}
                    {lead.email && (
                      <div className="flex items-center gap-2 text-sm text-slate-600">
                        <Mail className="w-4 h-4 text-slate-400" />
                        {lead.email}
                      </div>
                    )}
                  </div>
                  
                  {lead.mensagem && (
                    <div className="mt-3 p-3 bg-slate-50 rounded-xl border border-slate-100 text-sm text-slate-600 flex items-start gap-2">
                      <MessageSquare className="w-4 h-4 text-slate-400 mt-0.5 shrink-0" />
                      <p className="line-clamp-2 italic">"{lead.mensagem}"</p>
                    </div>
                  )}
                </div>
                
                <div className="flex md:flex-col gap-2 md:w-48 shrink-0">
                  <button
                    onClick={() => handleQualificar(lead)}
                    disabled={actionLoading === lead.id}
                    className="flex-1 md:flex-none flex items-center justify-center gap-2 bg-emerald-50 text-emerald-700 border border-emerald-200 hover:bg-emerald-100 px-4 py-2.5 rounded-xl font-medium transition-colors"
                  >
                    {actionLoading === lead.id ? <Loader2 className="w-5 h-5 animate-spin" /> : <Check className="w-5 h-5" />}
                    Qualificar
                  </button>
                  <button
                    onClick={() => abrirModalDescarte(lead)}
                    disabled={actionLoading === lead.id}
                    className="flex-1 md:flex-none flex items-center justify-center gap-2 bg-rose-50 text-rose-700 border border-rose-200 hover:bg-rose-100 px-4 py-2.5 rounded-xl font-medium transition-colors"
                  >
                    <X className="w-5 h-5" />
                    Descartar
                  </button>
                </div>
                
              </div>
            ))}
          </div>
        )}
      </div>
	  {leads.length < total && <div className="pt-4 text-center"><button type="button" onClick={() => void carregarLeads(pagina + 1, true)} disabled={loading} className="rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-50 disabled:opacity-60">Carregar mais leads ({leads.length} de {total})</button></div>}

      <style>{`
        .custom-scrollbar::-webkit-scrollbar { width: 6px; }
        .custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
        .custom-scrollbar::-webkit-scrollbar-thumb { background-color: #cbd5e1; border-radius: 20px; }
        .custom-scrollbar::-webkit-scrollbar-thumb:hover { background-color: #94a3b8; }
      `}</style>

      <Dialog open={!!leadParaDescartar} onOpenChange={(open) => !open && setLeadParaDescartar(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Descartar Lead</DialogTitle>
            <DialogDescription>
              Tem certeza que deseja descartar o lead <strong>{leadParaDescartar?.nome}</strong>? Essa ação removerá o lead da caixa de entrada.
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleDescartarConfirmado}>
            <div className="space-y-4 py-4">
              <div className="space-y-2">
                <Label htmlFor="motivo">Motivo do descarte *</Label>
                <Input
                  id="motivo"
                  placeholder="Ex: Não tem interesse no momento, número incorreto..."
                  value={motivoDescarte}
                  onChange={(e) => setMotivoDescarte(e.target.value)}
                  autoFocus
                  required
                />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setLeadParaDescartar(null)}>
                Cancelar
              </Button>
              <Button type="submit" variant="destructive" disabled={!motivoDescarte.trim() || !!actionLoading}>
                {actionLoading === leadParaDescartar?.id ? <Loader2 className="w-4 h-4 mr-2 animate-spin" /> : null}
                Confirmar Descarte
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
};
