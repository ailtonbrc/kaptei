import { useEffect, useState } from 'react';
import { CreditCard, Loader2, Save } from 'lucide-react';
import { toast } from 'sonner';
import { api } from '../../../services/api';
import { obterMensagemErro } from '../../../lib/http/erro-api';

interface PlanoGateway { codigo: string; nome: string; gateway_price_id: string }

export const ConfiguracaoCobranca = () => {
  const [planos, setPlanos] = useState<PlanoGateway[]>([]);
  const [carregando, setCarregando] = useState(true);
  const [salvando, setSalvando] = useState('');
  useEffect(() => { void api.get<PlanoGateway[]>('/v1/planos/administracao').then((r) => setPlanos(r.data)).finally(() => setCarregando(false)); }, []);
  const salvar = async (plano: PlanoGateway) => {
    setSalvando(plano.codigo);
    try { await api.put(`/v1/planos/${encodeURIComponent(plano.codigo)}/gateway`, { gateway_price_id: plano.gateway_price_id }); toast.success(`Preço do plano ${plano.nome} atualizado.`); }
    catch (erro: unknown) { toast.error(obterMensagemErro(erro, 'Não foi possível atualizar o preço.')); }
    finally { setSalvando(''); }
  };
  return <div className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm"><h2 className="flex items-center gap-2 text-lg font-bold"><CreditCard className="h-5 w-5 text-blue-600" /> Cobrança Stripe</h2><p className="mt-1 text-sm text-slate-500">Associe cada plano ao Price ID recorrente criado no painel Stripe.</p>{carregando ? <Loader2 className="mx-auto mt-6 h-6 w-6 animate-spin" /> : <div className="mt-5 space-y-3">{planos.map((plano) => <div key={plano.codigo} className="grid items-center gap-3 rounded-xl border border-slate-200 p-4 sm:grid-cols-[180px_1fr_auto]"><div><strong className="block text-sm">{plano.nome}</strong><span className="text-xs text-slate-400">{plano.codigo}</span></div><input className="rounded-lg border border-slate-200 px-3 py-2 text-sm" placeholder="price_..." value={plano.gateway_price_id} onChange={(e) => setPlanos((atuais) => atuais.map((item) => item.codigo === plano.codigo ? { ...item, gateway_price_id: e.target.value } : item))} /><button onClick={() => salvar(plano)} disabled={salvando === plano.codigo} className="inline-flex items-center justify-center gap-2 rounded-lg bg-slate-900 px-4 py-2 text-sm font-bold text-white"><Save className="h-4 w-4" /> Salvar</button></div>)}</div>}</div>;
};
