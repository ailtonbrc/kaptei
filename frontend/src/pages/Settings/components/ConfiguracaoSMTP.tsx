import { useState, type FormEvent } from 'react';
import { Eye, EyeOff, Loader2, Mail, Save } from 'lucide-react';
import { api } from '../../../services/api';
import type { ConfiguracaoSMTPDados } from '../hooks/useConfiguracoesSettings';

const campo = 'mt-1 w-full rounded-lg border border-slate-300 px-4 py-2 focus:border-blue-600 focus:ring-2 focus:ring-blue-600/20';

export function ConfiguracaoSMTP({ inicial }: { inicial: ConfiguracaoSMTPDados }) {
  const [dados, setDados] = useState(inicial);
  const [salvando, setSalvando] = useState(false);
  const [mostrarSenha, setMostrarSenha] = useState(false);
  const [mensagem, setMensagem] = useState('');
  const alterar = (chave: keyof ConfiguracaoSMTPDados, valor: string | number) => setDados((atual) => ({ ...atual, [chave]: valor }));

  async function salvar(evento: FormEvent) {
    evento.preventDefault(); setSalvando(true); setMensagem('');
    try {
      await api.put('/v1/configuracoes/SMTP_CONFIG', { valor: dados, descricao: 'Servidor SMTP para mensagens transacionais' });
      setDados((atual) => ({ ...atual, password: '' }));
      setMensagem('Configuração SMTP protegida e salva com sucesso.');
    } catch { setMensagem('Não foi possível salvar a configuração SMTP.'); }
    finally { setSalvando(false); }
  }

  return <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
    <header className="flex items-center gap-3 border-b border-slate-100 px-6 py-5"><span className="rounded-lg bg-blue-50 p-2 text-blue-600"><Mail className="h-5 w-5" /></span><div><h2 className="font-semibold text-slate-900">E-mail transacional (SMTP)</h2><p className="text-sm text-slate-500">Recuperação de senha, convites e notificações.</p></div></header>
    <form onSubmit={salvar} className="space-y-6 p-6"><div className="grid gap-5 md:grid-cols-2">
      <label className="text-sm font-medium text-slate-700">Host<input required maxLength={253} value={dados.host} onChange={(e) => alterar('host', e.target.value)} className={campo} /></label>
      <label className="text-sm font-medium text-slate-700">Porta<input required min={1} max={65535} type="number" value={dados.port} onChange={(e) => alterar('port', Number(e.target.value))} className={campo} /></label>
      <label className="text-sm font-medium text-slate-700">Usuário<input required maxLength={254} value={dados.user} onChange={(e) => alterar('user', e.target.value)} className={campo} /></label>
      <label className="text-sm font-medium text-slate-700">E-mail remetente<input required maxLength={254} type="email" value={dados.from_email} onChange={(e) => alterar('from_email', e.target.value)} className={campo} /></label>
      <label className="text-sm font-medium text-slate-700">Nome remetente<input required maxLength={120} value={dados.from_name} onChange={(e) => alterar('from_name', e.target.value)} className={campo} /></label>
      <label className="text-sm font-medium text-slate-700">Senha de aplicativo<div className="relative"><input maxLength={4096} type={mostrarSenha ? 'text' : 'password'} value={dados.password} onChange={(e) => alterar('password', e.target.value)} placeholder="Em branco mantém a atual" className={`${campo} pr-11`} /><button type="button" aria-label={mostrarSenha ? 'Ocultar senha' : 'Mostrar senha'} onClick={() => setMostrarSenha((valor) => !valor)} className="absolute inset-y-1 right-0 px-3 text-slate-400">{mostrarSenha ? <EyeOff className="h-5 w-5" /> : <Eye className="h-5 w-5" />}</button></div></label>
    </div>{mensagem && <p className="text-sm text-slate-600">{mensagem}</p>}<div className="flex justify-end border-t border-slate-100 pt-5"><button disabled={salvando} className="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-5 py-2.5 font-medium text-white disabled:opacity-60">{salvando ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}Salvar SMTP</button></div></form>
  </section>;
}
