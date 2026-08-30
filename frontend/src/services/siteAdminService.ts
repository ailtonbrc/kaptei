import { api } from './api';
import type { ConfiguracaoSitePublico, SitePublico } from '../types/sitePublico';
import type { DominioSite } from '../types/dominioSite';

export interface AtualizacaoSitePublico {
  slug: string;
  publicado: boolean;
  configuracao: ConfiguracaoSitePublico;
}

export const siteAdminService = {
  async obter(): Promise<SitePublico> {
    const resposta = await api.get<SitePublico>('/v1/site');
    return resposta.data;
  },
  async salvar(dados: AtualizacaoSitePublico): Promise<void> {
    await api.put('/v1/site', dados);
  },
  async obterDominio(): Promise<DominioSite | undefined> {
    const resposta = await api.get<DominioSite | { configurado: false }>('/v1/site/dominio');
    return 'configurado' in resposta.data ? undefined : resposta.data;
  },
  async configurarDominio(hostname: string): Promise<DominioSite> {
    const resposta = await api.put<DominioSite>('/v1/site/dominio', { hostname });
    return resposta.data;
  },
  async verificarDominio(): Promise<DominioSite> {
    const resposta = await api.post<DominioSite>('/v1/site/dominio/verificacao');
    return resposta.data;
  },
};
