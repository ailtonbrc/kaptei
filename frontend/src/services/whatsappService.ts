import { api } from './api';
import type {
  EnvioTemplateWhatsApp,
  ListaConversasWhatsApp,
  ListaMensagensWhatsApp,
} from '../types/whatsapp';

const base = '/v1/integracoes/whatsapp/conversas';

export const whatsappService = {
  async listarConversas(pagina = 1, busca = ''): Promise<ListaConversasWhatsApp> {
    const { data } = await api.get<ListaConversasWhatsApp>(base, { params: { pagina, limite: 50, busca } });
    return data;
  },

  async listarMensagens(conversaID: string, pagina = 1): Promise<ListaMensagensWhatsApp> {
    const { data } = await api.get<ListaMensagensWhatsApp>(`${base}/${conversaID}/mensagens`, {
      params: { pagina, limite: 100 },
    });
    return data;
  },

  async enviarTexto(conversaID: string, texto: string): Promise<void> {
    await api.post(`${base}/${conversaID}/mensagens`, { texto });
  },

  async enviarTemplate(conversaID: string, template: EnvioTemplateWhatsApp): Promise<void> {
    await api.post(`${base}/${conversaID}/templates`, template);
  },

  async registrarConsentimento(conversaID: string, consentiu: boolean, origem: string, evidencia: string): Promise<void> {
    await api.put(`${base}/${conversaID}/consentimento`, { consentiu, origem, evidencia });
  },
};
