import { useCallback, useEffect, useState } from 'react';
import { RefreshCw, Scale } from 'lucide-react';
import { Button } from '../../components/ui/button';
import { privacidadeService } from '../../services/privacidadeService';
import type { SolicitacaoTitular } from '../../types/privacidade';
import { PainelSolicitacao } from './components/PainelSolicitacao';
import { PainelRetencao } from './components/PainelRetencao';

export const CentralPrivacidade = () => {
  const [solicitacoes, setSolicitacoes] = useState<SolicitacaoTitular[]>([]);
  const [selecionada, setSelecionada] = useState<SolicitacaoTitular>();
  const [status, setStatus] = useState('');
  const [busca, setBusca] = useState('');
  const [carregando, setCarregando] = useState(true);
  const [erro, setErro] = useState('');

  const carregar = useCallback(async () => {
    setCarregando(true); setErro('');
    try {
      const pagina = await privacidadeService.listar({ pagina: 1, limite: 100, status, busca });
      setSolicitacoes(pagina.dados);
    } catch { setErro('Não foi possível carregar as solicitações de privacidade.'); }
    finally { setCarregando(false); }
  }, [busca, status]);

  useEffect(() => {
    let ativo = true;
    privacidadeService.listar({ pagina: 1, limite: 100, status, busca })
      .then((pagina) => { if (ativo) setSolicitacoes(pagina.dados); })
      .catch(() => { if (ativo) setErro('Não foi possível carregar as solicitações de privacidade.'); })
      .finally(() => { if (ativo) setCarregando(false); });
    return () => { ativo = false; };
  }, [busca, status]);
  const abrir = async (id: string) => {
    setErro('');
    try { setSelecionada(await privacidadeService.obter(id)); } catch { setErro('Não foi possível abrir a solicitação.'); }
  };
  const atualizarSelecionada = async () => {
    if (!selecionada) return;
    const atualizada = await privacidadeService.obter(selecionada.id);
    setSelecionada(atualizada);
    const pagina = await privacidadeService.listar({ pagina: 1, limite: 100, status, busca });
    setSolicitacoes(pagina.dados);
  };

  return <div className="h-full overflow-auto bg-slate-50 p-4 sm:p-6 lg:p-8">
    <div className="mx-auto max-w-7xl">
      <header className="mb-6 flex flex-wrap items-end justify-between gap-4"><div><p className="flex items-center gap-2 text-sm font-bold uppercase tracking-widest text-blue-600"><Scale className="h-4 w-4" /> Governança LGPD</p><h1 className="mt-1 text-3xl font-black tracking-tight text-slate-950">Central de Privacidade</h1><p className="mt-2 max-w-2xl text-sm text-slate-600">Verifique a identidade, documente a base legal e só então cumpra o direito solicitado.</p></div><Button variant="outline" onClick={() => void carregar()} disabled={carregando}><RefreshCw className={`h-4 w-4 ${carregando ? 'animate-spin' : ''}`} /> Atualizar</Button></header>
      <div className="mb-5 grid gap-3 rounded-2xl border border-slate-200 bg-white p-4 sm:grid-cols-[1fr_240px]"><input aria-label="Buscar por protocolo" className="rounded-xl border border-slate-300 px-4 py-2.5 text-sm outline-none focus:border-blue-500" placeholder="Buscar por protocolo" value={busca} onChange={(e) => setBusca(e.target.value)} /><select aria-label="Filtrar por status" className="rounded-xl border border-slate-300 px-4 py-2.5 text-sm outline-none focus:border-blue-500" value={status} onChange={(e) => setStatus(e.target.value)}><option value="">Todos os status</option><option value="RECEBIDA">Recebida</option><option value="EM_ANALISE">Em análise</option><option value="APROVADA">Aprovada</option><option value="REJEITADA">Rejeitada</option><option value="CONCLUIDA">Concluída</option></select></div>
      <PainelRetencao />
      {erro && <p role="alert" className="mb-4 rounded-xl bg-red-50 p-4 text-sm font-medium text-red-700">{erro}</p>}
      <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_minmax(380px,0.8fr)]"><section className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm"><div className="overflow-x-auto"><table className="w-full text-left text-sm"><thead className="bg-slate-50 text-xs uppercase tracking-wide text-slate-500"><tr><th className="px-4 py-3">Protocolo</th><th className="px-4 py-3">Direito</th><th className="px-4 py-3">Status</th><th className="px-4 py-3">Prazo</th></tr></thead><tbody>{solicitacoes.map((item) => <tr key={item.id} onClick={() => void abrir(item.id)} className={`cursor-pointer border-t border-slate-100 hover:bg-blue-50 ${selecionada?.id === item.id ? 'bg-blue-50' : ''}`}><td className="px-4 py-4 font-mono text-xs font-bold text-blue-700">{item.protocolo}</td><td className="px-4 py-4 font-semibold text-slate-800">{item.tipo}</td><td className="px-4 py-4"><span className="rounded-full bg-slate-100 px-2.5 py-1 text-xs font-bold">{item.status.replaceAll('_', ' ')}</span></td><td className="px-4 py-4 text-slate-600">{new Date(item.prazo_resposta_em).toLocaleDateString('pt-BR')}</td></tr>)}</tbody></table></div>{!carregando && !solicitacoes.length && <p className="p-10 text-center text-slate-500">Nenhuma solicitação encontrada.</p>}</section>{selecionada ? <PainelSolicitacao solicitacao={selecionada} aoAtualizar={atualizarSelecionada} /> : <div className="rounded-2xl border border-dashed border-slate-300 p-10 text-center text-sm text-slate-500">Selecione uma solicitação para analisar.</div>}</div>
    </div>
  </div>;
};
