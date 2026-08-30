import { Link, useLoaderData } from 'react-router-dom';
import { useMetadadosPagina } from '../../lib/seo/useMetadadosPagina';
import { rotaPublica } from '../../lib/sitePublico/rotas';
import type { SitePublico } from '../../types/sitePublico';
import { CabecalhoPublico } from './components/CabecalhoPublico';
import { FormularioDireitosTitular } from './components/FormularioDireitosTitular';

export const PoliticaPrivacidadePage = () => {
  const { site, basePublica } = useLoaderData() as { site: SitePublico; basePublica: string };
  useMetadadosPagina({ titulo: `Política de privacidade | ${site.nome}`, descricao: `Como ${site.nome} trata os dados enviados pelo site.` });

  return <div className="min-h-screen bg-slate-50">
    <CabecalhoPublico site={site} basePublica={basePublica} />
    <main className="mx-auto max-w-3xl px-4 py-14 sm:px-6">
      <p className="text-sm font-bold uppercase tracking-widest text-blue-600">Privacidade</p>
      <h1 className="mt-2 text-4xl font-black tracking-tight text-slate-950">Política de privacidade</h1>
      <p className="mt-4 text-sm text-slate-500">Controlador dos dados: {site.nome} · Versão de 12/08/2026</p>
      <div className="mt-10 space-y-8 leading-7 text-slate-600">
        <section><h2 className="text-xl font-bold text-slate-900">Dados coletados</h2><p className="mt-2">Coletamos os dados informados voluntariamente nos formulários, como nome, e-mail, telefone, mensagem, imóvel de interesse e parâmetros de campanha.</p></section>
        <section><h2 className="text-xl font-bold text-slate-900">Finalidade e base legal</h2><p className="mt-2">Os dados são utilizados para responder à solicitação e apresentar oportunidades relacionadas, com base no consentimento fornecido.</p></section>
        <section><h2 className="text-xl font-bold text-slate-900">Métricas essenciais do site</h2><p className="mt-2">Medimos visitas, visualizações, início e envio de formulário e cliques nos canais de atendimento usando um identificador pseudônimo da sessão e parâmetros de campanha. Essas métricas não armazenam IP, navegador, nome, e-mail, telefone ou mensagem e expiram em até 13 meses.</p></section>
        <section><h2 className="text-xl font-bold text-slate-900">Compartilhamento e segurança</h2><p className="mt-2">O acesso é limitado à equipe autorizada de {site.nome} e aos operadores técnicos necessários. Não comercializamos dados pessoais.</p></section>
        <section><h2 className="text-xl font-bold text-slate-900">Seus direitos</h2><p className="mt-2">Você pode solicitar confirmação, acesso, correção, portabilidade ou eliminação dos dados, além de revogar o consentimento.</p>{site.configuracao.email && <p className="mt-2">Canal de atendimento: <a className="font-semibold text-blue-700" href={`mailto:${site.configuracao.email}`}>{site.configuracao.email}</a>.</p>}</section>
        <section><h2 className="text-xl font-bold text-slate-900">Retenção</h2><p className="mt-2">Os dados de contato são mantidos somente pelo período necessário ao atendimento e ao cumprimento de obrigações legais e regulatórias. Eventos pseudônimos de conversão expiram em até 13 meses.</p></section>
      </div>
      <Link to={rotaPublica(basePublica)} className="mt-12 inline-flex font-bold text-blue-700">Voltar ao site</Link>
      <div className="mt-12">
        <FormularioDireitosTitular slug={site.slug} />
      </div>
    </main>
  </div>;
};
