import { useRef, useState } from 'react';
import { CheckCircle2, Loader2, Send } from 'lucide-react';
import { Link } from 'react-router-dom';
import { obterMensagemErro } from '../../../lib/http/erro-api';
import { sitePublicoService } from '../../../services/sitePublicoService';
import { obterAtribuicaoConversao, registrarEventoConversao } from '../../../lib/metricasConversao/metricasConversao';

interface FormularioCaptacaoProps {
  slugSite: string;
  slugImovel?: string;
  titulo?: string;
}

export const FormularioCaptacao = ({ slugSite, slugImovel, titulo = 'Receba atendimento personalizado' }: FormularioCaptacaoProps) => {
  const [enviando, setEnviando] = useState(false);
  const [enviado, setEnviado] = useState(false);
  const [erro, setErro] = useState('');
  const [dados, setDados] = useState({ nome: '', email: '', telefone: '', mensagem: '', consentimento: false });
  const [website, setWebsite] = useState('');
  const chaveIdempotencia = useRef(crypto.randomUUID());
  const formularioIniciado = useRef(false);

  const registrarInicio = () => {
    if (formularioIniciado.current) return;
    formularioIniciado.current = true;
    void registrarEventoConversao(slugSite, 'FORMULARIO_INICIADO', slugImovel);
  };

  const enviar = async (evento: React.FormEvent) => {

    evento.preventDefault();
    if (!dados.email.trim() && !dados.telefone.trim()) {
      setErro('Informe um e-mail ou telefone para receber o contato.');
      return;
    }
    setEnviando(true);
    setErro('');
    try {
      const atribuicao = obterAtribuicaoConversao(slugSite);
      await sitePublicoService.captarLead(slugSite, {
        nome: dados.nome,
        email: dados.email || undefined,
        telefone: dados.telefone || undefined,
        mensagem: dados.mensagem || undefined,
        imovel_slug: slugImovel,
        pagina_origem: window.location.pathname,
        utm_source: atribuicao.utm_source,
        utm_medium: atribuicao.utm_medium,
        utm_campaign: atribuicao.utm_campaign,
        consentimento_lgpd: dados.consentimento,
        chave_idempotencia: chaveIdempotencia.current,
        website,
      });
      void registrarEventoConversao(slugSite, 'LEAD_ENVIADO', slugImovel);
      setEnviado(true);
    } catch (falha: unknown) {
      setErro(obterMensagemErro(falha, 'Não foi possível enviar agora. Tente novamente.'));
    } finally {
      setEnviando(false);
    }
  };

  if (enviado) {
    return (
      <div className="rounded-2xl border border-emerald-200 bg-emerald-50 p-6 text-center text-emerald-800">
        <CheckCircle2 className="mx-auto mb-3 h-10 w-10" />
        <h3 className="font-bold">Contato recebido!</h3>
        <p className="mt-1 text-sm">Um especialista retornará pelos dados informados.</p>
      </div>
    );
  }

  const inputClass = 'w-full rounded-xl border border-slate-200 bg-white px-4 py-3 text-sm outline-none transition focus:border-blue-600 focus:ring-2 focus:ring-blue-600/15';
  return (
    <form onSubmit={enviar} onFocusCapture={registrarInicio} className="rounded-2xl border border-slate-200 bg-white p-6 shadow-xl shadow-slate-900/5">
      <h2 className="text-xl font-extrabold text-slate-950">{titulo}</h2>
      <p className="mb-5 mt-1 text-sm text-slate-500">Preencha seus dados e entraremos em contato.</p>
      <div className="space-y-3">
        <label className="absolute -left-[10000px]" aria-hidden="true">Website<input name="website" tabIndex={-1} autoComplete="off" value={website} onChange={(e) => setWebsite(e.target.value)} /></label>
        <input aria-label="Nome completo" autoComplete="name" required minLength={2} maxLength={120} className={inputClass} placeholder="Seu nome" value={dados.nome} onChange={(e) => setDados((atual) => ({ ...atual, nome: e.target.value }))} />
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-1">
          <input aria-label="E-mail" autoComplete="email" maxLength={254} type="email" className={inputClass} placeholder="E-mail" value={dados.email} onChange={(e) => setDados((atual) => ({ ...atual, email: e.target.value }))} />
          <input aria-label="Telefone ou WhatsApp" autoComplete="tel" maxLength={30} className={inputClass} placeholder="Telefone ou WhatsApp" value={dados.telefone} onChange={(e) => setDados((atual) => ({ ...atual, telefone: e.target.value }))} />
        </div>
        <textarea aria-label="Mensagem" maxLength={2000} rows={4} className={inputClass} placeholder="Como podemos ajudar?" value={dados.mensagem} onChange={(e) => setDados((atual) => ({ ...atual, mensagem: e.target.value }))} />
        <label className="flex items-start gap-2 text-xs leading-5 text-slate-500">
          <input required type="checkbox" checked={dados.consentimento} onChange={(e) => setDados((atual) => ({ ...atual, consentimento: e.target.checked }))} className="mt-1" />
          <span>Autorizo o contato sobre este imóvel e oportunidades relacionadas, conforme a <Link to={`/s/${slugSite}/privacidade`} target="_blank" className="font-semibold text-blue-700 underline">política de privacidade</Link>.</span>
        </label>
        {erro && <p className="rounded-lg bg-red-50 p-3 text-sm text-red-700">{erro}</p>}
        <button disabled={enviando} className="inline-flex w-full items-center justify-center gap-2 rounded-xl bg-blue-600 px-5 py-3 font-bold text-white hover:bg-blue-700 disabled:opacity-60">
          {enviando ? <Loader2 className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
          {enviando ? 'Enviando...' : 'Quero receber contato'}
        </button>
      </div>
    </form>
  );
};
