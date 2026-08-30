import { useState, type FormEvent } from 'react';
import { Loader2, MessageCircle, Save, ShieldCheck } from 'lucide-react';
import { api } from '../../../services/api';
import type { ConfiguracaoWhatsAppDados } from '../hooks/useConfiguracoesSettings';

export function ConfiguracaoWhatsApp({ inicial }: { inicial: ConfiguracaoWhatsAppDados }) {
  const [wabaID, setWabaID] = useState(inicial.waba_id ?? '');
  const [numeroID, setNumeroID] = useState(inicial.numero_telefone_id ?? '');
  const [numeroExibicao, setNumeroExibicao] = useState(inicial.numero_exibicao ?? '');
  const [token, setToken] = useState('');
  const [ativa, setAtiva] = useState(inicial.ativa ?? false);
  const [tokenConfigurado, setTokenConfigurado] = useState(inicial.token_acesso_configurado ?? false);
  const [salvando, setSalvando] = useState(false);
  const [mensagem, setMensagem] = useState('');
  const [erro, setErro] = useState('');

  async function salvar(evento: FormEvent) {
    evento.preventDefault(); setSalvando(true); setMensagem(''); setErro('');
    try {
      const resposta = await api.put<ConfiguracaoWhatsAppDados>('/v1/integracoes/whatsapp', {
        waba_id: wabaID.trim(), numero_telefone_id: numeroID.trim(),
        numero_exibicao: numeroExibicao.trim() || undefined,
        token_acesso: token.trim() || undefined, ativa,
      });
      setToken(''); setTokenConfigurado(resposta.data.token_acesso_configurado);
      setMensagem('Configuração do WhatsApp salva com segurança.');
    } catch {
      setErro('Não foi possível salvar. Confira os identificadores, o token e os vínculos da conta Meta.');
    } finally { setSalvando(false); }
  }

  return <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
    <header className="flex items-center gap-3 border-b border-slate-100 px-6 py-5"><span className="rounded-lg bg-emerald-50 p-2 text-emerald-600"><MessageCircle className="h-5 w-5" /></span><div><h2 className="font-semibold text-slate-900">WhatsApp Cloud API</h2><p className="text-sm text-slate-500">Base oficial para captar conversas e atender leads pelo CRM.</p></div></header>
    <form onSubmit={salvar} className="space-y-5 p-6">
      <div className="grid gap-5 md:grid-cols-2">
        <label className="text-sm font-medium text-slate-700">WABA ID<input required inputMode="numeric" maxLength={64} value={wabaID} onChange={(e) => setWabaID(e.target.value.replace(/\D/g, ''))} className="mt-1 w-full rounded-lg border border-slate-300 px-4 py-2.5" /></label>
        <label className="text-sm font-medium text-slate-700">Phone Number ID<input required inputMode="numeric" maxLength={64} value={numeroID} onChange={(e) => setNumeroID(e.target.value.replace(/\D/g, ''))} className="mt-1 w-full rounded-lg border border-slate-300 px-4 py-2.5" /></label>
        <label className="text-sm font-medium text-slate-700">Número exibido<input autoComplete="tel" value={numeroExibicao} onChange={(e) => setNumeroExibicao(e.target.value)} placeholder="+5565999999999" className="mt-1 w-full rounded-lg border border-slate-300 px-4 py-2.5" /></label>
        <label className="text-sm font-medium text-slate-700">Token de usuário do sistema<input type="password" autoComplete="new-password" value={token} onChange={(e) => setToken(e.target.value)} placeholder={tokenConfigurado ? 'Deixe vazio para manter o token protegido' : 'Cole o token com permissão de mensageria'} className="mt-1 w-full rounded-lg border border-slate-300 px-4 py-2.5" /></label>
      </div>
      <div className="flex flex-col gap-3 rounded-xl border border-slate-200 bg-slate-50 p-4 sm:flex-row sm:items-center sm:justify-between"><div className="flex gap-3"><ShieldCheck className="mt-0.5 h-5 w-5 text-emerald-600" /><div><p className="text-sm font-medium text-slate-800">Token cifrado e isolado por imobiliária</p><p className="text-xs text-slate-500">O backend não devolve o token persistido.</p></div></div><label className="flex items-center gap-2 text-sm font-medium"><input type="checkbox" checked={ativa} disabled={!inicial.disponivel_no_servidor} onChange={(e) => setAtiva(e.target.checked)} />Integração ativa</label></div>
      {!inicial.disponivel_no_servidor && <p role="alert" className="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">Configure o aplicativo Meta no servidor antes de ativar este canal.</p>}
      {mensagem && <p role="status" className="text-sm text-emerald-700">{mensagem}</p>}{erro && <p role="alert" className="text-sm text-red-600">{erro}</p>}
      <div className="flex justify-end border-t border-slate-100 pt-5"><button disabled={salvando} className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-5 py-2.5 font-medium text-white disabled:opacity-60">{salvando ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}Salvar WhatsApp</button></div>
    </form>
  </section>;
}
