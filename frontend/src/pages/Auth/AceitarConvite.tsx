import { useEffect, useState } from 'react';
import { CheckCircle2, Loader2 } from 'lucide-react';
import { Link } from 'react-router-dom';
import { obterMensagemErro } from '../../lib/http/erro-api';
import { equipeService } from '../../services/equipeService';

export const AceitarConvite = () => {
  const [token] = useState(() => new URLSearchParams(window.location.search).get('token') ?? '');
  const [nome, setNome] = useState('');
  const [senha, setSenha] = useState('');
  const [confirmacao, setConfirmacao] = useState('');
  const [enviando, setEnviando] = useState(false);
  const [concluido, setConcluido] = useState(false);
  const [erro, setErro] = useState('');

  useEffect(() => {
    window.history.replaceState({}, document.title, '/aceitar-convite');
  }, []);

  const enviar = async (evento: React.FormEvent) => {
    evento.preventDefault();
    if (!token) return setErro('O link do convite está incompleto.');
    if (senha !== confirmacao) return setErro('As senhas não coincidem.');
    setEnviando(true); setErro('');
    try {
      await equipeService.aceitarConvite(token, nome, senha);
      setConcluido(true);
    } catch (falha: unknown) {
      setErro(obterMensagemErro(falha, 'Não foi possível aceitar o convite.'));
    } finally {
      setEnviando(false);
    }
  };

  if (concluido) return <main className="grid min-h-screen place-items-center bg-slate-950 p-4"><div className="w-full max-w-md rounded-2xl bg-white p-8 text-center"><CheckCircle2 className="mx-auto h-12 w-12 text-emerald-600" /><h1 className="mt-4 text-2xl font-black">Conta criada</h1><p className="mt-2 text-slate-500">Seu acesso à equipe está pronto.</p><Link to="/login" className="mt-6 inline-flex rounded-xl bg-blue-600 px-5 py-3 font-bold text-white">Entrar no Kaptei</Link></div></main>;

  const inputClass = 'w-full rounded-xl border border-slate-200 px-4 py-3 outline-none focus:border-blue-600 focus:ring-2 focus:ring-blue-600/15';
  return <main className="grid min-h-screen place-items-center bg-slate-950 p-4"><form onSubmit={enviar} className="w-full max-w-md rounded-2xl bg-white p-8 shadow-2xl"><h1 className="text-2xl font-black text-slate-950">Aceitar convite</h1><p className="mb-6 mt-2 text-sm text-slate-500">Crie seu acesso pessoal à equipe da imobiliária.</p><div className="space-y-4"><label className="block text-sm font-semibold">Nome completo<input required minLength={2} maxLength={120} value={nome} onChange={(e) => setNome(e.target.value)} className={`${inputClass} mt-1`} /></label><label className="block text-sm font-semibold">Senha<input required type="password" minLength={6} maxLength={72} value={senha} onChange={(e) => setSenha(e.target.value)} className={`${inputClass} mt-1`} /></label><label className="block text-sm font-semibold">Confirme a senha<input required type="password" minLength={6} maxLength={72} value={confirmacao} onChange={(e) => setConfirmacao(e.target.value)} className={`${inputClass} mt-1`} /></label>{erro && <p role="alert" className="rounded-lg bg-red-50 p-3 text-sm text-red-700">{erro}</p>}<button disabled={enviando} className="inline-flex w-full items-center justify-center gap-2 rounded-xl bg-blue-600 px-5 py-3 font-bold text-white hover:bg-blue-700 disabled:opacity-60">{enviando && <Loader2 className="h-4 w-4 animate-spin" />}{enviando ? 'Criando acesso...' : 'Criar meu acesso'}</button></div></form></main>;
};
