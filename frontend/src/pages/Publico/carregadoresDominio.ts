import { redirect, type LoaderFunctionArgs } from 'react-router-dom';
import { sitePublicoService } from '../../services/sitePublicoService';

const obterSiteDoDominio = async () => {
  const site = await sitePublicoService.obterPorDominio(window.location.hostname);
  if (!site) throw redirect('/app');
  return site;
};

export const carregarSiteDoDominio = async () => {
  const site = await obterSiteDoDominio();
  const catalogo = await sitePublicoService.listarImoveis(site.slug, { pagina: 1, limite: 12 });
  return { site, catalogo, basePublica: '/' };
};

export const carregarImovelDoDominio = async ({ params }: LoaderFunctionArgs) => {
  if (!params.slugImovel) throw new Response('Imóvel não encontrado', { status: 404 });
  const site = await obterSiteDoDominio();
  const imovel = await sitePublicoService.obterImovel(site.slug, params.slugImovel);
  return { site, imovel, basePublica: '/' };
};

export const carregarPrivacidadeDoDominio = async () => {
  const site = await obterSiteDoDominio();
  return { site, basePublica: '/' };
};
