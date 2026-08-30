import { useCallback, useEffect, useState } from 'react';
import { Loader2, MailPlus, RefreshCw, UserCheck, UserX, X } from 'lucide-react';
import { toast } from 'sonner';
import { obterMensagemErro } from '../../lib/http/erro-api';
import { equipeService, type ResumoEquipe } from '../../services/equipeService';

export const EquipePage = () => {
  const [dados, setDados] = useState<ResumoEquipe>({ membros: [], convites: [] });
  const [email, setEmail] = useState('');
  const [carregando, setCarregando] = useState(true);
  const [processando, setProcessando] = useState('');

  const carregar = useCallback(async () => {
    try {
      setDados(await equipeService.listar());
    } catch (erro: unknown) {
      toast.error(obterMensagemErro(erro, 'Não foi possível carregar a equipe.'));
    } finally {
      setCarregando(false);
    }
  }, []);

  useEffect(() => {
    let ativo = true;
    equipeService
      .listar()
      .then((resumo) => {
        if (ativo) setDados(resumo);
      })
      .catch((erro: unknown) => {
        if (ativo) toast.error(obterMensagemErro(erro, 'Não foi possível carregar a equipe.'));
      })
      .finally(() => {
        if (ativo) setCarregando(false);
      });
    return () => {
      ativo = false;
    };
  }, []);

  const convidar = async (evento: React.FormEvent) => {
    evento.preventDefault();
    setProcessando('convite');
    try {
      await equipeService.convidar(email);
      setEmail('');
      toast.success('Convite enviado por e-mail.');
      await carregar();
    } catch (erro: unknown) {
      toast.error(obterMensagemErro(erro, 'Não foi possível enviar o convite.'));
    } finally {
      setProcessando('');
    }
  };

  const alterarStatus = async (id: string, ativo: boolean) => {
    setProcessando(id);
    try {
      await equipeService.atualizarStatus(id, ativo ? 'ATIVO' : 'INATIVO');
      toast.success(ativo ? 'Corretor ativado.' : 'Corretor inativado e sessões revogadas.');
      await carregar();
    } catch (erro: unknown) {
      toast.error(obterMensagemErro(erro, 'Não foi possível atualizar o corretor.'));
    } finally {
      setProcessando('');
    }
  };

  const cancelar = async (id: string) => {
    setProcessando(id);
    try {
      await equipeService.cancelarConvite(id);
      toast.success('Convite cancelado.');
      await carregar();
    } catch (erro: unknown) {
      toast.error(obterMensagemErro(erro, 'Não foi possível cancelar o convite.'));
    } finally {
      setProcessando('');
    }
  };

  return (
    <div className="flex-1 overflow-y-auto p-4 sm:p-6 lg:p-8">
      <div className="mx-auto max-w-6xl space-y-6">
        <header className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h1 className="text-2xl font-black text-slate-950">Equipe da imobiliária</h1>
            <p className="mt-1 text-sm text-slate-500">Convide corretores e controle quem participa da distribuição de leads.</p>
          </div>
          <button type="button" onClick={() => { setCarregando(true); void carregar(); }} disabled={carregando} className="inline-flex items-center gap-2 rounded-xl border border-slate-200 bg-white px-4 py-2 text-sm font-bold text-slate-700 hover:border-blue-400">
            <RefreshCw className={`h-4 w-4 ${carregando ? 'animate-spin' : ''}`} /> Atualizar
          </button>
        </header>

        <form onSubmit={convidar} className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
          <h2 className="flex items-center gap-2 font-extrabold text-slate-900"><MailPlus className="h-5 w-5 text-blue-600" /> Convidar corretor</h2>
          <div className="mt-4 flex flex-col gap-3 sm:flex-row">
            <input required type="email" maxLength={254} value={email} onChange={(evento) => setEmail(evento.target.value)} placeholder="corretor@imobiliaria.com.br" className="min-w-0 flex-1 rounded-xl border border-slate-200 px-4 py-2.5 outline-none focus:border-blue-600 focus:ring-2 focus:ring-blue-600/15" />
            <button disabled={processando === 'convite'} className="inline-flex items-center justify-center gap-2 rounded-xl bg-blue-600 px-5 py-2.5 font-bold text-white hover:bg-blue-700 disabled:opacity-60">
              {processando === 'convite' && <Loader2 className="h-4 w-4 animate-spin" />} Enviar convite
            </button>
          </div>
          <p className="mt-2 text-xs text-slate-500">O convite expira em 72 horas e ocupa uma vaga do plano enquanto estiver pendente.</p>
        </form>

        <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
          <div className="border-b border-slate-100 px-5 py-4"><h2 className="font-extrabold text-slate-900">Membros</h2></div>
          {carregando ? <Loader2 className="mx-auto my-10 h-6 w-6 animate-spin text-blue-600" /> : (
            <div className="divide-y divide-slate-100">
              {dados.membros.map((membro) => (
                <div key={membro.id} className="flex flex-wrap items-center justify-between gap-4 px-5 py-4">
                  <div><strong className="block text-sm text-slate-900">{membro.nome}</strong><span className="text-sm text-slate-500">{membro.email} · {membro.papel === 'GESTOR' ? 'Gestor' : 'Corretor'}</span></div>
                  {membro.papel === 'CORRETOR_EQUIPE' && <button type="button" disabled={processando === membro.id} onClick={() => void alterarStatus(membro.id, membro.status.toUpperCase() !== 'ATIVO')} className={`inline-flex items-center gap-2 rounded-lg px-3 py-2 text-xs font-bold ${membro.status.toUpperCase() === 'ATIVO' ? 'bg-red-50 text-red-700' : 'bg-emerald-50 text-emerald-700'}`}>
                    {membro.status.toUpperCase() === 'ATIVO' ? <UserX className="h-4 w-4" /> : <UserCheck className="h-4 w-4" />}{membro.status.toUpperCase() === 'ATIVO' ? 'Inativar' : 'Ativar'}
                  </button>}
                </div>
              ))}
            </div>
          )}
        </section>

        {dados.convites.length > 0 && <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
          <div className="border-b border-slate-100 px-5 py-4"><h2 className="font-extrabold text-slate-900">Convites pendentes</h2></div>
          <div className="divide-y divide-slate-100">{dados.convites.map((convite) => <div key={convite.id} className="flex items-center justify-between gap-4 px-5 py-4"><div><strong className="block text-sm">{convite.email}</strong><span className="text-xs text-slate-500">Expira em {new Date(convite.expira_em).toLocaleString('pt-BR')}</span></div><button type="button" onClick={() => void cancelar(convite.id)} disabled={processando === convite.id} aria-label={`Cancelar convite de ${convite.email}`} className="rounded-lg p-2 text-red-600 hover:bg-red-50"><X className="h-4 w-4" /></button></div>)}</div>
        </section>}
      </div>
    </div>
  );
};
