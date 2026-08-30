import { useEffect, useState } from 'react';
import { CheckCircle2, CreditCard, Loader2, LockKeyhole } from 'lucide-react';
import { api } from '../../services/api';
import { useAuthStore } from '../../store/useAuthStore';
import { PlanoCard } from '../../components';
import type { Plano } from '../../constants/planos';
import { obterMensagemErro } from '../../lib/http/erro-api';

export const Billing = () => {
  const usuario = useAuthStore((estado) => estado.user);
  const [planos, setPlanos] = useState<Plano[]>([]);
  const [planoSelecionado, setPlanoSelecionado] = useState(usuario?.plano ?? '');
  const [carregando, setCarregando] = useState(true);
  const [redirecionando, setRedirecionando] = useState(false);
	const [abrindoPortal, setAbrindoPortal] = useState(false);
  const [erro, setErro] = useState('');

  useEffect(() => {
    void api.get<Plano[]>('/v1/planos')
      .then((resposta) => setPlanos(resposta.data))
      .catch((falha: unknown) => setErro(obterMensagemErro(falha, 'Não foi possível carregar os planos.')))
      .finally(() => setCarregando(false));
  }, []);

  const iniciarCheckout = async () => {
    if (!planoSelecionado) return;
    setRedirecionando(true); setErro('');
    try {
      const resposta = await api.post<{ checkout_url: string }>('/v1/billing/checkout',
        { plano: planoSelecionado },
        { headers: { 'Idempotency-Key': crypto.randomUUID() } },
      );
      window.location.assign(resposta.data.checkout_url);
    } catch (falha: unknown) {
      setErro(obterMensagemErro(falha, 'Não foi possível iniciar o pagamento.'));
      setRedirecionando(false);
    }
  };

	const abrirPortal = async () => {
	  setAbrindoPortal(true); setErro('');
	  try {
		const resposta = await api.post<{ portal_url: string }>('/v1/billing/portal');
		window.location.assign(resposta.data.portal_url);
	  } catch (falha: unknown) {
		setErro(obterMensagemErro(falha, 'Não foi possível abrir o gerenciamento da assinatura.'));
		setAbrindoPortal(false);
	  }
	};

  const tipo = usuario?.papel === 'GESTOR' ? 'IMOBILIARIA' : 'CORRETOR';
  const disponiveis = planos.filter((plano) => plano.tipo === tipo && !plano.codigo.includes('TRIAL'));
  const ativo = usuario?.status_plano === 'ATIVO';

  return (
    <main className="min-h-screen bg-slate-50 px-4 py-12 sm:px-6 lg:px-8">
      <div className="mx-auto max-w-6xl">
        <div className="mx-auto mb-10 max-w-2xl text-center">
          <span className="mx-auto grid h-14 w-14 place-items-center rounded-2xl bg-blue-100 text-blue-700"><CreditCard className="h-7 w-7" /></span>
          <h1 className="mt-5 text-3xl font-black tracking-tight text-slate-950">Assinatura Kaptei</h1>
          <p className="mt-3 text-slate-500">Escolha o plano ideal. O pagamento ocorre no ambiente seguro do Stripe; o Kaptei nunca recebe os dados do seu cartão.</p>
        </div>

		{ativo && <div className="mx-auto mb-8 flex max-w-2xl items-start gap-3 rounded-xl border border-emerald-200 bg-emerald-50 p-4 text-emerald-800"><CheckCircle2 className="mt-0.5 h-5 w-5 shrink-0" /><div className="flex-1"><strong>Assinatura ativa</strong><p className="text-sm">Seu plano atual é {usuario?.plano}. Alterações de plano, forma de pagamento e cancelamento são feitos no portal seguro.</p><button onClick={abrirPortal} disabled={abrindoPortal} className="mt-3 inline-flex items-center gap-2 rounded-lg bg-emerald-700 px-4 py-2 text-sm font-bold text-white disabled:opacity-60">{abrindoPortal && <Loader2 className="h-4 w-4 animate-spin" />}Gerenciar assinatura</button></div></div>}
        {erro && <div className="mx-auto mb-8 max-w-2xl rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">{erro}</div>}

        {carregando ? <Loader2 className="mx-auto h-8 w-8 animate-spin text-blue-600" /> : (
          <div className="grid gap-6 md:grid-cols-3">{disponiveis.map((plano) => <PlanoCard key={plano.id} plano={plano} selecionado={planoSelecionado === plano.codigo} onClick={() => setPlanoSelecionado(plano.codigo)} />)}</div>
        )}

        <div className="mx-auto mt-10 max-w-md text-center">
		  <button onClick={iniciarCheckout} disabled={ativo || !planoSelecionado || redirecionando} className="inline-flex w-full items-center justify-center gap-2 rounded-xl bg-blue-600 px-6 py-3.5 font-bold text-white hover:bg-blue-700 disabled:opacity-60">{redirecionando ? <Loader2 className="h-5 w-5 animate-spin" /> : <LockKeyhole className="h-5 w-5" />} {ativo ? 'Assinatura já ativa' : redirecionando ? 'Abrindo checkout seguro...' : 'Continuar para pagamento'}</button>
          <p className="mt-3 text-xs text-slate-400">A ativação é realizada automaticamente após a confirmação do pagamento.</p>
        </div>
      </div>
    </main>
  );
};
