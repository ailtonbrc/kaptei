import { useState, type FormEvent } from 'react';
import { CheckCircle2, Copy, Link as LinkIcon, Loader2, RotateCcw, Save, ShieldCheck, Zap } from 'lucide-react';
import { api } from '../../../services/api';
import type { ConfiguracaoContaDados } from '../hooks/useConfiguracoesSettings';

export function ConfiguracaoMotorLeads({ inicial }: { inicial: ConfiguracaoContaDados }) {
  const [estrategia, setEstrategia] = useState(inicial.lead_estrategia || 'CAIXA_ENTRADA');
  const [token, setToken] = useState(inicial.lead_token_integracao || '');
  const [prefixo, setPrefixo] = useState(inicial.lead_token_prefixo || '');
  const [salvando, setSalvando] = useState(false);
  const [rotacionando, setRotacionando] = useState(false);
  const [copiado, setCopiado] = useState(false);
  const [mensagem, setMensagem] = useState('');

  async function salvar(evento: FormEvent) {
    evento.preventDefault(); setSalvando(true); setMensagem('');
    try { await api.put('/v1/conta/leads', { lead_estrategia: estrategia }); setMensagem('Estratégia de distribuição atualizada.'); }
    catch { setMensagem('Não foi possível atualizar a estratégia.'); }
    finally { setSalvando(false); }
  }

  async function rotacionar() {
    if (!window.confirm('O token atual deixará de funcionar imediatamente. Deseja continuar?')) return;
    setRotacionando(true); setMensagem('');
    try {
      const resposta = await api.post<{ token: string; prefixo: string }>('/v1/conta/leads/token');
      setToken(resposta.data.token); setPrefixo(resposta.data.prefixo);
      setMensagem('Copie o novo token agora. O valor completo não será exibido novamente.');
    } catch { setMensagem('Não foi possível rotacionar o token.'); }
    finally { setRotacionando(false); }
  }

  async function copiar() {
    if (!token) return;
    await navigator.clipboard.writeText(token); setCopiado(true);
    window.setTimeout(() => setCopiado(false), 2000);
  }

  return <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
    <header className="flex items-center gap-3 border-b border-slate-100 px-6 py-5"><span className="rounded-lg bg-blue-50 p-2 text-blue-600"><Zap className="h-5 w-5" /></span><div><h2 className="font-semibold text-slate-900">Motor de leads</h2><p className="text-sm text-slate-500">Distribuição e entrada segura para portais e campanhas.</p></div></header>
    <form onSubmit={salvar} className="space-y-6 p-6"><div className="grid gap-6 md:grid-cols-2">
      <div><label className="text-sm font-medium text-slate-700">Estratégia de distribuição<select value={estrategia} onChange={(e) => setEstrategia(e.target.value)} className="mt-1 w-full rounded-lg border border-slate-300 bg-white px-4 py-2"><option value="CAIXA_ENTRADA">Caixa de entrada (manual)</option><option value="ROLETA">Roleta / round-robin</option></select></label><p className="mt-2 text-xs text-slate-500">Na roleta, o backend seleciona de forma atômica o próximo corretor ativo.</p></div>
      <div><div className="mb-2 flex items-center gap-2 text-sm font-medium text-slate-700"><ShieldCheck className="h-4 w-4 text-emerald-600" />Token seguro</div><div className="relative"><LinkIcon className="absolute left-3 top-2.5 h-5 w-5 text-slate-400" /><input readOnly value={token || (prefixo ? `${prefixo}… (protegido)` : 'Gere um token para integrar')} className="w-full rounded-lg border border-slate-200 bg-slate-50 py-2 pl-10 pr-12 text-slate-600" /><button type="button" onClick={() => void copiar()} disabled={!token} aria-label="Copiar token" className="absolute inset-y-0 right-0 px-3 text-blue-600 disabled:text-slate-300">{copiado ? <CheckCircle2 className="h-5 w-5 text-emerald-600" /> : <Copy className="h-5 w-5" />}</button></div><p className="mt-2 text-xs text-slate-500">POST <code>/api/webhooks/leads</code> com o cabeçalho <code>X-Kaptei-Token</code>.</p><button type="button" onClick={() => void rotacionar()} disabled={rotacionando} className="mt-3 inline-flex items-center gap-2 rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 disabled:opacity-60">{rotacionando ? <Loader2 className="h-4 w-4 animate-spin" /> : <RotateCcw className="h-4 w-4" />}{prefixo || token ? 'Rotacionar token' : 'Gerar token seguro'}</button></div>
    </div>{mensagem && <p className="text-sm text-slate-600">{mensagem}</p>}<div className="flex justify-end border-t border-slate-100 pt-5"><button disabled={salvando} className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-5 py-2.5 font-medium text-white disabled:opacity-60">{salvando ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}Salvar estratégia</button></div></form>
  </section>;
}
