import { api } from './api';
import type {
  ConfiguracaoPortal,
  CredencialFeedPortal,
  DiagnosticoFeedPortal,
  PublicacaoPortal,
} from '../types/portalImobiliario';

const base = '/v1/integracoes/portais/grupo-olx';

export const portalImobiliarioService = {
  async obterConfiguracao(): Promise<ConfiguracaoPortal> {
    const resposta = await api.get<ConfiguracaoPortal>(base);
    return resposta.data;
  },
  async salvarConfiguracao(configuracao: ConfiguracaoPortal): Promise<ConfiguracaoPortal> {
    const resposta = await api.put<ConfiguracaoPortal>(base, configuracao);
    return resposta.data;
  },
  async rotacionarToken(): Promise<CredencialFeedPortal> {
    const resposta = await api.post<CredencialFeedPortal>(`${base}/token`);
    return resposta.data;
  },
  async listarPublicacoes(): Promise<PublicacaoPortal[]> {
    const resposta = await api.get<PublicacaoPortal[]>(`${base}/publicacoes`);
    return resposta.data;
  },
  async salvarPublicacao(imovelId: string, ativa: boolean, tipoPublicacao: PublicacaoPortal['tipo_publicacao']): Promise<void> {
    await api.put(`${base}/publicacoes/${imovelId}`, { ativa, tipo_publicacao: tipoPublicacao });
  },
  async diagnosticar(): Promise<DiagnosticoFeedPortal> {
    const resposta = await api.get<DiagnosticoFeedPortal>(`${base}/diagnostico`);
    return resposta.data;
  },
};
