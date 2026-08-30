import { useState, type FormEvent } from 'react';
import { Loader2, Megaphone, Save, ShieldCheck } from 'lucide-react';
import { api } from '../../../services/api';
import type { ConfiguracaoMetaLeadsDados } from '../hooks/useConfiguracoesSettings';

interface Props {
  inicial: ConfiguracaoMetaLeadsDados;
}

export function ConfiguracaoMetaLeads({ inicial }: Props) {
  const [paginaID, setPaginaID] = useState(inicial.pagina_id ?? '');
  const [tokenPagina, setTokenPagina] = useState('');
  const [ativa, setAtiva] = useState(inicial.ativa ?? false);
  const [tokenConfigurado, setTokenConfigurado] = useState(inicial.token_pagina_configurado ?? false);
  const disponivelNoServidor = inicial.disponivel_no_servidor ?? false;
  const [salvando, setSalvando] = useState(false);
  const [mensagem, setMensagem] = useState('');
  const [erro, setErro] = useState('');

  async function salvar(evento: FormEvent) {
    evento.preventDefault();
    setSalvando(true);
    setMensagem('');
    setErro('');
    try {
      const resposta = await api.put<ConfiguracaoMetaLeadsDados>('/v1/integracoes/meta/leads', {
        pagina_id: paginaID.trim(),
        token_pagina: tokenPagina.trim() || undefined,
        ativa,
      });
      setPaginaID(resposta.data.pagina_id);
      setAtiva(resposta.data.ativa);
      setTokenConfigurado(resposta.data.token_pagina_configurado);
      setTokenPagina('');
      setMensagem('IntegraÃ§Ã£o Meta Leads atualizada com seguranÃ§a.');
    } catch {
      setErro('NÃ£o foi possÃ­vel salvar. Confira o ID da pÃ¡gina, o token e se a pÃ¡gina jÃ¡ estÃ¡ vinculada a outra conta.');
    } finally {
      setSalvando(false);
    }
  }

  return <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
    <header className="flex items-center gap-3 border-b border-slate-100 px-6 py-5">
      <span className="rounded-lg bg-blue-50 p-2 text-blue-600"><Megaphone className="h-5 w-5" /></span>
      <div><h2 className="font-semibold text-slate-900">Meta Lead Ads</h2><p className="text-sm text-slate-500">Receba formulÃ¡rios do Facebook e Instagram no motor de leads.</p></div>
    </header>
    <form onSubmit={salvar} className="space-y-5 p-6">
      <div className="grid gap-5 md:grid-cols-2">
        <label className="text-sm font-medium text-slate-700">ID da pÃ¡gina Meta
          <input inputMode="numeric" autoComplete="off" required maxLength={64} value={paginaID} onChange={(e) => setPaginaID(e.target.value.replace(/\D/g, ''))} placeholder="Ex.: 123456789012345" className="mt-1 w-full rounded-lg border border-slate-300 px-4 py-2.5 outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-100" />
        </label>
        <label className="text-sm font-medium text-slate-700">Token de acesso da pÃ¡gina
          <input type="password" autoComplete="new-password" value={tokenPagina} onChange={(e) => setTokenPagina(e.target.value)} placeholder={tokenConfigurado ? 'Deixe vazio para manter o token protegido' : 'Cole o token de acesso da pÃ¡gina'} className="mt-1 w-full rounded-lg border border-slate-300 px-4 py-2.5 outline-none focus:border-blue-500 focus:ring-2 focus:ring-blue-100" />
        </label>
      </div>
      <div className="flex flex-col gap-3 rounded-xl border border-slate-200 bg-slate-50 p-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex gap-3"><ShieldCheck className="mt-0.5 h-5 w-5 shrink-0 text-emerald-600" /><div><p className="text-sm font-medium text-slate-800">Credencial cifrada no banco</p><p className="text-xs text-slate-500">O token nunca Ã© devolvido pela API. Para trocar, informe um novo valor e salve.</p></div></div>
        <label className="flex shrink-0 items-center gap-2 text-sm font-medium text-slate-700"><input type="checkbox" checked={ativa} disabled={!disponivelNoServidor} onChange={(e) => setAtiva(e.target.checked)} className="h-4 w-4 rounded border-slate-300 text-blue-600 disabled:opacity-50" />IntegraÃ§Ã£o ativa</label>
      </div>
      {!disponivelNoServidor && <p role="alert" className="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">O canal estÃ¡ desabilitado no servidor. Configure as variÃ¡veis META_* antes de ativÃ¡-lo.</p>}
      <p className="text-xs text-slate-500">No aplicativo Meta, configure a URL de callback <code>/api/webhooks/meta/leads</code>. O token de verificaÃ§Ã£o Ã© administrado somente no servidor.</p>
      {mensagem && <p role="status" className="text-sm text-emerald-700">{mensagem}</p>}
      {erro && <p role="alert" className="text-sm text-red-600">{erro}</p>}
      <div className="flex justify-end border-t border-slate-100 pt-5"><button disabled={salvando} className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-5 py-2.5 font-medium text-white disabled:opacity-60">{salvando ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}Salvar integraÃ§Ã£o</button></div>
    </form>
  </section>;
}
