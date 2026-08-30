import { useState } from 'react';
import { Bath, BedDouble, Car, ChevronLeft, MapPin, Maximize2 } from 'lucide-react';
import { Link, useLoaderData } from 'react-router-dom';
import { formatarMoedaBR } from '../../lib/formatadores/moeda';
import { listarValoresImovel } from '../../lib/imoveis/valores';
import { useMetadadosPagina } from '../../lib/seo/useMetadadosPagina';
import { rotaPublica } from '../../lib/sitePublico/rotas';
import { useEventoConversao } from '../../lib/metricasConversao/useEventoConversao';
import type { ImovelPublico, SitePublico } from '../../types/sitePublico';
import { CabecalhoPublico } from './components/CabecalhoPublico';
import { FormularioCaptacao } from './components/FormularioCaptacao';

interface DadosLoaderImovel { site: SitePublico; imovel: ImovelPublico; basePublica: string }

export const ImovelPublicoPage = () => {
  const { site, imovel, basePublica } = useLoaderData() as DadosLoaderImovel;
  const [fotoAtiva, setFotoAtiva] = useState(imovel.fotos.find((foto) => foto.is_capa)?.url ?? imovel.fotos[0]?.url);
  const valores = listarValoresImovel(imovel);
  const valorPrincipal = valores[0]?.valor;
  const descricaoSEO = imovel.descricao_seo || `${imovel.tipo} para ${imovel.finalidade.toLowerCase()} em ${[imovel.bairro, imovel.cidade].filter(Boolean).join(', ')}.`;
  const dadosEstruturados = {
    '@context': 'https://schema.org', '@type': 'RealEstateListing', name: imovel.titulo,
    description: descricaoSEO, image: imovel.fotos.map((foto) => foto.url),
    address: { '@type': 'PostalAddress', addressLocality: imovel.cidade, addressRegion: imovel.estado },
    offers: valorPrincipal ? { '@type': 'Offer', priceCurrency: 'BRL', price: valorPrincipal, availability: 'https://schema.org/InStock' } : undefined,
  };
  useMetadadosPagina({ titulo: `${imovel.titulo_seo || imovel.titulo} | ${site.nome}`, descricao: descricaoSEO, imagem: fotoAtiva, dadosEstruturados });
  useEventoConversao(site.slug, 'SITE_VISUALIZADO');
  useEventoConversao(site.slug, 'IMOVEL_VISUALIZADO', imovel.slug);

  const atributos = [
    { icone: BedDouble, rotulo: 'Quartos', valor: imovel.quartos },
    { icone: Bath, rotulo: 'Banheiros', valor: imovel.banheiros },
    { icone: Car, rotulo: 'Vagas', valor: imovel.vagas },
    { icone: Maximize2, rotulo: 'Área', valor: `${imovel.area_util ?? imovel.area_total ?? '--'} m²` },
  ];

  return (
    <div className="min-h-screen bg-slate-50">
      <CabecalhoPublico site={site} basePublica={basePublica} />
      <main className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
        <Link to={rotaPublica(basePublica)} className="mb-6 inline-flex items-center gap-2 text-sm font-semibold text-slate-600 hover:text-blue-700"><ChevronLeft className="h-4 w-4" /> Voltar aos imóveis</Link>
        <div className="grid gap-4 lg:grid-cols-[1fr_280px]">
          <div className="aspect-[16/10] overflow-hidden rounded-2xl bg-slate-200">{fotoAtiva ? <img src={fotoAtiva} alt={imovel.titulo} className="h-full w-full object-cover" /> : <div className="grid h-full place-items-center text-slate-500">Imagens em preparação</div>}</div>
          <div className="grid grid-cols-3 gap-3 lg:grid-cols-1 lg:overflow-y-auto">{imovel.fotos.slice(0, 4).map((foto, indice) => <button type="button" aria-label={`Visualizar foto ${indice + 1} de ${imovel.titulo}`} key={foto.id} onClick={() => setFotoAtiva(foto.url)} className={`aspect-[4/3] overflow-hidden rounded-xl border-2 ${fotoAtiva === foto.url ? 'border-blue-600' : 'border-transparent'}`}><img src={foto.url} alt="" loading="lazy" className="h-full w-full object-cover" /></button>)}</div>
        </div>

        <div className="mt-10 grid gap-10 lg:grid-cols-[1fr_390px]">
          <article>
            <div className="flex flex-wrap gap-2"><span className="rounded-full bg-blue-100 px-3 py-1 text-xs font-bold text-blue-700">{imovel.finalidade}</span><span className="rounded-full bg-slate-200 px-3 py-1 text-xs font-bold text-slate-700">{imovel.tipo}</span></div>
            <h1 className="mt-4 text-3xl font-black tracking-tight text-slate-950 sm:text-4xl">{imovel.titulo}</h1>
            <p className="mt-3 flex items-center gap-2 text-slate-500"><MapPin className="h-5 w-5" /> {[imovel.bairro, imovel.cidade, imovel.estado].filter(Boolean).join(', ')}</p>
            <div className="mt-6 space-y-1 text-blue-700">{valores.length ? valores.map((item) => <p key={item.rotulo} className="text-3xl font-black"><span className="text-sm uppercase text-slate-500">{item.rotulo}</span> R$ {formatarMoedaBR(item.valor)}</p>) : <p className="text-3xl font-black">Consulte o valor</p>}</div>
            <div className="my-8 grid grid-cols-2 gap-3 sm:grid-cols-4">{atributos.map(({ icone: Icone, rotulo, valor: valorAtributo }) => <div key={rotulo} className="rounded-xl border border-slate-200 bg-white p-4"><Icone className="h-5 w-5 text-blue-600" /><span className="mt-3 block text-xs text-slate-500">{rotulo}</span><strong className="text-slate-900">{valorAtributo}</strong></div>)}</div>
            {imovel.descricao && <section className="border-t border-slate-200 pt-8"><h2 className="text-xl font-extrabold text-slate-950">Sobre este imóvel</h2><p className="mt-4 whitespace-pre-line leading-8 text-slate-600">{imovel.descricao}</p></section>}
            {(imovel.valor_condominio || imovel.valor_iptu) && <section className="mt-8 border-t border-slate-200 pt-8"><h2 className="text-xl font-extrabold">Custos adicionais</h2><div className="mt-4 flex gap-8 text-sm text-slate-600">{imovel.valor_condominio && <span>Condomínio: <strong>R$ {formatarMoedaBR(imovel.valor_condominio)}</strong></span>}{imovel.valor_iptu && <span>IPTU: <strong>R$ {formatarMoedaBR(imovel.valor_iptu)}</strong></span>}</div></section>}
          </article>
          <aside className="lg:sticky lg:top-24 lg:self-start"><FormularioCaptacao slugSite={site.slug} slugImovel={imovel.slug} titulo="Tenho interesse neste imóvel" /></aside>
        </div>
      </main>
    </div>
  );
};
