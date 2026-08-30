import { useMemo, useState } from 'react';
import { Building2, Search, ShieldCheck, Sparkles } from 'lucide-react';
import { Link, useLoaderData } from 'react-router-dom';
import { useMetadadosPagina } from '../../lib/seo/useMetadadosPagina';
import { rotaPublica } from '../../lib/sitePublico/rotas';
import { useEventoConversao } from '../../lib/metricasConversao/useEventoConversao';
import { sitePublicoService } from '../../services/sitePublicoService';
import type { FiltrosCatalogo, ImovelPublico, SitePublico } from '../../types/sitePublico';
import { CabecalhoPublico } from './components/CabecalhoPublico';
import { CardImovelPublico } from './components/CardImovelPublico';
import { FormularioCaptacao } from './components/FormularioCaptacao';

interface DadosLoaderSite {
  site: SitePublico;
  catalogo: { dados: ImovelPublico[]; total: number; pagina: number; limite: number };
  basePublica: string;
}

export const SitePublicoPage = () => {
  const inicial = useLoaderData() as DadosLoaderSite;
  const [catalogo, setCatalogo] = useState(inicial.catalogo);
  const [filtros, setFiltros] = useState<FiltrosCatalogo>({ finalidade: '', tipo: '', cidade: '', quartos_min: undefined });
  const [buscando, setBuscando] = useState(false);
	const [erroCatalogo, setErroCatalogo] = useState('');
  const { site } = inicial;
  const { basePublica } = inicial;
  const titulo = site.configuracao.titulo || `Encontre seu próximo imóvel com ${site.nome}`;
  const descricao = site.configuracao.descricao || site.configuracao.subtitulo || `Imóveis selecionados por ${site.nome}.`;
  const dadosEstruturados = useMemo(() => ({
    '@context': 'https://schema.org', '@type': 'RealEstateAgent', name: site.nome,
    telephone: site.configuracao.telefone, email: site.configuracao.email,
  }), [site]);
  useMetadadosPagina({ titulo: `${titulo} | ${site.nome}`, descricao, imagem: site.configuracao.logo_url, dadosEstruturados });
  useEventoConversao(site.slug, 'SITE_VISUALIZADO');

  const pesquisar = async (evento: React.FormEvent) => {
    evento.preventDefault();
	setBuscando(true);
	setErroCatalogo('');
	try {
	  setCatalogo(await sitePublicoService.listarImoveis(site.slug, { ...filtros, pagina: 1, limite: 12 }));
	} catch {
	  setErroCatalogo('Não foi possível consultar os imóveis agora. Tente novamente.');
	} finally {
      setBuscando(false);
    }
  };

	const carregarMais = async () => {
	  setBuscando(true); setErroCatalogo('');
	  try {
		const proxima = await sitePublicoService.listarImoveis(site.slug, { ...filtros, pagina: catalogo.pagina + 1, limite: catalogo.limite });
		setCatalogo((atual) => ({ ...proxima, dados: [...atual.dados, ...proxima.dados] }));
	  } catch {
		setErroCatalogo('Não foi possível carregar mais imóveis. Tente novamente.');
	  } finally {
		setBuscando(false);
	  }
	};

  const inputClass = 'rounded-xl border border-white/15 bg-white/10 px-4 py-3 text-sm text-white outline-none placeholder:text-slate-400 focus:border-blue-400 focus:bg-white/15';
  return (
    <div className="min-h-screen bg-slate-50 text-slate-900">
      <CabecalhoPublico site={site} basePublica={basePublica} />
      <main>
        <section className="relative overflow-hidden bg-slate-950 px-4 py-20 text-white sm:px-6 lg:py-28">
          <div className="absolute inset-0 bg-[radial-gradient(circle_at_20%_20%,rgba(37,99,235,.35),transparent_38%),radial-gradient(circle_at_90%_10%,rgba(14,165,233,.16),transparent_30%)]" />
          <div className="relative mx-auto max-w-7xl">
            <div className="max-w-3xl">
              <span className="mb-5 inline-flex items-center gap-2 rounded-full border border-blue-400/20 bg-blue-500/10 px-4 py-2 text-xs font-bold uppercase tracking-widest text-blue-300"><Sparkles className="h-4 w-4" /> Curadoria imobiliária</span>
              <h1 className="text-4xl font-black tracking-tight sm:text-6xl">{titulo}</h1>
              <p className="mt-5 max-w-2xl text-lg leading-8 text-slate-300">{site.configuracao.subtitulo || descricao}</p>
            </div>
            <form onSubmit={pesquisar} className="mt-10 grid gap-3 rounded-2xl border border-white/10 bg-white/5 p-4 backdrop-blur md:grid-cols-5">
              <select aria-label="Finalidade" className={inputClass} value={filtros.finalidade} onChange={(e) => setFiltros((atual) => ({ ...atual, finalidade: e.target.value }))}><option className="text-slate-900" value="">Comprar ou alugar</option><option className="text-slate-900" value="Venda">Comprar</option><option className="text-slate-900" value="Locação">Alugar</option></select>
              <select aria-label="Tipo do imóvel" className={inputClass} value={filtros.tipo} onChange={(e) => setFiltros((atual) => ({ ...atual, tipo: e.target.value }))}><option className="text-slate-900" value="">Todos os tipos</option><option className="text-slate-900">Casa</option><option className="text-slate-900">Apartamento</option><option className="text-slate-900">Terreno</option><option className="text-slate-900">Comercial</option></select>
              <input aria-label="Cidade" maxLength={120} className={inputClass} placeholder="Cidade" value={filtros.cidade} onChange={(e) => setFiltros((atual) => ({ ...atual, cidade: e.target.value }))} />
              <select aria-label="Quantidade mínima de quartos" className={inputClass} value={filtros.quartos_min ?? ''} onChange={(e) => setFiltros((atual) => ({ ...atual, quartos_min: e.target.value ? Number(e.target.value) : undefined }))}><option className="text-slate-900" value="">Quartos</option><option className="text-slate-900" value="1">1+</option><option className="text-slate-900" value="2">2+</option><option className="text-slate-900" value="3">3+</option><option className="text-slate-900" value="4">4+</option></select>
              <button disabled={buscando} className="inline-flex items-center justify-center gap-2 rounded-xl bg-blue-600 px-5 py-3 font-bold hover:bg-blue-500 disabled:opacity-60"><Search className="h-4 w-4" /> {buscando ? 'Buscando...' : 'Buscar imóveis'}</button>
            </form>
          </div>
        </section>

        <section id="imoveis" className="mx-auto max-w-7xl px-4 py-16 sm:px-6 lg:px-8">
          <div className="mb-8 flex flex-wrap items-end justify-between gap-4"><div><p className="text-sm font-bold uppercase tracking-widest text-blue-600">Portfólio</p><h2 className="mt-1 text-3xl font-black tracking-tight">Imóveis selecionados</h2></div><span className="text-sm text-slate-500">{catalogo.total} {catalogo.total === 1 ? 'oportunidade' : 'oportunidades'}</span></div>
          {catalogo.dados.length ? <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">{catalogo.dados.map((imovel) => <CardImovelPublico key={imovel.id} imovel={imovel} basePublica={basePublica} />)}</div> : <div className="rounded-2xl border border-dashed border-slate-300 bg-white p-12 text-center text-slate-500">Nenhum imóvel corresponde aos filtros informados.</div>}
		  {erroCatalogo && <p role="alert" className="mt-6 text-center text-sm font-medium text-red-600">{erroCatalogo}</p>}
		  {catalogo.dados.length < catalogo.total && <div className="mt-10 text-center"><button type="button" onClick={carregarMais} disabled={buscando} className="rounded-xl border border-slate-300 bg-white px-6 py-3 font-bold text-slate-800 hover:border-blue-500 hover:text-blue-700 disabled:opacity-60">{buscando ? 'Carregando...' : 'Carregar mais imóveis'}</button></div>}
        </section>

        <section className="border-y border-slate-200 bg-white px-4 py-16 sm:px-6"><div className="mx-auto grid max-w-7xl gap-10 lg:grid-cols-[1fr_420px] lg:items-center"><div><Building2 className="h-10 w-10 text-blue-600" /><h2 className="mt-5 text-3xl font-black">Conte com quem conhece o mercado</h2><p className="mt-4 max-w-2xl leading-7 text-slate-600">{descricao}</p><div className="mt-7 flex flex-wrap gap-5 text-sm font-semibold text-slate-700"><span className="flex items-center gap-2"><ShieldCheck className="h-5 w-5 text-emerald-600" /> Atendimento responsável</span>{site.configuracao.creci && <span>CRECI {site.configuracao.creci}</span>}</div></div><FormularioCaptacao slugSite={site.slug} /></div></section>
      </main>
      <footer className="bg-slate-950 px-4 py-10 text-slate-400"><div className="mx-auto flex max-w-7xl flex-col justify-between gap-4 text-sm sm:flex-row"><span>© {new Date().getFullYear()} {site.nome}</span><div className="flex gap-5"><Link to={rotaPublica(basePublica, 'privacidade')} className="hover:text-white">Privacidade</Link><span>Plataforma de captação Kaptei</span></div></div></footer>
    </div>
  );
};
