import type { LoaderFunctionArgs } from 'react-router-dom';
import { sitePublicoService } from '../../services/sitePublicoService';

export const carregarSitePublico = async ({ params }: LoaderFunctionArgs) => {
  const slug = params.slug;
  if (!slug) throw new Response('Site não encontrado', { status: 404 });
  const [site, catalogo] = await Promise.all([
    sitePublicoService.obter(slug),
    sitePublicoService.listarImoveis(slug, { pagina: 1, limite: 12 }),
  ]);
  return { site, catalogo, basePublica: `/s/${site.slug}` };
};

export const carregarImovelPublico = async ({ params }: LoaderFunctionArgs) => {
  const { slug, slugImovel } = params;
  if (!slug || !slugImovel) throw new Response('Imóvel não encontrado', { status: 404 });
  const [site, imovel] = await Promise.all([
    sitePublicoService.obter(slug),
    sitePublicoService.obterImovel(slug, slugImovel),
  ]);
  return { site, imovel, basePublica: `/s/${site.slug}` };
};

export const carregarPoliticaPrivacidade = async ({ params }: LoaderFunctionArgs) => {
  if (!params.slug) throw new Response('Site não encontrado', { status: 404 });
  const site = await sitePublicoService.obter(params.slug);
  return { site, basePublica: `/s/${site.slug}` };
};
