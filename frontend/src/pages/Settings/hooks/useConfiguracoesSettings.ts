import { useEffect, useState } from 'react';
import { api } from '../../../services/api';
import { siteAdminService, type AtualizacaoSitePublico } from '../../../services/siteAdminService';

export interface ConfiguracaoSMTPDados {
  host: string;
  port: number;
  user: string;
  password: string;
  from_email: string;
  from_name: string;
}

export interface ConfiguracaoContaDados {
  lead_estrategia?: string;
  lead_token_integracao?: string;
  lead_token_prefixo?: string;
}

export interface ConfiguracaoMetaLeadsDados {
  pagina_id: string;
  token_pagina_configurado: boolean;
  disponivel_no_servidor: boolean;
  ativa: boolean;
  criado_em?: string;
  atualizado_em?: string;
}

export interface ConfiguracaoWhatsAppDados {
  waba_id: string;
  numero_telefone_id: string;
  numero_exibicao?: string;
  token_acesso_configurado: boolean;
  disponivel_no_servidor: boolean;
  ativa: boolean;
}

export interface ConfiguracaoObservabilidadeDados {
  ativa: boolean;
  token_configurado: boolean;
}

const smtpVazio: ConfiguracaoSMTPDados = { host: '', port: 587, user: '', password: '', from_email: '', from_name: '' };

export function useConfiguracoesSettings(superAdmin: boolean, podeAdministrarConta: boolean) {
  const [carregando, setCarregando] = useState(true);
  const [smtp, setSmtp] = useState<ConfiguracaoSMTPDados>(smtpVazio);
  const [googleClientID, setGoogleClientID] = useState('');
  const [conta, setConta] = useState<ConfiguracaoContaDados>({});
  const [site, setSite] = useState<AtualizacaoSitePublico | null>(null);
  const [metaLeads, setMetaLeads] = useState<ConfiguracaoMetaLeadsDados>({ pagina_id: '', token_pagina_configurado: false, disponivel_no_servidor: false, ativa: false });
  const [whatsApp, setWhatsApp] = useState<ConfiguracaoWhatsAppDados>({ waba_id: '', numero_telefone_id: '', token_acesso_configurado: false, disponivel_no_servidor: false, ativa: false });
  const [observabilidade, setObservabilidade] = useState<ConfiguracaoObservabilidadeDados>({ ativa: false, token_configurado: false });

  useEffect(() => {
    let ativo = true;
    async function carregar() {
      setCarregando(true);
      if (superAdmin) {
        const [smtpResultado, googleResultado, observabilidadeResultado] = await Promise.allSettled([
          api.get('/v1/configuracoes/SMTP_CONFIG'),
          api.get('/v1/configuracoes/GOOGLE_CLIENT_ID'),
          api.get('/v1/configuracoes/OBSERVABILIDADE_CONFIG'),
        ]);
        if (!ativo) return;
        if (smtpResultado.status === 'fulfilled' && smtpResultado.value.data?.valor) {
          setSmtp({ ...smtpVazio, ...smtpResultado.value.data.valor, password: '' });
        }
        if (googleResultado.status === 'fulfilled' && googleResultado.value.data?.valor) setGoogleClientID(googleResultado.value.data.valor);
        if (observabilidadeResultado.status === 'fulfilled' && observabilidadeResultado.value.data?.valor) setObservabilidade(observabilidadeResultado.value.data.valor);
      }
      const [contaResultado, siteResultado, metaResultado, whatsAppResultado] = await Promise.allSettled([
        api.get<ConfiguracaoContaDados>('/v1/conta'),
        podeAdministrarConta ? siteAdminService.obter() : Promise.resolve(null),
        podeAdministrarConta ? api.get<ConfiguracaoMetaLeadsDados>('/v1/integracoes/meta/leads') : Promise.resolve(null),
        podeAdministrarConta ? api.get<ConfiguracaoWhatsAppDados>('/v1/integracoes/whatsapp') : Promise.resolve(null),
      ]);
      if (!ativo) return;
      if (contaResultado.status === 'fulfilled') setConta(contaResultado.value.data);
      if (siteResultado.status === 'fulfilled' && siteResultado.value) {
        setSite({ slug: siteResultado.value.slug, publicado: siteResultado.value.publicado, configuracao: siteResultado.value.configuracao ?? {} });
      }
      if (metaResultado.status === 'fulfilled' && metaResultado.value) setMetaLeads(metaResultado.value.data);
      if (whatsAppResultado.status === 'fulfilled' && whatsAppResultado.value) setWhatsApp(whatsAppResultado.value.data);
      setCarregando(false);
    }
    void carregar();
    return () => { ativo = false; };
  }, [superAdmin, podeAdministrarConta]);

  return { carregando, smtp, googleClientID, conta, site, metaLeads, whatsApp, observabilidade };
}
