import React, { useState, useEffect } from 'react';
import { Clock, Phone, Mail, MapPin, FileText, Check, Plus, MessageSquare } from 'lucide-react';
import { clientesService } from '../../services/clientesService';
import type { Interacao } from '../../types/cliente';
import { toast } from 'sonner';

interface TimelineLeadProps {
  clienteId: string;
}

const iconMap: Record<string, React.ReactNode> = {
  'LIGACAO': <Phone className="w-4 h-4 text-emerald-600" />,
  'MENSAGEM': <MessageSquare className="w-4 h-4 text-blue-600" />,
  'EMAIL': <Mail className="w-4 h-4 text-amber-600" />,
  'VISITA': <MapPin className="w-4 h-4 text-rose-600" />,
  'PROPOSTA': <FileText className="w-4 h-4 text-indigo-600" />,
  'ANOTACAO': <FileText className="w-4 h-4 text-slate-600" />,
};

const bgMap: Record<string, string> = {
  'LIGACAO': 'bg-emerald-100',
  'MENSAGEM': 'bg-blue-100',
  'EMAIL': 'bg-amber-100',
  'VISITA': 'bg-rose-100',
  'PROPOSTA': 'bg-indigo-100',
  'ANOTACAO': 'bg-slate-100',
};

export const TimelineLead: React.FC<TimelineLeadProps> = ({ clienteId }) => {
  const [interacoes, setInteracoes] = useState<Interacao[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  const [novaInteracao, setNovaInteracao] = useState<Partial<Interacao>>({
    tipo: 'ANOTACAO',
    descricao: '',
  });

  useEffect(() => {
    carregarInteracoes();
    // eslint-disable-next-line react-hooks/exhaustive-deps -- a consulta depende apenas do cliente da rota.
  }, [clienteId]);

  async function carregarInteracoes() {
    try {
      const data = await clientesService.listarInteracoes(clienteId);
      setInteracoes(data || []);
    } catch (error) {
      console.error('Erro ao carregar timeline:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!novaInteracao.descricao?.trim()) return;

    try {
      setIsSaving(true);
      await clientesService.adicionarInteracao(clienteId, {
        ...novaInteracao,
        cliente_id: clienteId,
      });
      setNovaInteracao({ tipo: 'ANOTACAO', descricao: '' });
      setShowForm(false);
      toast.success('Interação registrada com sucesso!');
      carregarInteracoes();
    } catch (error) {
      toast.error('Não foi possível registrar a interação.');
      console.error('Erro ao salvar interação:', error);
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="bg-slate-50 border border-slate-200 rounded-xl p-5 shadow-sm mt-8">
      <div className="flex items-center justify-between mb-6">
        <h3 className="text-lg font-bold text-slate-800 flex items-center gap-2">
          <Clock className="w-5 h-5 text-emerald-600" />
          Histórico e Timeline
        </h3>
        <button
          type="button"
          onClick={() => setShowForm(!showForm)}
          className="flex items-center gap-1.5 px-3 py-1.5 bg-white border border-slate-200 text-sm font-semibold text-slate-700 rounded-lg hover:bg-slate-100 transition-colors"
        >
          <Plus className="w-4 h-4" /> Registrar Ação
        </button>
      </div>

      {showForm && (
        <form onSubmit={handleSave} className="bg-white border border-emerald-100 p-4 rounded-lg shadow-sm mb-6 animate-in slide-in-from-top-2">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-4">
            <div className="col-span-1">
              <label className="block text-xs font-semibold text-slate-600 mb-1">Tipo de Ação</label>
              <select
                value={novaInteracao.tipo}
                onChange={(e) => setNovaInteracao(prev => ({ ...prev, tipo: e.target.value }))}
                className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/20"
              >
                <option value="ANOTACAO">Anotação Livre</option>
                <option value="LIGACAO">Ligação (Telefone/WhatsApp)</option>
                <option value="MENSAGEM">Mensagem (Texto/Email)</option>
                <option value="VISITA">Visita Realizada</option>
                <option value="PROPOSTA">Envio de Proposta</option>
              </select>
            </div>
            <div className="col-span-3">
              <label className="block text-xs font-semibold text-slate-600 mb-1">Detalhes (O que aconteceu?)</label>
              <input
                type="text"
                required
                value={novaInteracao.descricao}
                onChange={(e) => setNovaInteracao(prev => ({ ...prev, descricao: e.target.value }))}
                placeholder="Ex: Cliente atendeu, pediu para retornar às 14h..."
                className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/20"
              />
            </div>
          </div>
          <div className="flex justify-end gap-2">
            <button
              type="button"
              onClick={() => setShowForm(false)}
              className="px-4 py-2 text-sm font-medium text-slate-600 hover:bg-slate-100 rounded-lg transition-colors"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={isSaving}
              className="flex items-center gap-1.5 px-4 py-2 text-sm font-bold text-white bg-emerald-600 hover:bg-emerald-700 rounded-lg transition-colors disabled:opacity-50"
            >
              <Check className="w-4 h-4" />
              {isSaving ? 'Salvando...' : 'Salvar Registro'}
            </button>
          </div>
        </form>
      )}

      <div className="relative pl-6">
        {/* Linha vertical da timeline */}
        <div className="absolute top-2 bottom-2 left-[11px] w-px bg-slate-200"></div>

        {loading ? (
          <p className="text-sm text-slate-500 py-4">Carregando histórico...</p>
        ) : interacoes.length === 0 ? (
          <p className="text-sm text-slate-500 py-4 italic">Nenhuma interação registrada ainda para este cliente.</p>
        ) : (
          <div className="space-y-6 relative">
            {interacoes.map((item, index) => (
              <div key={item.id || index} className="relative group">
                <div className={`absolute -left-[30px] w-7 h-7 rounded-full border-4 border-slate-50 flex items-center justify-center ${bgMap[item.tipo] || 'bg-slate-100'}`}>
                  {iconMap[item.tipo] || <FileText className="w-3 h-3 text-slate-500" />}
                </div>
                
                <div className="bg-white border border-slate-100 p-4 rounded-xl shadow-sm hover:shadow transition-shadow ml-2">
                  <div className="flex justify-between items-start mb-1">
                    <span className="text-xs font-bold uppercase tracking-wider text-slate-500">
                      {item.tipo}
                    </span>
                    <span className="text-[11px] font-medium text-slate-400 bg-slate-100 px-2 py-0.5 rounded-full">
                      {item.data_hora ? new Date(item.data_hora).toLocaleString('pt-BR') : ''}
                    </span>
                  </div>
                  <p className="text-sm text-slate-700 leading-relaxed mt-2">
                    {item.descricao}
                  </p>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
