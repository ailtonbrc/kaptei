import { useState, type FormEvent } from 'react';
import { Activity, Loader2, Save, ShieldCheck } from 'lucide-react';
import { api } from '../../../services/api';
import type { ConfiguracaoObservabilidadeDados } from '../hooks/useConfiguracoesSettings';

interface Props {
  inicial: ConfiguracaoObservabilidadeDados;
}

export function ConfiguracaoObservabilidade({ inicial }: Props) {
  const [ativa, setAtiva] = useState(inicial.ativa);
  const [token, setToken] = useState('');
  const [tokenConfigurado, setTokenConfigurado] = useState(inicial.token_configurado);
  const [salvando, setSalvando] = useState(false);
  const [mensagem, setMensagem] = useState('');
  const [erro, setErro] = useState('');

  async function salvar(evento: FormEvent) {
    evento.preventDefault();
    setSalvando(true);
    setMensagem('');
    setErro('');
    try {
      await api.put('/v1/configuracoes/OBSERVABILIDADE_CONFIG', {
        valor: { ativa, token: token.trim() },
        descricao: 'Acesso protegido ao endpoint Prometheus',
      });
      if (token.trim()) setTokenConfigurado(true);
      setToken('');
      setMensagem('Configuração de métricas atualizada. A alteração passa a valer em até 30 segundos.');
    } catch (falha) {
      const resposta = falha as { response?: { data?: { erro?: string } } };
      setErro(resposta.response?.data?.erro || 'Não foi possível salvar a configuração de métricas.');
    } finally {
      setSalvando(false);
    }
  }

  return <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
    <header className="flex items-center gap-3 border-b border-slate-100 px-6 py-5">
      <span className="rounded-lg bg-violet-50 p-2 text-violet-600"><Activity className="h-5 w-5" /></span>
      <div><h2 className="font-semibold text-slate-900">Observabilidade</h2><p className="text-sm text-slate-500">Exponha métricas operacionais para coleta pelo Prometheus.</p></div>
    </header>
    <form onSubmit={salvar} className="space-y-5 p-6">
      <label className="block text-sm font-medium text-slate-700">Token Bearer de coleta
        <input type="password" autoComplete="new-password" minLength={32} maxLength={512} value={token} onChange={(e) => setToken(e.target.value)} placeholder={tokenConfigurado ? 'Deixe vazio para manter o token protegido' : 'Informe um token aleatório com no mínimo 32 caracteres'} className="mt-1 w-full rounded-lg border border-slate-300 px-4 py-2.5 outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-100" />
      </label>
      <div className="flex flex-col gap-3 rounded-xl border border-slate-200 bg-slate-50 p-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex gap-3"><ShieldCheck className="mt-0.5 h-5 w-5 shrink-0 text-emerald-600" /><div><p className="text-sm font-medium text-slate-800">Endpoint protegido e desativado por padrão</p><p className="text-xs text-slate-500">Use <code>Authorization: Bearer TOKEN</code> ao coletar <code>/metrics</code>. O token é cifrado e nunca retorna pela API.</p></div></div>
        <label className="flex shrink-0 items-center gap-2 text-sm font-medium text-slate-700"><input type="checkbox" checked={ativa} onChange={(e) => setAtiva(e.target.checked)} className="h-4 w-4 rounded border-slate-300 text-blue-600" />Métricas ativas</label>
      </div>
      {mensagem && <p role="status" className="text-sm text-emerald-700">{mensagem}</p>}
      {erro && <p role="alert" className="text-sm text-red-600">{erro}</p>}
      <div className="flex justify-end border-t border-slate-100 pt-5"><button disabled={salvando || (ativa && !tokenConfigurado && token.trim().length < 32)} className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-5 py-2.5 font-medium text-white disabled:opacity-60">{salvando ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}Salvar observabilidade</button></div>
    </form>
  </section>;
}
