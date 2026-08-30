import { useCallback, useEffect, useState } from 'react';
import { LockKeyhole, PlayCircle, RefreshCw, Save, ShieldAlert, Trash2 } from 'lucide-react';
import { Button } from '../../../components/ui/button';
import { privacidadeService } from '../../../services/privacidadeService';
import type {
  BloqueioRetencao,
  PoliticaRetencao,
  RelatorioRetencao,
  ResultadoRetencao,
} from '../../../types/privacidade';

const CONFIRMACAO_EXECUCAO = 'ANONIMIZAR DADOS EXPIRADOS';

const politicaInicial: PoliticaRetencao = {
  ativa: false,
  dias_leads_descartados: 730,
  dias_clientes_perdidos: 1825,
  tamanho_lote: 200,
  fundamento_legal: '',
};

const mensagemErro = (erro: unknown, alternativa: string) => {
  if (typeof erro === 'object' && erro !== null && 'response' in erro) {
    const resposta = (erro as { response?: { data?: { erro?: string } } }).response;
    if (resposta?.data?.erro) return resposta.data.erro;
  }
  return alternativa;
};

export const PainelRetencao = () => {
  const [politica, setPolitica] = useState<PoliticaRetencao>(politicaInicial);
  const [relatorio, setRelatorio] = useState<RelatorioRetencao>();
  const [bloqueios, setBloqueios] = useState<BloqueioRetencao[]>([]);
  const [confirmacao, setConfirmacao] = useState('');
  const [tipoRecurso, setTipoRecurso] = useState<'LEAD' | 'CLIENTE'>('LEAD');
  const [recursoId, setRecursoId] = useState('');
  const [motivo, setMotivo] = useState('');
  const [validoAte, setValidoAte] = useState('');
  const [carregando, setCarregando] = useState(true);
  const [processando, setProcessando] = useState(false);
  const [erro, setErro] = useState('');
  const [sucesso, setSucesso] = useState('');

  const carregar = useCallback(async () => {
    setCarregando(true);
    setErro('');
    try {
      const [politicaAtual, relatorioAtual, bloqueiosAtuais] = await Promise.all([
        privacidadeService.obterPoliticaRetencao(),
        privacidadeService.obterRelatorioRetencao(),
        privacidadeService.listarBloqueiosRetencao(),
      ]);
      setPolitica(politicaAtual);
      setRelatorio(relatorioAtual);
      setBloqueios(bloqueiosAtuais);
    } catch (falha) {
      setErro(mensagemErro(falha, 'Não foi possível carregar a política de retenção.'));
    } finally {
      setCarregando(false);
    }
  }, []);

  useEffect(() => {
    let ativo = true;
    Promise.all([
      privacidadeService.obterPoliticaRetencao(),
      privacidadeService.obterRelatorioRetencao(),
      privacidadeService.listarBloqueiosRetencao(),
    ])
      .then(([politicaAtual, relatorioAtual, bloqueiosAtuais]) => {
        if (!ativo) return;
        setPolitica(politicaAtual);
        setRelatorio(relatorioAtual);
        setBloqueios(bloqueiosAtuais);
      })
      .catch((falha) => {
        if (ativo) setErro(mensagemErro(falha, 'Não foi possível carregar a política de retenção.'));
      })
      .finally(() => { if (ativo) setCarregando(false); });
    return () => { ativo = false; };
  }, []);

  const salvarPolitica = async () => {
    setProcessando(true); setErro(''); setSucesso('');
    try {
      await privacidadeService.salvarPoliticaRetencao(politica);
      setSucesso('Política de retenção salva.');
      await carregar();
    } catch (falha) {
      setErro(mensagemErro(falha, 'Não foi possível salvar a política.'));
    } finally { setProcessando(false); }
  };

  const executar = async () => {
    setProcessando(true); setErro(''); setSucesso('');
    try {
      const resultado: ResultadoRetencao = await privacidadeService.executarRetencao(confirmacao);
      setConfirmacao('');
      setSucesso(`${resultado.leads_anonimizados} lead(s) e ${resultado.clientes_anonimizados} cliente(s) anonimizados.`);
      await carregar();
    } catch (falha) {
      setErro(mensagemErro(falha, 'Não foi possível executar a retenção.'));
    } finally { setProcessando(false); }
  };

  const adicionarBloqueio = async () => {
    setProcessando(true); setErro(''); setSucesso('');
    try {
      await privacidadeService.salvarBloqueioRetencao({
        tipo_recurso: tipoRecurso,
        recurso_id: recursoId.trim(),
        motivo: motivo.trim(),
        valido_ate: validoAte ? new Date(validoAte).toISOString() : undefined,
      });
      setRecursoId(''); setMotivo(''); setValidoAte('');
      setSucesso('Bloqueio legal registrado.');
      await carregar();
    } catch (falha) {
      setErro(mensagemErro(falha, 'Não foi possível registrar o bloqueio.'));
    } finally { setProcessando(false); }
  };

  const removerBloqueio = async (id: string) => {
    setProcessando(true); setErro(''); setSucesso('');
    try {
      await privacidadeService.removerBloqueioRetencao(id);
      setSucesso('Bloqueio removido.');
      await carregar();
    } catch (falha) {
      setErro(mensagemErro(falha, 'Não foi possível remover o bloqueio.'));
    } finally { setProcessando(false); }
  };

  const atualizarNumero = (campo: 'dias_leads_descartados' | 'dias_clientes_perdidos' | 'tamanho_lote', valor: string) => {
    setPolitica((atual) => ({ ...atual, [campo]: Number(valor) }));
  };

  return <section className="mt-8 space-y-6" aria-labelledby="titulo-retencao">
    <div className="flex flex-wrap items-center justify-between gap-3">
      <div><p className="flex items-center gap-2 text-sm font-bold uppercase tracking-widest text-amber-700"><ShieldAlert className="h-4 w-4" /> Governança de dados</p><h2 id="titulo-retencao" className="mt-1 text-2xl font-black text-slate-950">Retenção e anonimização</h2><p className="mt-1 max-w-3xl text-sm text-slate-600">Somente leads descartados e clientes perdidos fora do prazo são elegíveis. Bloqueios legais e atendimentos futuros impedem o processamento.</p></div>
      <Button variant="outline" onClick={() => void carregar()} disabled={carregando || processando}><RefreshCw className={`h-4 w-4 ${carregando ? 'animate-spin' : ''}`} /> Recalcular</Button>
    </div>

    {erro && <p role="alert" className="rounded-xl bg-red-50 p-4 text-sm font-medium text-red-700">{erro}</p>}
    {sucesso && <p role="status" className="rounded-xl bg-emerald-50 p-4 text-sm font-medium text-emerald-800">{sucesso}</p>}

    <div className="grid gap-4 sm:grid-cols-3">
      {[['Leads elegíveis', relatorio?.leads_elegiveis ?? 0], ['Clientes elegíveis', relatorio?.clientes_elegiveis ?? 0], ['Bloqueios vigentes', relatorio?.bloqueios_vigentes ?? 0]].map(([rotulo, valor]) => <div key={String(rotulo)} className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm"><p className="text-sm font-semibold text-slate-600">{rotulo}</p><p className="mt-2 text-3xl font-black text-slate-950">{valor}</p></div>)}
    </div>

    <div className="grid gap-6 xl:grid-cols-2">
      <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
        <div className="flex items-center justify-between gap-4"><div><h3 className="font-black text-slate-950">Política da conta</h3><p className="text-sm text-slate-500">Desativada por padrão.</p></div><label className="flex items-center gap-2 text-sm font-bold text-slate-700"><input type="checkbox" checked={politica.ativa} onChange={(e) => setPolitica((atual) => ({ ...atual, ativa: e.target.checked }))} /> Ativa</label></div>
        <div className="mt-5 grid gap-4 sm:grid-cols-3">
          <label className="text-sm font-semibold text-slate-700">Dias para leads<input type="number" min={30} max={3650} className="mt-1 w-full rounded-xl border border-slate-300 px-3 py-2" value={politica.dias_leads_descartados} onChange={(e) => atualizarNumero('dias_leads_descartados', e.target.value)} /></label>
          <label className="text-sm font-semibold text-slate-700">Dias para clientes<input type="number" min={30} max={3650} className="mt-1 w-full rounded-xl border border-slate-300 px-3 py-2" value={politica.dias_clientes_perdidos} onChange={(e) => atualizarNumero('dias_clientes_perdidos', e.target.value)} /></label>
          <label className="text-sm font-semibold text-slate-700">Lote por tipo<input type="number" min={1} max={1000} className="mt-1 w-full rounded-xl border border-slate-300 px-3 py-2" value={politica.tamanho_lote} onChange={(e) => atualizarNumero('tamanho_lote', e.target.value)} /></label>
        </div>
        <label className="mt-4 block text-sm font-semibold text-slate-700">Fundamento legal<textarea rows={4} maxLength={2000} className="mt-1 w-full rounded-xl border border-slate-300 px-3 py-2" placeholder="Finalidade, obrigação legal e critérios aprovados pelo controlador" value={politica.fundamento_legal} onChange={(e) => setPolitica((atual) => ({ ...atual, fundamento_legal: e.target.value }))} /></label>
        <Button className="mt-4" onClick={() => void salvarPolitica()} disabled={processando}><Save className="h-4 w-4" /> Salvar política</Button>
      </div>

      <div className="rounded-2xl border border-red-200 bg-white p-5 shadow-sm">
        <h3 className="font-black text-slate-950">Execução controlada</h3><p className="mt-1 text-sm text-slate-600">A operação anonimiza dados pessoais e registra o lote em auditoria. Ela não pode ser desfeita pelo sistema.</p>
        <label className="mt-5 block text-sm font-semibold text-slate-700">Digite <span className="font-mono text-red-700">{CONFIRMACAO_EXECUCAO}</span><input className="mt-2 w-full rounded-xl border border-red-300 px-3 py-2 font-mono text-sm" value={confirmacao} onChange={(e) => setConfirmacao(e.target.value)} /></label>
        <Button variant="destructive" className="mt-4" onClick={() => void executar()} disabled={processando || !politica.ativa || confirmacao !== CONFIRMACAO_EXECUCAO}><PlayCircle className="h-4 w-4" /> Anonimizar lote elegível</Button>
      </div>
    </div>

    <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
      <div><h3 className="flex items-center gap-2 font-black text-slate-950"><LockKeyhole className="h-4 w-4" /> Bloqueios legais</h3><p className="mt-1 text-sm text-slate-600">Use para litígios, auditorias, obrigações regulatórias ou investigações que exijam preservação.</p></div>
      <div className="mt-5 grid gap-3 lg:grid-cols-[140px_1fr_1fr_220px_auto]">
        <select aria-label="Tipo do recurso" className="rounded-xl border border-slate-300 px-3 py-2 text-sm" value={tipoRecurso} onChange={(e) => setTipoRecurso(e.target.value as 'LEAD' | 'CLIENTE')}><option value="LEAD">Lead</option><option value="CLIENTE">Cliente</option></select>
        <input aria-label="Identificador do recurso" className="rounded-xl border border-slate-300 px-3 py-2 text-sm" placeholder="UUID do recurso" value={recursoId} onChange={(e) => setRecursoId(e.target.value)} />
        <input aria-label="Motivo do bloqueio" className="rounded-xl border border-slate-300 px-3 py-2 text-sm" placeholder="Motivo documentado" value={motivo} onChange={(e) => setMotivo(e.target.value)} />
        <input aria-label="Válido até" type="datetime-local" className="rounded-xl border border-slate-300 px-3 py-2 text-sm" value={validoAte} onChange={(e) => setValidoAte(e.target.value)} />
        <Button onClick={() => void adicionarBloqueio()} disabled={processando || !recursoId.trim() || motivo.trim().length < 5}>Adicionar</Button>
      </div>
      <div className="mt-5 overflow-x-auto"><table className="w-full text-left text-sm"><thead className="bg-slate-50 text-xs uppercase text-slate-500"><tr><th className="px-3 py-2">Tipo</th><th className="px-3 py-2">Recurso</th><th className="px-3 py-2">Motivo</th><th className="px-3 py-2">Vigência</th><th className="px-3 py-2"><span className="sr-only">Ações</span></th></tr></thead><tbody>{bloqueios.map((bloqueio) => <tr key={bloqueio.id} className="border-t border-slate-100"><td className="px-3 py-3 font-bold">{bloqueio.tipo_recurso}</td><td className="px-3 py-3 font-mono text-xs">{bloqueio.recurso_id}</td><td className="px-3 py-3">{bloqueio.motivo}</td><td className="px-3 py-3">{bloqueio.valido_ate ? new Date(bloqueio.valido_ate).toLocaleString('pt-BR') : 'Sem prazo'}</td><td className="px-3 py-3 text-right"><Button variant="ghost" size="icon" aria-label="Remover bloqueio" onClick={() => void removerBloqueio(bloqueio.id)} disabled={processando}><Trash2 className="h-4 w-4 text-red-600" /></Button></td></tr>)}</tbody></table>{!bloqueios.length && <p className="p-6 text-center text-sm text-slate-500">Nenhum bloqueio registrado.</p>}</div>
    </div>
  </section>;
};
