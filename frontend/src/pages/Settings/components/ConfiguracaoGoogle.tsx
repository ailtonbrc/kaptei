import { useState, type FormEvent } from 'react';
import { Loader2, Save, User } from 'lucide-react';
import { api } from '../../../services/api';

export function ConfiguracaoGoogle({ inicial }: { inicial: string }) {
  const [clientID, setClientID] = useState(inicial);
  const [salvando, setSalvando] = useState(false);
  const [mensagem, setMensagem] = useState('');

  async function salvar(evento: FormEvent) {
    evento.preventDefault(); setSalvando(true); setMensagem('');
    try {
      await api.put('/v1/configuracoes/GOOGLE_CLIENT_ID', { valor: clientID, descricao: 'ID do cliente para autenticação via Google' });
      setMensagem('Google Client ID salvo com sucesso.');
    } catch { setMensagem('Não foi possível salvar o Google Client ID.'); }
    finally { setSalvando(false); }
  }

  return <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
    <header className="flex items-center gap-3 border-b border-slate-100 px-6 py-5"><span className="rounded-lg bg-blue-50 p-2 text-blue-600"><User className="h-5 w-5" /></span><div><h2 className="font-semibold text-slate-900">Autenticação social</h2><p className="text-sm text-slate-500">Credencial pública usada pelo login Google.</p></div></header>
    <form onSubmit={salvar} className="space-y-4 p-6"><label className="block text-sm font-medium text-slate-700">Google Client ID<input required maxLength={300} value={clientID} onChange={(e) => setClientID(e.target.value)} placeholder="123.apps.googleusercontent.com" className="mt-1 w-full rounded-lg border border-slate-300 px-4 py-2 focus:border-blue-600 focus:ring-2 focus:ring-blue-600/20" /></label>{mensagem && <p className="text-sm text-slate-600">{mensagem}</p>}<div className="flex justify-end"><button disabled={salvando} className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-5 py-2.5 font-medium text-white disabled:opacity-60">{salvando ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}Salvar</button></div></form>
  </section>;
}
