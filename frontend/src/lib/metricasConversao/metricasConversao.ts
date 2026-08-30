import { sitePublicoService } from '../../services/sitePublicoService';
import type { AtribuicaoConversao, TipoEventoConversao } from '../../types/sitePublico';

const prefixoSessao = 'kaptei:sessao-conversao:';
const prefixoAtribuicao = 'kaptei:atribuicao-conversao:';
const sessoesVolateis = new Map<string, string>();
const atribuicoesVolateis = new Map<string, AtribuicaoConversao>();

function acessarSessao(): Storage | undefined {
  try {
    return window.sessionStorage;
  } catch {
    return undefined;
  }
}

function obterSessaoID(slugSite: string): string {
  const chaveSite = slugSite.toLowerCase();
  const chaveSessao = `${prefixoSessao}${chaveSite}`;
  const armazenamento = acessarSessao();
  const existente = armazenamento?.getItem(chaveSessao);
  if (existente) return existente;
  const sessaoVolatil = sessoesVolateis.get(chaveSite) ?? crypto.randomUUID();
  sessoesVolateis.set(chaveSite, sessaoVolatil);
  try {
    armazenamento?.setItem(chaveSessao, sessaoVolatil);
  } catch {
    // O identificador permanece apenas em memória quando o navegador bloqueia storage.
  }
  return sessaoVolatil;
}

function atribuicaoDaURL(): AtribuicaoConversao {
  const parametros = new URLSearchParams(window.location.search);
  return {
    utm_source: parametros.get('utm_source')?.slice(0, 120) || undefined,
    utm_medium: parametros.get('utm_medium')?.slice(0, 120) || undefined,
    utm_campaign: parametros.get('utm_campaign')?.slice(0, 180) || undefined,
  };
}

export function obterAtribuicaoConversao(slugSite: string): AtribuicaoConversao {
  const chaveSite = slugSite.toLowerCase();
  const chaveAtribuicao = `${prefixoAtribuicao}${chaveSite}`;
  const armazenamento = acessarSessao();
  const capturada = atribuicaoDaURL();
  const possuiCampanha = Boolean(capturada.utm_source || capturada.utm_medium || capturada.utm_campaign);
  if (possuiCampanha) {
    atribuicoesVolateis.set(chaveSite, capturada);
    try {
      armazenamento?.setItem(chaveAtribuicao, JSON.stringify(capturada));
    } catch {
      // A atribuição volátil ainda acompanha a navegação atual.
    }
    return capturada;
  }

  try {
    const persistida = armazenamento?.getItem(chaveAtribuicao);
    if (persistida) return JSON.parse(persistida) as AtribuicaoConversao;
  } catch {
    // Dados inválidos ou storage bloqueado voltam para a atribuição em memória.
  }
  return atribuicoesVolateis.get(chaveSite) ?? {};
}

export async function registrarEventoConversao(slugSite: string, tipo: TipoEventoConversao, slugImovel?: string): Promise<void> {
  try {
    await sitePublicoService.registrarEventoConversao(slugSite, {
      chave_evento: crypto.randomUUID(),
      sessao_id: obterSessaoID(slugSite),
      tipo,
      imovel_slug: slugImovel,
      ...obterAtribuicaoConversao(slugSite),
    });
  } catch {
    // As métricas nunca bloqueiam navegação, contato ou captação do lead.
  }
}
