import { useEffect, useState } from 'react';
import { CheckCircle2, Clock3, Loader2 } from 'lucide-react';
import { Link } from 'react-router-dom';
import { api } from '../../services/api';
import { useAuthStore } from '../../store/useAuthStore';

interface ContaAssinatura {
  status_plano: string;
  plano: string;
  trial_vence_em: string | null;
}

export const Checkout = () => {
  const [confirmado, setConfirmado] = useState(false);
  const [encerrado, setEncerrado] = useState(false);
  const atualizarConta = useAuthStore((estado) => estado.atualizarConta);

  useEffect(() => {
    let cancelado = false;
    let tentativas = 0;
    let temporizador: number | undefined;

    const consultar = async () => {
      try {
        const resposta = await api.get<ContaAssinatura>('/v1/conta');
        if (cancelado) return;
        atualizarConta(resposta.data);
        if (resposta.data.status_plano === 'ATIVO') {
          setConfirmado(true);
          return;
        }
      } catch {
        // O interceptor global trata uma eventual sessão expirada.
      }
      tentativas += 1;
      if (tentativas >= 10) {
        setEncerrado(true);
        return;
      }
      temporizador = window.setTimeout(consultar, 2000);
    };

    void consultar();
    return () => {
      cancelado = true;
      if (temporizador) window.clearTimeout(temporizador);
    };
  }, [atualizarConta]);

  return (
    <main className="grid min-h-screen place-items-center bg-slate-950 p-6 text-center text-white">
      <div className="max-w-lg rounded-3xl border border-white/10 bg-white/5 p-10 backdrop-blur">
        {confirmado ? <CheckCircle2 className="mx-auto h-16 w-16 text-emerald-400" /> : encerrado ? <Clock3 className="mx-auto h-16 w-16 text-amber-400" /> : <Loader2 className="mx-auto h-16 w-16 animate-spin text-blue-400" />}
        <h1 className="mt-6 text-3xl font-black">{confirmado ? 'Assinatura confirmada' : 'Confirmando assinatura'}</h1>
        <p className="mt-3 leading-7 text-slate-300">
          {confirmado ? 'Seu plano já está ativo e os recursos foram liberados.' : encerrado ? 'A confirmação ainda não chegou. Isso pode levar alguns instantes; consulte novamente pelo painel de assinatura.' : 'Aguardando a confirmação segura do provedor de pagamento.'}
        </p>
        <Link to={confirmado ? '/app' : '/app/assinatura'} className="mt-8 inline-flex rounded-xl bg-blue-600 px-6 py-3 font-bold hover:bg-blue-700">{confirmado ? 'Acessar o painel' : 'Ver assinatura'}</Link>
      </div>
    </main>
  );
};
